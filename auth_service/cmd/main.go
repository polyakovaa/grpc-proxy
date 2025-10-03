package main

import (
	"log"
	"net"

	"github.com/polyakovaa/grpcproxy/auth_service/internal/service"
	"github.com/polyakovaa/grpcproxy/gen/auth"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	auth.RegisterAuthServiceServer(grpcServer, &service.AuthServer{})

	log.Println("Auth service running on :50051")
	grpcServer.Serve(lis)
}
