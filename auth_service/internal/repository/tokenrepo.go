package repository

import (
	"database/sql"
	"errors"

	"github.com/polyakovaa/grpcproxy/auth_service/internal/model"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

func (r *TokenRepository) CreateRefreshToken(token *model.RefreshToken) error {

	query := `INSERT INTO refresh_tokens (user_id, token_hash, access_token_id, expires_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(
		query,
		token.UserID,
		token.TokenHash,
		token.AccessTokenID,
		token.ExpiresAt,
	)

	return err
}

func (r *TokenRepository) FindRefreshToken(accessTokenID string) (*model.RefreshToken, error) {
	t := &model.RefreshToken{}
	query := `SELECT user_id, token_hash, expires_at FROM refresh_tokens WHERE access_token_id = $1 AND expires_at > NOW()`

	err := r.db.QueryRow(query, accessTokenID).Scan(
		&t.UserID,
		&t.TokenHash,
		&t.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("token not found or expired")
	}
	return t, err
}
