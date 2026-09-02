package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ametowartem/orderflow/orders-service/internal/domain"
	"github.com/google/uuid"
)

type OrderStore interface {
	Create(domain.Order) error
	GetByID(id string) (domain.Order, error)
	List() ([]domain.Order, error)
}

type StockChecker interface {
	CheckStock(ctx context.Context, productID string, quantity int32) (bool, error)
}

type OrderService struct {
	store        OrderStore
	stockChecker StockChecker
}

func NewOrderService(store OrderStore, stockChecker StockChecker) *OrderService {
	return &OrderService{store: store, stockChecker: stockChecker}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID string, items []domain.OrderItem) (domain.Order, error) {

	id := uuid.New()
	status := domain.StatusPending

	var total float64
	for i := range items {
		items[i].ID = uuid.NewString()

		if items[i].PriceAtOrder <= 0 {
			return domain.Order{}, domain.ErrInvalidAmount
		}

		total += items[i].PriceAtOrder * float64(items[i].Quantity)
	}

	if total > 1_000_000 {
		return domain.Order{}, domain.ErrAmountTooHigh
	}

	for _, item := range items {
		available, err := s.stockChecker.CheckStock(ctx, item.ProductID, int32(item.Quantity))

		if err != nil {
			return domain.Order{}, fmt.Errorf("check stock for %s: %w", item.ProductID, err)
		}
		if !available {
			return domain.Order{}, fmt.Errorf("product %s: %w", item.ProductID, domain.ErrInsufficientStock)
		}
	}

	order := domain.Order{
		ID:          id.String(),
		UserID:      userID,
		Status:      status,
		TotalAmount: total,
		Items:       items,
		CreatedAt:   time.Now(),
	}
	if err := s.store.Create(order); err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func (s *OrderService) GetOrderById(id string) (domain.Order, error) {

	order, err := s.store.GetByID(id)
	if err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func (s *OrderService) GetOrders() ([]domain.Order, error) {
	orders, err := s.store.List()
	if err != nil {
		return nil, err
	}

	return orders, nil
}
