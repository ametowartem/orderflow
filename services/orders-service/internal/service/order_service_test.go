package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ametowartem/orderflow/orders-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	createFn  func(domain.Order) error
	getByIdFn func(id string) (domain.Order, error)
	getFn     func() ([]domain.Order, error)
}

func (f *fakeStore) Create(order domain.Order) error {
	if f.createFn != nil {
		return f.createFn(order)
	}
	return nil
}

func (f *fakeStore) GetByID(id string) (domain.Order, error) {
	if f.getByIdFn != nil {
		return f.getByIdFn(id)
	}
	// Оставляем domain.ErrNotFound, если он определен в пакете domain.
	// В противном случае можно заменить на errors.New("not found")
	return domain.Order{}, domain.ErrNotFound
}

func (f *fakeStore) List() ([]domain.Order, error) {
	if f.getFn != nil {
		return f.getFn()
	}
	return []domain.Order{}, nil
}

type fakeStockChecker struct {
	checkStockFn func(
		ctx context.Context,
		productID string,
		quantity int32,
	) (bool, error)
}

func (f *fakeStockChecker) CheckStock(
	ctx context.Context,
	productID string,
	quantity int32,
) (bool, error) {
	if f.checkStockFn != nil {
		return f.checkStockFn(ctx, productID, quantity)
	}

	return true, nil
}

func TestOrderService_CreateOrder_Success(t *testing.T) {
	store := &fakeStore{}
	stockChecker := &fakeStockChecker{}

	svc := NewOrderService(store, stockChecker)

	items := []domain.OrderItem{
		{
			ProductID:    "prod-1",
			Quantity:     2,
			PriceAtOrder: 50.25,
		},
		{
			ProductID:    "prod-2",
			Quantity:     1,
			PriceAtOrder: 10.00,
		},
	}

	order, err := svc.CreateOrder(
		context.Background(),
		"user-123",
		items,
	)

	require.NoError(t, err)
	assert.Equal(t, "user-123", order.UserID)
	assert.Equal(t, 110.50, order.TotalAmount)
	assert.Equal(t, domain.StatusPending, order.Status)
	assert.NotEmpty(t, order.ID)
	assert.Len(t, order.Items, 2)
	assert.NotEmpty(t, order.Items[0].ID)
	assert.NotEmpty(t, order.Items[1].ID)
	assert.False(t, order.CreatedAt.IsZero())
}

// Тест на корректную обработку пустого списка товаров
func TestOrderService_CreateOrder_EmptyItems(t *testing.T) {
	store := &fakeStore{}
	stockChecker := &fakeStockChecker{}

	svc := NewOrderService(store, stockChecker)

	order, err := svc.CreateOrder(
		context.Background(),
		"user-123",
		[]domain.OrderItem{},
	)

	require.NoError(t, err)
	assert.Equal(t, "user-123", order.UserID)
	assert.Equal(t, 0.0, order.TotalAmount)
	assert.Equal(t, domain.StatusPending, order.Status)
	assert.NotEmpty(t, order.ID)
	assert.Empty(t, order.Items)
}

// Тест на проброс ошибки от хранилища.
// Заменяет старый тест на невалидную сумму, так как в новой реализации
// CreateOrder рассчитывает сумму автоматически на основе элементов и не
// выполняет валидацию на отрицательные значения.
func TestOrderService_CreateOrder_StoreError(t *testing.T) {
	expectedErr := errors.New("db connection failed")

	store := &fakeStore{
		createFn: func(order domain.Order) error {
			return expectedErr
		},
	}

	stockChecker := &fakeStockChecker{}

	svc := NewOrderService(store, stockChecker)

	items := []domain.OrderItem{
		{
			ProductID:    "prod-1",
			Quantity:     1,
			PriceAtOrder: 100.0,
		},
	}

	order, err := svc.CreateOrder(
		context.Background(),
		"user-123",
		items,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
	assert.Empty(t, order.ID)
}

func TestOrderService_CreateOrder_StockCheckError(t *testing.T) {
	expectedErr := errors.New("stock service unavailable")

	store := &fakeStore{}

	stockChecker := &fakeStockChecker{
		checkStockFn: func(
			ctx context.Context,
			productID string,
			quantity int32,
		) (bool, error) {
			assert.Equal(t, "prod-1", productID)
			assert.Equal(t, int32(2), quantity)

			return false, expectedErr
		},
	}

	svc := NewOrderService(store, stockChecker)

	items := []domain.OrderItem{
		{
			ProductID:    "prod-1",
			Quantity:     2,
			PriceAtOrder: 50,
		},
	}

	order, err := svc.CreateOrder(
		context.Background(),
		"user-123",
		items,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
	assert.Contains(t, err.Error(), "check stock for prod-1")
	assert.Empty(t, order.ID)
}

func TestOrderService_CreateOrder_InsufficientStock(t *testing.T) {
	store := &fakeStore{}

	stockChecker := &fakeStockChecker{
		checkStockFn: func(
			ctx context.Context,
			productID string,
			quantity int32,
		) (bool, error) {
			return false, nil
		},
	}

	svc := NewOrderService(store, stockChecker)

	items := []domain.OrderItem{
		{
			ProductID:    "prod-1",
			Quantity:     3,
			PriceAtOrder: 100,
		},
	}

	order, err := svc.CreateOrder(
		context.Background(),
		"user-123",
		items,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "product prod-1: insufficient stock")
	assert.Empty(t, order.ID)
}

func TestOrderService_CreateOrder_InvalidAmount(t *testing.T) {
	store := &fakeStore{}
	stockChecker := &fakeStockChecker{}

	svc := NewOrderService(store, stockChecker)

	items := []domain.OrderItem{
		{
			ProductID:    "prod-1",
			Quantity:     1,
			PriceAtOrder: 0,
		},
	}

	order, err := svc.CreateOrder(
		context.Background(),
		"user-123",
		items,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
	assert.Empty(t, order.ID)
}

func TestOrderService_CreateOrder_AmountTooHigh(t *testing.T) {
	store := &fakeStore{}
	stockChecker := &fakeStockChecker{}

	svc := NewOrderService(store, stockChecker)

	items := []domain.OrderItem{
		{
			ProductID:    "prod-1",
			Quantity:     2,
			PriceAtOrder: 600_000,
		},
	}

	order, err := svc.CreateOrder(
		context.Background(),
		"user-123",
		items,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAmountTooHigh)
	assert.Empty(t, order.ID)
}
