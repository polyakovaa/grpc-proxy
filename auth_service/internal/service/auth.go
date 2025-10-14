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

func (s *AuthService) RefreshToken(oldAccessToken, oldRefreshToken string) (*model.Token, error) {
	claims, err := s.parseTokenClaims(oldAccessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	accessID, ok := claims["token_id"].(string)
	if !ok {
		return nil, errors.New("invalid claims: missing jti")
	}

	isValid, err := s.validateRefreshToken(oldRefreshToken, accessID)
	if err != nil || !isValid {
		log.Printf("Invalid refresh token for user %s", claims["user_id"].(string))
		return nil, errors.New("invalid refresh token")
	}

	_ = s.tokenRepo.DeleteByAccessTokenID(accessID)

	return s.GenerateTokens(claims["user_id"].(string))
}

func (s *AuthService) parseTokenClaims(tokenString string) (jwt.MapClaims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Printf("Error parsing token: %v", err)
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil
}

func (s *AuthService) validateRefreshToken(rawToken, accessTokenID string) (bool, error) {
	storedToken, err := s.tokenRepo.FindRefreshToken(accessTokenID)
	if err != nil {
		log.Printf("Error finding refresh token in database: %v", err)
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedToken.TokenHash), []byte(rawToken))
	return err == nil, nil
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
		return nil, fmt.Errorf("user already exists")
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

func (s *AuthService) Authenticate(email, password string) (*model.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}
