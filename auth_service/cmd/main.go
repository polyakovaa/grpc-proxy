package main

import (
	"log"
	"net"

	"github.com/polyakovaa/grpcproxy/auth_service/config"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/handler"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/repository"
	"github.com/polyakovaa/grpcproxy/auth_service/internal/service"
	"github.com/polyakovaa/grpcproxy/gen/auth"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := config.ConnectToDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	authService := service.NewAuthService(
		userRepo,
		tokenRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
	)
	authHandler := handler.NewAuthHandler(authService)

	lis, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	auth.RegisterAuthServiceServer(grpcServer, authHandler)

	log.Printf("Auth service running on :%s", cfg.Server.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
