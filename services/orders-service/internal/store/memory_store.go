package store

import (
	"maps"
	"slices"
	"sync"

	"github.com/ametowartem/orderflow/orders-service/internal/domain"
)

type MemoryStore struct {
	mu     sync.Mutex
	orders map[string]domain.Order
}

func (s *MemoryStore) Create(order domain.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[order.ID] = order

	return nil
}

func (s *MemoryStore) GetByID(id string) (domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.orders[id]
	if !ok {
		return domain.Order{}, domain.ErrNotFound
	}

	return val, nil
}

func (s *MemoryStore) List() ([]domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	orders := slices.Collect(maps.Values(s.orders))

	return orders, nil
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{orders: make(map[string]domain.Order)}
}
