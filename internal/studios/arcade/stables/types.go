// Package stables implements the Stables sub-app for the Arcade Studio.
// Train neuroevolution snake gladiators and pit champions against AI opponents.
package stables

// Stable represents a training stable from the daemon API.
type Stable struct {
	StableID             string  `json:"stable_id"`
	Status               string  `json:"status"` // training, completed, halted
	PopulationSize       int     `json:"population_size"`
	MaxGenerations       int     `json:"max_generations"`
	OpponentAF           int     `json:"opponent_af"`
	EpisodesPerEval      int     `json:"episodes_per_eval"`
	BestFitness          float64 `json:"best_fitness"`
	GenerationsCompleted int     `json:"generations_completed"`
	StartedAt            int64   `json:"started_at"`
	CompletedAt          *int64  `json:"completed_at"`
}

// Champion is the best-performing network from a stable.
type Champion struct {
	StableID    string  `json:"stable_id"`
	NetworkJSON string  `json:"network_json"`
	Fitness     float64 `json:"fitness"`
	Generation  int     `json:"generation"`
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	Draws       int     `json:"draws"`
	ExportedAt  int64   `json:"exported_at"`
}

// GenerationStats holds fitness data for one generation.
type GenerationStats struct {
	StableID    string  `json:"stable_id"`
	Generation  int     `json:"generation"`
	BestFitness float64 `json:"best_fitness"`
	AvgFitness  float64 `json:"avg_fitness"`
	WorstFitness float64 `json:"worst_fitness"`
	Timestamp   int64   `json:"timestamp"`
}

// TrainingProgress is a single SSE frame from the training stream.
type TrainingProgress struct {
	StableID        string  `json:"stable_id"`
	Status          string  `json:"status"` // training, completed, halted
	Generation      int     `json:"generation"`
	BestFitness     float64 `json:"best_fitness"`
	AvgFitness      float64 `json:"avg_fitness"`
	WorstFitness    float64 `json:"worst_fitness"`
	ChampionFitness float64 `json:"champion_fitness"`
	Running         bool    `json:"running"`
}

// StablesListResponse wraps the GET /stables response.
type StablesListResponse struct {
	OK      bool     `json:"ok"`
	Stables []Stable `json:"stables"`
}

// StableResponse wraps a single stable GET response.
type StableResponse struct {
	OK bool `json:"ok"`
	Stable
}

// ChampionResponse wraps the GET /champion response.
type ChampionResponse struct {
	OK bool `json:"ok"`
	Champion
}

// GenerationsResponse wraps the GET /generations response.
type GenerationsResponse struct {
	OK          bool              `json:"ok"`
	Generations []GenerationStats `json:"generations"`
}

// InitiateStableRequest is the POST body for creating a new stable.
type InitiateStableRequest struct {
	PopulationSize  int    `json:"population_size,omitempty"`
	MaxGenerations  int    `json:"max_generations,omitempty"`
	OpponentAF      int    `json:"opponent_af,omitempty"`
	EpisodesPerEval int    `json:"episodes_per_eval,omitempty"`
	SeedStableID    string `json:"seed_stable_id,omitempty"`
}

// InitiateStableResponse is the POST response for creating a new stable.
type InitiateStableResponse struct {
	OK             bool   `json:"ok"`
	StableID       string `json:"stable_id"`
	PopulationSize int    `json:"population_size"`
	MaxGenerations int    `json:"max_generations"`
	OpponentAF     int    `json:"opponent_af"`
	EpisodesPerEval int   `json:"episodes_per_eval"`
	Status         string `json:"status"`
}

// DuelResponse is returned by POST /stables/:id/duel.
type DuelResponse struct {
	OK      bool   `json:"ok"`
	MatchID string `json:"match_id"`
}
