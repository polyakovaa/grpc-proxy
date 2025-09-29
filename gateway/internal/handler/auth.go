package handler

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polyakovaa/grpcproxy/gateway/internal/utils"
	"github.com/polyakovaa/grpcproxy/gen/auth"
)

type AuthHandler struct {
	authClient auth.AuthServiceClient
}

func NewAuthHandler(authClient auth.AuthServiceClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

func (h *AuthHandler) Register(c *gin.Context) {

	if h.authClient == nil {
		c.JSON(503, gin.H{"error": "Auth service unavailable"})
		return
	}

	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		UserName string `json:"user_name"`
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	response, err := h.authClient.Register(c.Request.Context(), &auth.RegisterRequest{
		Email:    request.Email,
		Password: request.Password,
		UserName: request.UserName,
	})
	if err != nil {
		utils.HandleGRPCError(c, err)
		log.Printf("Register error: %v", err)
		return
	}

	c.SetCookie("refresh_token", response.RefreshToken, int(response.ExpiresAt.AsTime().Unix()), "/", "", false, true)

	c.JSON(201, gin.H{
		"user_id":       response.UserId,
		"access_token":  response.AccessToken,
		"expires_at":    response.ExpiresAt.AsTime().Format(time.RFC3339),
		"refresh_token": response.RefreshToken,
	})

}

func (h *AuthHandler) Login(c *gin.Context) {
	if h.authClient == nil {
		c.JSON(503, gin.H{"error": "Auth service unavailable"})
		return
	}

	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	response, err := h.authClient.Login(c.Request.Context(), &auth.LoginRequest{
		Email:    request.Email,
		Password: request.Password,
	})

	if err != nil {
		utils.HandleGRPCError(c, err)
		log.Printf("Login error: %v", err)
		return
	}

	c.SetCookie("refresh_token", response.RefreshToken, int(response.ExpiresAt.AsTime().Unix()), "/", "", false, true)

	c.JSON(200, gin.H{
		"user_id":       response.UserId,
		"access_token":  response.AccessToken,
		"expires_at":    response.ExpiresAt.AsTime().Format(time.RFC3339),
		"refresh_token": response.RefreshToken,
	})

}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	if h.authClient == nil {
		c.JSON(503, gin.H{"error": "auth service unavailable"})
		return
	}

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(401, gin.H{"error": "refresh token required"})
		return
	}

	response, err := h.authClient.RefreshToken(c.Request.Context(), &auth.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})

	if err != nil {
		utils.HandleGRPCError(c, err)
	}

	c.SetCookie("refresh_token", response.RefreshToken, int(response.ExpiresAt.AsTime().Unix()), "/", "", false, true)

	c.JSON(200, gin.H{
		"access_token": response.AccessToken,
		"expires_at":   response.ExpiresAt.AsTime().Format(time.RFC3339),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.JSON(200, gin.H{"message": "successfully logged out"})
}
