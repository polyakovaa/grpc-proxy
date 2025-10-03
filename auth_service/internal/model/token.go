package model

import "time"

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	TokenHash     string    `json:"-" db:"token_hash"`
	AccessTokenID string    `json:"-" db:"access_token_id"`
	UserID        string    `json:"-" db:"user_id"`
	Expires       time.Time `json:"-" db:"expires"`
}
