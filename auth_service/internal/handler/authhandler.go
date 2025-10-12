package handler

import (
	"github.com/polyakovaa/grpcproxy/auth_service/internal/service"
	"github.com/polyakovaa/grpcproxy/gen/auth"
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
