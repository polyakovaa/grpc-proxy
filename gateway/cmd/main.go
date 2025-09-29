package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/polyakovaa/grpcproxy/gateway/config"
	"github.com/polyakovaa/grpcproxy/gateway/internal/handler"
	"github.com/polyakovaa/grpcproxy/gen/auth"
	"github.com/polyakovaa/grpcproxy/gen/event"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	authClient, authConn, err := NewAuthClient(cfg.Services["auth"].Address)
	if authConn != nil {
		defer authConn.Close()
	}

	if err != nil {
		log.Println("Error connecting Auth service")
	}

	eventClient, eventConn, err := NewEventClient(cfg.Services["event"].Address)
	if eventConn != nil {
		defer eventConn.Close()
	}

	if err != nil {
		log.Println("Error connecting Event service")
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"services": gin.H{
				"auth":  getServiceStatus(authConn),
				"event": getServiceStatus(eventConn),
			},
		})
	})

	authHandler := handler.NewAuthHandler(authClient)
	eventHandler := handler.NewEventHandler(eventClient, authClient)

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/logout", authHandler.Logout)
	}

	eventGroup := router.Group("/events")
	{
		eventGroup.GET("/:id", eventHandler.GetEvent)
		eventGroup.POST("/", eventHandler.CreateEvent)
		eventGroup.POST("/:id/join", eventHandler.JoinEvent)
	}

	log.Printf("Gateway running on :%s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func NewAuthClient(address string) (auth.AuthServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return auth.NewAuthServiceClient(conn), conn, nil
}

func NewEventClient(address string) (event.EventServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return event.NewEventServiceClient(conn), conn, nil
}

func getServiceStatus(conn *grpc.ClientConn) string {
	if conn == nil {
		return "unavailable"
	}

	state := conn.GetState()
	switch state {
	case connectivity.Ready:
		return "healthy"
	case connectivity.Connecting, connectivity.Idle:
		return "connecting"
	case connectivity.TransientFailure:
		return "unhealthy"
	case connectivity.Shutdown:
		return "shutdown"
	default:
		return state.String()
	}
}
