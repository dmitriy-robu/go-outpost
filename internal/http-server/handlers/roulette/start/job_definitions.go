package start

import "go-outpost/internal/config"

type RouletteUpdatePlayedAtJob struct {
	rouletteID int64
}

func (j *RouletteUpdatePlayedAtJob) Execute() {
	// update roulette played_at time
}

type RouletteWinnerJob struct {
	winColor  config.Color
	winNumber int
}

func (j *RouletteWinnerJob) Execute() {
	// notify about roulette winner
}

type RouletteBetReplicateJob struct {
	rouletteID int64
}

func (j *RouletteBetReplicateJob) Execute() {
	// replicate roulette bet
}
