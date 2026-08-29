package service

import (
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

func TestOrderService_CreateOrder_Success(t *testing.T) {
	// Arrange
	store := &fakeStore{}
	svc := NewOrderService(store)

	items := []domain.OrderItem{
		{ProductID: "prod-1", Quantity: 2, PriceAtOrder: 50.25},
		{ProductID: "prod-2", Quantity: 1, PriceAtOrder: 10.00},
	}

	// Act
	order, err := svc.CreateOrder("user-123", items)

	// Assert
	require.NoError(t, err, "не ожидали ошибку при валидных данных")
	assert.Equal(t, "user-123", order.UserID)
	assert.Equal(t, 110.50, order.TotalAmount) // 2*50.25 + 1*10.00
	assert.Equal(t, domain.StatusPending, order.Status)
	assert.NotEmpty(t, order.ID, "ID заказа должен был сгенерироваться")
	assert.Len(t, order.Items, 2, "количество элементов заказа должно совпадать")

	// Проверяем, что ID сгенерировались для каждого элемента заказа
	assert.NotEmpty(t, order.Items[0].ID, "ID элемента заказа должен был сгенерироваться")
	assert.NotEmpty(t, order.Items[1].ID, "ID элемента заказа должен был сгенерироваться")
	assert.False(t, order.CreatedAt.IsZero(), "CreatedAt должен был быть установлен")
}

// Тест на корректную обработку пустого списка товаров
func TestOrderService_CreateOrder_EmptyItems(t *testing.T) {
	// Arrange
	store := &fakeStore{}
	svc := NewOrderService(store)

	// Act
	order, err := svc.CreateOrder("user-123", []domain.OrderItem{})

	// Assert
	require.NoError(t, err, "ошибка не ожидалась при пустом списке товаров")
	assert.Equal(t, "user-123", order.UserID)
	assert.Equal(t, 0.0, order.TotalAmount, "сумма должна быть 0 при пустом списке")
	assert.Equal(t, domain.StatusPending, order.Status)
	assert.NotEmpty(t, order.ID)
	assert.Empty(t, order.Items)
}

// Тест на проброс ошибки от хранилища.
// Заменяет старый тест на невалидную сумму, так как в новой реализации
// CreateOrder рассчитывает сумму автоматически на основе элементов и не
// выполняет валидацию на отрицательные значения.
func TestOrderService_CreateOrder_StoreError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("db connection failed")
	store := &fakeStore{
		createFn: func(order domain.Order) error {
			return expectedErr
		},
	}
	svc := NewOrderService(store)

	items := []domain.OrderItem{
		{ProductID: "prod-1", Quantity: 1, PriceAtOrder: 100.0},
	}

	// Act
	order, err := svc.CreateOrder("user-123", items)

	// Assert
	require.Error(t, err, "ожидалась ошибка от хранилища")
	assert.ErrorIs(t, err, expectedErr, "ошибка должна совпадать с ошибкой хранилища")
	assert.Empty(t, order.ID, "заказ не должен был быть создан при ошибке хранилища")
}
