package repository_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/model"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestTokenRepo(t *testing.T) {

	t.Run("CreateRefreshToken success", func(t *testing.T) {
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

}
