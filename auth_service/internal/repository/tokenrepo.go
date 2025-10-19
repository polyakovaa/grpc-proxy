package repository

import (
	"database/sql"
	"errors"

	"github.com/polyakovaa/grpcproxy/auth_service/internal/model"
	"golang.org/x/crypto/bcrypt"
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

func (r *TokenRepository) FindByTokenHash(rawToken string) (*model.RefreshToken, error) {
	rows := &model.RefreshToken{}
	query := `SELECT id, user_id, token_hash, expires_at FROM refresh_tokens WHERE expires_at > NOW()`

	err := r.db.QueryRow(query).Scan(
		&rows.ID,
		&rows.UserID,
		&rows.TokenHash,
		&rows.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("refresh token not found or expired")
	}

	if bcrypt.CompareHashAndPassword([]byte(rows.TokenHash), []byte(rawToken)) != nil {
		return nil, errors.New("invalid refresh token")
	}

	return rows, nil
}

func (r *TokenRepository) DeleteByID(id string) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
