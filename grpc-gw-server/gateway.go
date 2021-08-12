package main

import (
	"context"
	"log"
	"net/http"

	userpb "example.com/grpc-mongo-crud/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

const (
	server_port      = "0.0.0.0:50051"
	grpc_server_port = "0.0.0.0:50050"
)

func main() {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	options := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	if err := userpb.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		mux,
		server_port,
		options,
	); err != nil {
		log.Fatalf("failed to register gRPC gateway: %v", err)
	}

	log.Printf("start HTTP server on %s", grpc_server_port)
	if err := http.ListenAndServe(grpc_server_port, mux); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

}
