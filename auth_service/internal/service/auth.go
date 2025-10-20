package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/model"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	tokenRepo  *repository.TokenRepository
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func (s *AuthService) AccessTTL() time.Duration {
	return s.accessTTL
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

func (s *AuthService) GenerateTokens(userID string) (*model.Token, error) {
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	accessID := uuid.New().String()

	claims := jwt.MapClaims{
		"user_id":    userID,
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

func (s *AuthService) ValidateAccessToken(tokenStr string) (*model.User, time.Time, bool) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		log.Printf("Invaild access token: %v", err)
		return nil, time.Time{}, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, time.Time{}, false
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, time.Time{}, false
	}

	exp := time.Unix(int64(claims["expires_at"].(float64)), 0)
	if exp.Before(time.Now()) {
		return nil, exp, false
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, exp, false
	}

	return user, exp, true

}

func (s *AuthService) RegisterUser(username, email, password string) (*model.User, error) {
	_, err := s.userRepo.FindByEmail(email)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &model.User{
		UserName:     username,
		Email:        email,
		PasswordHash: string(hashed),
	}
	user, err := s.userRepo.CreateUser(u)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil

}

func (s *AuthService) RefreshToken(oldRefreshToken string) (*model.Token, error) {

	storedToken, err := s.tokenRepo.FindByTokenHash(oldRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	_ = s.tokenRepo.DeleteByID(storedToken.ID)

	return s.GenerateTokens(storedToken.UserID)
}

func (s *AuthService) Authenticate(email, password string) (*model.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	return user, nil
}
