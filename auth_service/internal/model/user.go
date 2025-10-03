package model

type User struct {
	ID           int    `json:"id" db:"id"`
	Login        string `json:"login" db:"login"`
	Email        string `json:"email" db:"email"`
	PasswordHash string `json:"-" db:"password_hash"`
}
