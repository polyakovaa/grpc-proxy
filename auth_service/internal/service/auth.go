package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/model"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	tokenRepo  *repository.TokenRepository
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	secret string,
	accessTTL, refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtSecret:  secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *AuthService) GenerateTokens(userID, email string) (*model.Token, error) {
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	accessID := uuid.New().String()

	claims := jwt.MapClaims{
		"user_guid":  userID,
		"token_id":   accessID,
		"expires_at": time.Now().Add(s.accessTTL).Unix(),
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	accessToken, err := access.SignedString([]byte(s.jwtSecret))
	if err != nil {
		log.Printf("Error signing access token for user %s: %v", userID, err)
		return nil, err
	}

	b := make([]byte, 32)
	rand.Read(b)
	refreshID := base64.URLEncoding.EncodeToString(b)
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(refreshID), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}

	err = s.tokenRepo.CreateRefreshToken(&model.RefreshToken{
		UserID:        userID,
		TokenHash:     string(hashedToken),
		AccessTokenID: accessID,
		ExpiresAt:     time.Now().Add(s.refreshTTL),
	})
	if err != nil {
		log.Printf("Error storing refresh token for user %s: %v", userID, err)
		return nil, err
	}

	return &model.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshID,
	}, nil
}
