package model

import "time"

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	TokenHash     string    `json:"token_hash" db:"token_hash"`
	AccessTokenID string    `json:"access_token_id" db:"access_token_id"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
}
