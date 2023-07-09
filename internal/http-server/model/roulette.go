package model

import (
	"github.com/google/uuid"
	"time"
)

type Roulette struct {
	ID        int64      `json:"id"`
	UUID      uuid.UUID  `json:"uuid"`
	Round     int64      `json:"round"`
	PlayedAt  *time.Time `json:"played_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
