package grpcserver

import (
	"context"
	"errors"

	"github.com/ametowartem/orderflow/inventory-service/internal/repository/sqlcgen"
	inventoryv1 "github.com/ametowartem/orderflow/proto/inventory/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InventoryServer struct {
	inventoryv1.UnimplementedInventoryServiceServer
	queries *sqlcgen.Queries
}

func NewInventoryServer(queris *sqlcgen.Queries) *InventoryServer {
	return &InventoryServer{queries: queris}
}

func (s *InventoryServer) CheckStock(ctx context.Context, req *inventoryv1.CheckStockRequest) (*inventoryv1.CheckStockResponse, error) {

	var productID pgtype.UUID

	if err := productID.Scan(req.ProductId); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product_id: %v", err)
	}

	product, err := s.queries.GetProductByID(ctx, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "product %s not found", req.ProductId)
		}
		return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
	}

	available := product.StockQuantity >= req.Quantity
	return &inventoryv1.CheckStockResponse{Available: available}, nil

}
