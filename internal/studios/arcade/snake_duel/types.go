// Package snake_duel implements the Snake Duel game for the Arcade Studio.
// Two AI snakes compete on a 30x24 grid, rendered with Unicode half-blocks.
package snake_duel

// Grid dimensions — must match daemon's snake_duel.hrl.
const (
	GridWidth  = 30
	GridHeight = 24
)

// Point is an [x, y] grid coordinate (0-indexed, origin top-left).
type Point [2]int

// GameEvent records something that happened during a game tick.
type GameEvent struct {
	Type  string `json:"type"`  // food, turn, collision, win, poison_drop, poison_eat
	Value string `json:"value"` // human-readable description
	Tick  int    `json:"tick"`  // game tick when event occurred
}

// Snake holds one player's state.
type Snake struct {
	Body          []Point     `json:"body"`
	Direction     string      `json:"direction"` // up, down, left, right
	Score         int         `json:"score"`
	AssFactor     int         `json:"asshole_factor"`
	Events        []GameEvent `json:"events"`
}

// PoisonApple is a poison pellet dropped by a snake.
type PoisonApple struct {
	Pos   Point  `json:"pos"`
	Owner string `json:"owner"` // player1, player2
}

// GameState is the full state broadcast per tick via SSE.
type GameState struct {
	MatchID      string        `json:"match_id"`
	Snake1       Snake         `json:"snake1"`
	Snake2       Snake         `json:"snake2"`
	Food         Point         `json:"food"`
	PoisonApples []PoisonApple `json:"poison_apples"`
	Status       string        `json:"status"`    // idle, countdown, running, finished
	Winner       string        `json:"winner"`    // none, player1, player2, draw
	Tick         int           `json:"tick"`
	Countdown    int           `json:"countdown"`
}

// StartMatchResponse is the JSON returned by POST /api/arcade/snake-duel/matches.
type StartMatchResponse struct {
	MatchID string `json:"match_id"`
	AF1     int    `json:"af1"`
	AF2     int    `json:"af2"`
	TickMs  int    `json:"tick_ms"`
	Status  string `json:"status"`
}
