package main

import (
	_ "time/tzdata"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/simplycycling/jack-ppg-tracker/bluesky"
	"github.com/simplycycling/jack-ppg-tracker/nhl"
)

func main() {
	bskyHandle := os.Getenv("BSKY_HANDLE")
	bskyPassword := os.Getenv("BSKY_APP_PASSWORD")

	if bskyHandle == "" || bskyPassword == "" {
		log.Fatal("BSKY_HANDLE and BSKY_APP_PASSWORD environment variables are required")
	}

	nhlClient := nhl.NewClient()

	// 1. Check if Devils played today
	game, err := nhlClient.GetLastDevilsGame()
	if err != nil {
		log.Fatalf("Failed to get today's game: %v", err)
	}
	if game == nil {
		fmt.Println("No Devils game today. Nothing to post.")
		return
	}
	if !game.IsFinal() {
		fmt.Printf("Devils game today (vs %s) is not final yet (state: %s). Nothing to post.\n",
			game.Opponent(), game.GameState)
		return
	}

	fmt.Printf("Devils game vs %s is final.\n", game.Opponent())

	// 2. Get Hughes' points in tonight's game
	gamePoints, err := nhlClient.GetHughesGamePoints(game.ID)
	if err != nil {
		log.Fatalf("Failed to get game stats: %v", err)
	}

	// 3. Get Hughes' career totals
	career, err := nhlClient.GetHughesCareerStats()
	if err != nil {
		log.Fatalf("Failed to get career stats: %v", err)
	}

	// 4. Get next Devils game
	nextGame, err := nhlClient.GetNextDevilsGame()
		if err != nil {
    	fmt.Printf("Warning: could not get next game: %v\n", err)
		} else if nextGame == nil {
    	fmt.Println("Warning: nextGame is nil, no upcoming game found")
		} else {
    	fmt.Printf("Next game: %s on %s\n", nextGame.OpponentName(), nextGame.FormattedDate())
	}

	// 5. Compose post
	post := buildPost(game, gamePoints, career, nextGame)
	fmt.Println("Composed post:")
	fmt.Println(post)

	// 6. Post to Bluesky
	bskyClient := bluesky.NewClient(bskyHandle, bskyPassword)
	if err := bskyClient.Login(); err != nil {
		log.Fatalf("Bluesky login failed: %v", err)
	}
	if err := bskyClient.PostText(post); err != nil {
		log.Fatalf("Failed to post to Bluesky: %v", err)
	}
	fmt.Println("✅ Posted to Bluesky successfully.")
}

func buildPost(game *nhl.Game, gamePoints int, career *nhl.SeasonStats, nextGame *nhl.ScheduleGame) string {
	gp := career.GamesPlayed
	pts := career.Points
	ppgGap := gp - pts // positive = points he still needs

	var sb strings.Builder

	// Header
	sb.WriteString("🏒 Jack Hughes PPG Tracker\n\n")

	// Tonight's game
	if gamePoints == 0 {
		sb.WriteString(fmt.Sprintf("Last game vs %s: no points 😬\n\n", game.Opponent()))
	} else {
		sb.WriteString(fmt.Sprintf("Last game vs %s: %s 🔥\n\n",
			game.Opponent(), pointsStr(gamePoints)))
	}

	// Career totals
	sb.WriteString(fmt.Sprintf("Career: %d pts in %d GP\n", pts, gp))

	// PPG status
	if ppgGap <= 0 {
		sb.WriteString("🎉 JACK IS A PPG PLAYER! 🎉\n")
	} else {
		sb.WriteString(fmt.Sprintf("Needs %d more point%s to reach PPG\n",
			ppgGap, plural(ppgGap)))
	}

	// Next game
	if nextGame != nil {
		sb.WriteString(fmt.Sprintf("\nNext up: vs %s on %s\n\n",
			nextGame.OpponentName(), nextGame.FormattedDate()))
	}

	// Hashtag
	sb.WriteString("#NJDevils")

	return sb.String()
}

func pointsStr(pts int) string {
	if pts == 1 {
		return "1 point"
	}
	return fmt.Sprintf("%d points", pts)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
