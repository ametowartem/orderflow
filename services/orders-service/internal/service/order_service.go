package service

import (
	"time"

	"github.com/ametowartem/orderflow/orders-service/internal/domain"
	"github.com/google/uuid"
)

type OrderStore interface {
	Create(domain.Order) error
	GetByID(id string) (domain.Order, error)
	List() ([]domain.Order, error)
}

type OrderService struct {
	store OrderStore
}

func NewOrderService(store OrderStore) *OrderService {
	return &OrderService{store: store}
}

func (s *OrderService) CreateOrder(userID string, items []domain.OrderItem) (domain.Order, error) {
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
