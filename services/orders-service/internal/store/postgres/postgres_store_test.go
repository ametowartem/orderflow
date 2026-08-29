package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/ametowartem/orderflow/orders-service/internal/domain"
	"github.com/ametowartem/orderflow/orders-service/internal/store/postgres"
)

func setupTestStore(t *testing.T) *postgres.PostgresStore {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("test_db"),
		tcpostgres.WithUsername("test_user"),
		tcpostgres.WithPassword("test_pass"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	store, err := postgres.NewPostgresStore(dsn)
	require.NoError(t, err)

	return store
}

func TestPostgresStore_CreateAndGetByID(t *testing.T) {
	store := setupTestStore(t)

	order := domain.Order{
		ID:     "11111111-1111-1111-1111-111111111111",
		UserID: "22222222-2222-2222-2222-222222222222",
		Status: domain.StatusPending,
		Items: []domain.OrderItem{
			{ID: "33333333-3333-3333-3333-333333333333", ProductID: "p1", Quantity: 2, PriceAtOrder: 100},
		},
		TotalAmount: 200,
	}

	err := store.Create(order)
	require.NoError(t, err)

	got, err := store.GetByID(order.ID)
	require.NoError(t, err)
	require.Equal(t, order.UserID, got.UserID)
	require.Equal(t, order.TotalAmount, got.TotalAmount)

	require.Len(t, got.Items, 1)
	require.Equal(t, order.Items[0].ProductID, got.Items[0].ProductID)
	require.Equal(t, order.Items[0].Quantity, got.Items[0].Quantity)
}

func TestPostgresStore_GetByID_NotFound(t *testing.T) {
	store := setupTestStore(t)

	_, err := store.GetByID("a137de03-80f7-452f-9560-62fa7d70ed70")
	require.ErrorIs(t, err, domain.ErrNotFound)
}
