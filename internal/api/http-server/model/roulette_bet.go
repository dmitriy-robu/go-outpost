package model

import (
	"go-outpost/internal/api/config"
	"time"
)

type RouletteBet struct {
	ID         int64        `json:"id"`
	RouletteID int64        `json:"roulette_id"`
	Amount     int          `json:"amount"`
	Color      config.Color `json:"color"`
	UserID     int64        `json:"user_id"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}
