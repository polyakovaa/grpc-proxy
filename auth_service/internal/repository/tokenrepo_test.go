package repository_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/model"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/repository"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestTokenRepo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock db: %v", err)
	}
	defer db.Close()

	repo := repository.NewTokenRepository(db)

	userId := uuid.NewString()
	tokenHash := "123hash"
	accessId := uuid.NewString()
	exp := time.Now().Add(24 * time.Hour)

	t.Run("CreateRefreshToken success", func(t *testing.T) {

		token := &model.RefreshToken{
			UserID:        userId,
			TokenHash:     tokenHash,
			AccessTokenID: accessId,
			ExpiresAt:     exp,
		}

		mock.ExpectExec(`INSERT INTO refresh_tokens`).WithArgs(userId, tokenHash, accessId, exp).WillReturnResult(sqlmock.NewResult(1, 1))
		err = repo.CreateRefreshToken(token)
		assert.NoError(t, err)
	})

	t.Run("FindByTokenHash success", func(t *testing.T) {
		rawToken := "my_refresh_token"
		hashedToken1, _ := bcrypt.GenerateFromPassword([]byte("wrong_token"), bcrypt.DefaultCost)
		hashedToken2, _ := bcrypt.GenerateFromPassword([]byte(rawToken), bcrypt.DefaultCost)

		rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at"}).
			AddRow(1, "user1", string(hashedToken1), exp).
			AddRow(2, userId, string(hashedToken2), exp)
		mock.ExpectQuery(`SELECT id, user_id, token_hash, expires_at FROM refresh_tokens
		 WHERE expires_at > NOW\(\)`).WillReturnRows(rows)
		token, err := repo.FindByTokenHash(rawToken)

		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, userId, token.UserID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

}
