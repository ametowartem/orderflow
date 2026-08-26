package service

import (
	"time"

	"github.com/ametowartem/orderflow/orders-service/internal/domain"
	"github.com/google/uuid"
)

type OrderStore interface {
	Create(domain.Order) error
	GetById(id string) (domain.Order, error)
	List() ([]domain.Order, error)
}

type OrderService struct {
	store OrderStore
}

func NewOrderService(store OrderStore) *OrderService {
	return &OrderService{store: store}
}

func (s *OrderService) CreateOrder(userID string, amount float64) (domain.Order, error) {
	id := uuid.New()
	status := domain.StatusPending

	order := domain.Order{ID: id.String(), UserID: userID, Status: status, TotalAmount: amount, CreatedAt: time.Now()}
	if err := s.store.Create(order); err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func (s *OrderService) GetOrderById(id string) (domain.Order, error) {

	order, err := s.store.GetById(id)
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
