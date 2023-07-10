package model

import (
	"go-outpost/internal/api/config"
	"time"
)

type GameDraw struct {
	ID        int64       `json:"id"`
	GameID    int64       `json:"game_id"`
	UserID    int64       `json:"user_id"`
	Game      config.Game `json:"game"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
