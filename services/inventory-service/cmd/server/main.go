package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/ametowartem/orderflow/inventory-service/internal/grpcserver"
	"github.com/ametowartem/orderflow/inventory-service/internal/repository"
	"github.com/ametowartem/orderflow/inventory-service/internal/repository/sqlcgen"
	inventoryv1 "github.com/ametowartem/orderflow/proto/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=inventory-postgres user=inventory_user password=inventory_pass dbname=inventory_db port=5432 sslmode=disable"
	}

	pool, err := repository.NewPool(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}

	queries := sqlcgen.New(pool)
	server := grpcserver.NewInventoryServer(queries)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()

	inventoryv1.RegisterInventoryServiceServer(grpcServer, server)
	reflection.Register(grpcServer)
	log.Println("inventory-service gRPC listening on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
