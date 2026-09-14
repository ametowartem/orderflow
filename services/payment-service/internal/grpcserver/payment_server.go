package grpcserver

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
	"time"

	"github.com/ametowartem/orderflow/payment-service/internal/kafka"
	"github.com/ametowartem/orderflow/payment-service/internal/repository/sqlcgen"
	paymentv1 "github.com/ametowartem/orderflow/proto/payment/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventPublisher interface {
	Publish(ctx context.Context, key string, event any) error
}

type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	queries   *sqlcgen.Queries
	pool      *pgxpool.Pool
	publisher EventPublisher
}

func NewPaymentServer(queries *sqlcgen.Queries, pool *pgxpool.Pool, publisher EventPublisher) *PaymentServer {
	return &PaymentServer{queries: queries, pool: pool, publisher: publisher}
}

func (s *PaymentServer) InitiatePayment(ctx context.Context, req *paymentv1.InitiatePaymentRequest) (*paymentv1.InitiatePaymentResponse, error) {
	var orderID pgtype.UUID

	if err := orderID.Scan(req.OrderId); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order_id: %v", err)
	}

	existing, err := s.queries.GetPaymentByOrderID(ctx, orderID)

	if err == nil {
		return &paymentv1.InitiatePaymentResponse{PaymentId: existing.ID.String(), Status: existing.Status}, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, "check existing payment: %v", err)
	}

	var amount pgtype.Numeric

	if err = amount.Scan(req.Amount); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid amount: %v", err)
	}

	payment, err := s.queries.CreatePayment(ctx, sqlcgen.CreatePaymentParams{OrderID: orderID, Amount: amount})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create payment: %v", err)
	}

	go s.processPaymentAsync(context.WithoutCancel(ctx), payment.ID, req.Amount, req.OrderId)

	return &paymentv1.InitiatePaymentResponse{PaymentId: payment.ID.String(), Status: "processing"}, nil

}

func (s *PaymentServer) processPaymentAsync(ctx context.Context, paymentID pgtype.UUID, _ float64, orderID string) {
	time.Sleep(2 * time.Second)

	//nolint:gosec
	success := rand.Float64() > 0.2

	newStatus := "completed"
	if !success {
		newStatus = "failed"
	}

	if err := s.queries.UpdatePaymentStatus(ctx, sqlcgen.UpdatePaymentStatusParams{ID: paymentID, Status: newStatus}); err != nil {
		log.Printf("failed to update payment status: %v", err)
		return
	}

	event := kafka.PaymentEvent{PaymentID: paymentID.String(), Success: success, OrderID: orderID}
	if err := s.publisher.Publish(ctx, orderID, event); err != nil {
		log.Printf("failed to publish payment event: %v", err)
	}
}
