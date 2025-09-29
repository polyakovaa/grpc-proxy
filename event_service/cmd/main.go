package main

import (
	"log"
	"net"

	_ "github.com/lib/pq"

	"github.com/polyakovaa/grpcproxy/event_service/config"
	"github.com/polyakovaa/grpcproxy/event_service/internal/handler"
	"github.com/polyakovaa/grpcproxy/event_service/internal/repository"
	"github.com/polyakovaa/grpcproxy/event_service/internal/service"
	"github.com/polyakovaa/grpcproxy/gen/event"

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

	eventRepo := repository.NewEventRepository(db)
	eventService := service.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventService)

	lis, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	event.RegisterEventServiceServer(grpcServer, eventHandler)

	log.Printf("Event service running on :%s", cfg.Server.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
