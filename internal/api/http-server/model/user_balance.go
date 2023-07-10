package model

import "time"

type UserBalance struct {
	ID        int64      `json:"id"`
	Balance   int        `json:"balance"`
	UserID    int64      `json:"user_id"`
	UpdatedAt *time.Time `json:"updated_at"`
}
