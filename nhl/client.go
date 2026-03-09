package nhl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	baseURL       = "https://api-web.nhle.com/v1"
	HughesID      = 8481559
	DevilsTeamAbb = "NJD"
)

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) get(url string, v any) error {
	resp, err := c.http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("NHL API returned %d for %s", resp.StatusCode, url)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// --- Scoreboard ---

type Scoreboard struct {
	Games []Game `json:"games"`
}

type Game struct {
	ID           int    `json:"id"`
	GameState    string `json:"gameState"` // "FUT", "LIVE", "FINAL", "OFF"
	HomeTeam     Team   `json:"homeTeam"`
	AwayTeam     Team   `json:"awayTeam"`
	StartTimeUTC string `json:"startTimeUTC"`
}

type Team struct {
	Abbrev string `json:"abbrev"`
	Score  int    `json:"score"`
}

// GetLastDevilsGame returns the most recent final Devils game, checking
// both today and yesterday in ET to catch late west coast games that
// finished after midnight ET.
func (c *Client) GetLastDevilsGame() (*Game, error) {
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(loc)

	for _, date := range []string{
		now.Format("2006-01-02"),
		now.AddDate(0, 0, -1).Format("2006-01-02"),
	} {
		var sb Scoreboard
		if err := c.get(fmt.Sprintf("%s/score/%s", baseURL, date), &sb); err != nil {
			return nil, fmt.Errorf("scoreboard for %s: %w", date, err)
		}
		for _, g := range sb.Games {
			if g.HomeTeam.Abbrev == DevilsTeamAbb || g.AwayTeam.Abbrev == DevilsTeamAbb {
				if g.IsFinal() {
					game := g
					return &game, nil
				}
			}
		}
	}
	return nil, nil // no recent final game found
}

func (g *Game) IsFinal() bool {
	return g.GameState == "FINAL" || g.GameState == "OFF"
}

func (g *Game) Opponent() string {
	if g.HomeTeam.Abbrev == DevilsTeamAbb {
		return g.AwayTeam.Abbrev
	}
	return g.HomeTeam.Abbrev
}

// --- Boxscore ---

type Boxscore struct {
	PlayerByGameStats PlayerByGameStats `json:"playerByGameStats"`
}

type PlayerByGameStats struct {
	HomeTeam BoxscoreTeam `json:"homeTeam"`
	AwayTeam BoxscoreTeam `json:"awayTeam"`
}

type BoxscoreTeam struct {
	Forwards  []PlayerGameStats `json:"forwards"`
	Defense   []PlayerGameStats `json:"defense"`
	Goalies   []PlayerGameStats `json:"goalies"`
}

type PlayerGameStats struct {
	PlayerID int `json:"playerId"`
	Assists  int `json:"assists"`
	Goals    int `json:"goals"`
	Points   int `json:"points"`
}

func (c *Client) GetHughesGamePoints(gameID int) (int, error) {
	var bs Boxscore
	if err := c.get(fmt.Sprintf("%s/gamecenter/%d/boxscore", baseURL, gameID), &bs); err != nil {
		return 0, err
	}

	// Hughes is always on NJD, search both home and away
	for _, team := range []BoxscoreTeam{bs.PlayerByGameStats.HomeTeam, bs.PlayerByGameStats.AwayTeam} {
		for _, p := range team.Forwards {
			if p.PlayerID == HughesID {
				return p.Points, nil
			}
		}
	}
	return 0, nil // played but no points (or not in lineup — e.g. injured)
}

// --- Player career stats ---

type PlayerLanding struct {
	CareerTotals CareerTotals `json:"careerTotals"`
}

type CareerTotals struct {
	RegularSeason SeasonStats `json:"regularSeason"`
}

type SeasonStats struct {
	GamesPlayed int `json:"gamesPlayed"`
	Goals       int `json:"goals"`
	Assists     int `json:"assists"`
	Points      int `json:"points"`
}

func (c *Client) GetHughesCareerStats() (*SeasonStats, error) {
	var landing PlayerLanding
	if err := c.get(fmt.Sprintf("%s/player/%d/landing", baseURL, HughesID), &landing); err != nil {
		return nil, err
	}
	s := landing.CareerTotals.RegularSeason
	return &s, nil
}

// --- Next game ---

type ScheduleWeek struct {
	Games []ScheduleGame `json:"games"`
}

type ScheduleGame struct {
	ID           int    `json:"id"`
	GameDate     string `json:"gameDate"` // "2025-03-10"
	StartTimeUTC string `json:"startTimeUTC"`
	HomeTeam     ScheduleTeam `json:"homeTeam"`
	AwayTeam     ScheduleTeam `json:"awayTeam"`
	GameState    string `json:"gameState"`
}

type ScheduleTeam struct {
	Abbrev   string `json:"abbrev"`
	FullName string `json:"placeName"`
}

func (c *Client) GetNextDevilsGame() (*ScheduleGame, error) {
	loc, _ := time.LoadLocation("America/New_York")
	today := time.Now().In(loc).Format("2006-01-02")
	nextWeek := time.Now().In(loc).AddDate(0, 0, 7).Format("2006-01-02")

	// Try this week and next week
	for _, endpoint := range []string{today, nextWeek} {
		var sw ScheduleWeek
		url := fmt.Sprintf("%s/club-schedule/%s/week/%s", baseURL, DevilsTeamAbb, endpoint)
		if err := c.get(url, &sw); err != nil {
			fmt.Printf("URL: %s, Games count: %d\n", url, len(sw.Games))
			continue
		}
		for _, g := range sw.Games {
			fmt.Printf("Found game: %s vs %s on %s\n", g.HomeTeam.Abbrev, g.AwayTeam.Abbrev, g.GameDate)
			if g.GameDate > today {
				game := g
				return &game, nil
			}
		}
	}
	return nil, fmt.Errorf("no upcoming game found")
}

func (g *ScheduleGame) OpponentName() string {
	if g.HomeTeam.Abbrev == DevilsTeamAbb {
		return g.AwayTeam.Abbrev
	}
	return g.HomeTeam.Abbrev
}

func (g *ScheduleGame) FormattedDate() string {
	t, err := time.Parse("2006-01-02", g.GameDate)
	if err != nil {
		return g.GameDate
	}
	return t.Format("Mon Jan 2")
}
