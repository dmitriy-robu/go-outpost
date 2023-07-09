package model

import "time"

type ProvablyFair struct {
	ID                   int64     `json:"id"`
	GameDrawID           int64     `json:"game_draw_id"`
	ClientSeed           string    `json:"client_seed"`
	ServerSeed           string    `json:"server_seed"`
	ResultedHash         string    `json:"resulted_hash"`
	ResultedRandomNumber float64   `json:"resulted_random_number"`
	Min                  int       `json:"min"`
	Max                  int       `json:"max"`
	Nonce                int       `json:"nonce"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
