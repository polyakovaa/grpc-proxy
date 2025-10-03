package storage

import (
	"auth_service/config"
	"database/sql"
)

type Storage struct {
	config          *config.Config
	db              *sql.DB
	userRepository  *storage.UserRepository
	tokenRepository *storage.TokenRepository
}

func NewStorage(config *Config) *Storage {
	return &Storage{
		config: config,
	}
}
