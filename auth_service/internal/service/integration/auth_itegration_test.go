package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/repository"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *sql.DB {
	composePath := filepath.Join("..", "..", "docker-compose_test.yaml")
	cmd := exec.Command("docker-compose", "-f", composePath, "up", "-d", "--wait")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start container: %v\nOutput: %s", err, string(output))
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		"localhost",
		5544,
		"test_user",
		"test_password",
		"test_db",
		"disable",
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect to test DB: %v", err)
	}

	return db
}

func setupTestDBContainers(t *testing.T) *sql.DB {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	connStr = connStr + " sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE IF NOT EXISTS users(
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_name  TEXT NOT NULL UNIQUE,
			email  TEXT NOT NULL UNIQUE,
			password_hash  TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS refresh_tokens(
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL,
			access_token_id TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL
		);`)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
		container.Terminate(ctx)
	})

	return db
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	db := setupTestDBContainers(t)

	urepo := repository.NewUserRepository(db)
	trepo := repository.NewTokenRepository(db)

	svc := service.NewAuthService(urepo, trepo, "secret123", time.Minute*15, time.Hour*24)

	t.Run("Registration and login success", func(t *testing.T) {
		userName := "integr_test"
		email := "user_test@example.com"
		pass := "passwordhash123"
		u, err := svc.RegisterUser(userName, email, pass)
		assert.NoError(t, err)
		assert.NotNil(t, u)
		assert.Equal(t, email, u.Email)
		assert.Equal(t, userName, u.UserName)
		assert.NotEmpty(t, u.ID)

		_, err = svc.RegisterUser(userName, email, pass)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		loggedInUser, err := svc.Login(email, pass)
		assert.NoError(t, err)
		assert.NotNil(t, loggedInUser)
		assert.Equal(t, u.ID, loggedInUser.ID)
		assert.Equal(t, email, loggedInUser.Email)

		_, err = svc.Login(email, "wrongpassword")
		assert.Error(t, err)

	})

	t.Run("Token Generation", func(t *testing.T) {
		userName := "token_test"
		email := "token_test@example.com"
		pass := "passwordhash123"

		user, err := svc.RegisterUser(userName, email, pass)
		require.NoError(t, err)

		tokenPair, err := svc.GenerateTokens(user.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)

		u, exp, ok := svc.ValidateAccessToken(tokenPair.AccessToken)
		assert.True(t, ok, "access token must be valid")
		assert.Equal(t, user.ID, u.ID)
		assert.True(t, exp.After(time.Now()))

		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1`, user.ID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

	})

}
