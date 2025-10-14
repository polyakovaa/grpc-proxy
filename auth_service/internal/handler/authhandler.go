package handler

import (
	"context"
	"log"
	"time"

	"github.com/polyakovaa/grpcproxy/auth_service/internal/service"
	"github.com/polyakovaa/grpcproxy/gen/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHandler struct {
	auth.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewAuthHandler(authtService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authtService,
	}
}

func (h *AuthHandler) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.AuthResponse, error) {
	user, err := h.authService.RegisterUser(req.UserName, req.Email, req.Password)
	if err != nil {
		log.Printf("Failed to register user %v", err)
		return nil, err
	}

	token, err := h.authService.GenerateTokens(user.ID)
	if err != nil {
		log.Printf("Failed to generate tokens for user %v", err)
		return nil, err
	}

	return &auth.AuthResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    timestamppb.New(time.Now().Add(h.authService.AccessTTL())),
		UserId:       user.ID,
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.UserResponse, error) {
	user, exp, valid := h.authService.ValidateAccessToken(req.Token)
	if !valid {
		return &auth.UserResponse{Valid: false}, nil
	}
	return &auth.UserResponse{
		Valid:          true,
		UserId:         user.ID,
		Email:          user.Email,
		UserName:       user.UserName,
		TokenExpiresAt: timestamppb.New(exp),
	}, nil

}

func (h *AuthHandler) Login(ctx context.Context, req *auth.LoginRequest) (*auth.AuthResponse, error) {
	user, err := h.authService.Authenticate(req.Email, req.Password)
	if err != nil {
		log.Printf("Failed login for %s: %v", req.Email, err)
		return nil, err
	}
	token, err := h.authService.GenerateTokens(user.ID)
	if err != nil {
		log.Printf("Failed to generate tokens for user %v", err)
		return nil, err
	}

	return &auth.AuthResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    timestamppb.New(time.Now().Add(h.authService.AccessTTL())),
		UserId:       user.ID,
	}, nil

}
func (h *AuthHandler) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.AuthResponse, error) {
	newTokens, err := h.authService.RefreshToken(req.AccessToken, req.RefreshToken)
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		return nil, err
	}

	return &auth.AuthResponse{
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
		ExpiresAt:    timestamppb.New(time.Now().Add(h.authService.AccessTTL())),
	}, nil
}
