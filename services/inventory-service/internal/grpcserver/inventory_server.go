package grpcserver

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ametowartem/orderflow/inventory-service/internal/repository/sqlcgen"
	inventoryv1 "github.com/ametowartem/orderflow/proto/inventory/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InventoryServer struct {
	inventoryv1.UnimplementedInventoryServiceServer
	queries *sqlcgen.Queries
	pool    *pgxpool.Pool
}

func NewInventoryServer(queris *sqlcgen.Queries, pool *pgxpool.Pool) *InventoryServer {
	return &InventoryServer{queries: queris, pool: pool}
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

func (s *InventoryServer) ReserveStock(ctx context.Context, req *inventoryv1.ReserveStockRequest) (*inventoryv1.ReserveStockResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can not start transaction: %v", err)
	}

	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			log.Printf("failed to rollback transaction: %v", rollbackErr)
		}
	}()

	qtx := s.queries.WithTx(tx)

	var productID pgtype.UUID

	if err = productID.Scan(req.ProductId); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product_id: %v", err)
	}

	product, err := qtx.GetProductForUpdate(ctx, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "product %s not found", req.ProductId)
		}
		return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
	}

	if product.StockQuantity < req.Quantity {
		return nil, status.Errorf(codes.FailedPrecondition, "insufficient stock for product %s", productID)
	}

	if err = qtx.DecrementStock(ctx, sqlcgen.DecrementStockParams{
		ID:            productID,
		StockQuantity: req.Quantity,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to decrement stock quantity: %v", err)

	}

	var orderID pgtype.UUID
	if err = orderID.Scan(req.OrderId); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order_id: %v", err)
	}

	reservation, err := qtx.CreateReservation(ctx, sqlcgen.CreateReservationParams{
		ProductID: productID,
		OrderID:   orderID,
		Quantity:  req.Quantity,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(15 * time.Minute), Valid: true},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create reservation: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)

	}

	return &inventoryv1.ReserveStockResponse{ReservationId: reservation.ID.String()}, nil
}
