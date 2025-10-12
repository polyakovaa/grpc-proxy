package model

import "time"

type User struct {
	ID           string    `db:"id"`
	UserName     string    `db:"user_name"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}
