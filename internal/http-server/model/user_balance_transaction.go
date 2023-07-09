package model

import (
	"encoding/json"
	"go-outpost/internal/config"
	"time"
)

type UserBalanceTransaction struct {
	ID        int64              `json:"id"`
	UserID    int64              `json:"user_id"`
	Type      config.BalanceType `json:"type"`
	Module    string             `json:"module"`
	Details   json.RawMessage    `json:"details"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}
