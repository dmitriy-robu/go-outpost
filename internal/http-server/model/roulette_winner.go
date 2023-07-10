package model

import (
	"go-outpost/internal/config"
	"time"
)

type RouletteWinner struct {
	ID         int64        `json:"id"`
	RouletteID int64        `json:"roulette_id"`
	Color      config.Color `json:"color"`
	Number     int          `json:"number"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}
