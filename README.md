# Jack Hughes PPG Tracker 🏒😈

(README is 100% AI generated, with the exception of this line.)

A Bluesky bot that tracks the inevitable, unstoppable, mathematically-certain journey of Jack Hughes — the most gifted hockey player of his generation, a generational talent so blindingly obvious that even casual fans stop mid-bite of their hot dog to watch him skate — toward becoming a points-per-game player in the NHL.

## What Is This?

Every morning after a New Jersey Devils game, this bot posts to [@jackppgtracker.bsky.social](https://bsky.app/profile/jackppgtracker.bsky.social) with:

- How many points Jack had in last night's game
- His career totals (games played and points)
- How many points he still needs to reach PPG
- Who the Devils play next and when

It will keep posting until Jack achieves PPG status, at which point it will erupt in digital celebration and then presumably retire in glory.

## About Jack Hughes

Jack Hughes is, to put it mildly, a bit good at hockey.

Drafted first overall in 2019 by the New Jersey Devils, Hughes has spent his career casually rewriting the franchise record books while making it look effortless. His skating is the kind that makes you question whether the ice has different physics when he's on it. His hands are so fast that video review has been required on multiple occasions just to confirm that yes, he actually did that. His hockey IQ is so advanced that scientists are still debating whether it constitutes a separate branch of mathematics.

In the 2022-23 season, Hughes put up 99 points — a New Jersey Devils franchise record — and had the audacity to record his 99th point on an assist on his younger brother Luke's first career NHL goal, in the final game of the season, on an overtime winner. If you wrote that in a movie script, it would be rejected for being too unrealistic.

He then went and scored the overtime gold medal winner for the United States at the 2026 Winter Olympics in Milan, because apparently he was bored.

The PPG milestone isn't a question of _if_. It's a question of _when_, and this bot is here to count down every single point.

## How It Works

Built in Go, deployed via GitHub Actions.

```
├── main.go                          # Orchestration
├── nhl/client.go                    # NHL API integration
├── bluesky/client.go                # AT Protocol / Bluesky posting
└── .github/workflows/tracker.yml   # Runs every morning at 8 AM ET
```

The bot:

1. Checks if the Devils played yesterday (or late last night)
2. If yes and the game is final, fetches Hughes' stats from the boxscore
3. Pulls his career totals from the NHL API
4. Calculates the PPG gap
5. Fetches the next Devils game
6. Posts a summary to Bluesky with proper hashtag facets

## Setup

### Prerequisites

- Go 1.22+
- A Bluesky account (we recommend a dedicated bot account)
- A GitHub repository

### Configuration

Add the following secrets to your GitHub repository (Settings → Secrets and variables → Actions):

| Secret              | Value                                                 |
| ------------------- | ----------------------------------------------------- |
| `BSKY_HANDLE`       | Your Bluesky handle e.g. `jackppgtracker.bsky.social` |
| `BSKY_APP_PASSWORD` | A Bluesky app password (not your main password)       |

### Running Locally

```bash
export BSKY_HANDLE=yourhandle.bsky.social
export BSKY_APP_PASSWORD=your-app-password
go run .
```

### Deployment

Push to `main` and GitHub Actions handles the rest. The workflow runs automatically at 1 PM UTC (8 AM ET) daily. You can also trigger it manually from the Actions tab.

## Contributing

Found a bug? The NHL API is notoriously undocumented and occasionally returns whatever it feels like, so contributions and fixes are welcome. Open a PR.

## License

MIT. Share it, fork it, adapt it for your own favourite player. Just know that no other player is Jack Hughes.

---

_"He's the best player I've ever seen." — Probably everyone who has watched him play_
