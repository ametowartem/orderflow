package grpcserver_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/ametowartem/orderflow/inventory-service/internal/grpcserver"
	"github.com/ametowartem/orderflow/inventory-service/internal/repository/sqlcgen"
	inventoryv1 "github.com/ametowartem/orderflow/proto/inventory/v1"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type testEnv struct {
	client inventoryv1.InventoryServiceClient
	pool   *pgxpool.Pool
}

func setupTestEnv(t *testing.T) testEnv {
	t.Helper()
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

	runMigrations(t, dsn)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	queries := sqlcgen.New(pool)
	server := grpcserver.NewInventoryServer(queries, pool)

	lis := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	inventoryv1.RegisterInventoryServiceServer(grpcServer, server)

	go func() {
		//nolint:errcheck
		_ = grpcServer.Serve(lis)
	}()
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return testEnv{
		client: inventoryv1.NewInventoryServiceClient(conn),
		pool:   pool,
	}
}

func runMigrations(t *testing.T, dsn string) {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(db, "../../db/migrations"))
}

func insertTestProduct(t *testing.T, pool *pgxpool.Pool, id string, stock int32) {
	t.Helper()

	_, err := pool.Exec(context.Background(),
		`INSERT INTO products (id, name, sku, price, stock_quantity) VALUES ($1, $2, $3, $4, $5)`,
		id, "Test Product", fmt.Sprintf("SKU-%s", id[:8]), 9.99, stock,
	)
	require.NoError(t, err)
}

func TestInventoryServer_CheckStock(t *testing.T) {
	env := setupTestEnv(t)
	productID := uuid.NewString()
	insertTestProduct(t, env.pool, productID, 5)

	t.Run("enough stock", func(t *testing.T) {
		resp, err := env.client.CheckStock(context.Background(), &inventoryv1.CheckStockRequest{
			ProductId: productID,
			Quantity:  3,
		})
		require.NoError(t, err)
		require.True(t, resp.Available)
	})

	t.Run("not enough stock", func(t *testing.T) {
		resp, err := env.client.CheckStock(context.Background(), &inventoryv1.CheckStockRequest{
			ProductId: productID,
			Quantity:  10,
		})
		require.NoError(t, err)
		require.False(t, resp.Available)
	})

	t.Run("product does not exist", func(t *testing.T) {
		_, err := env.client.CheckStock(context.Background(), &inventoryv1.CheckStockRequest{
			ProductId: uuid.NewString(),
			Quantity:  1,
		})
		require.Error(t, err)
	})
}

func TestInventoryServer_ReserveStock_Success(t *testing.T) {
	env := setupTestEnv(t)
	productID := uuid.NewString()
	insertTestProduct(t, env.pool, productID, 5)

	resp, err := env.client.ReserveStock(context.Background(), &inventoryv1.ReserveStockRequest{
		ProductId: productID,
		OrderId:   uuid.NewString(),
		Quantity:  3,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.ReservationId)

	var remaining int32
	err = env.pool.QueryRow(context.Background(),
		"SELECT stock_quantity FROM products WHERE id = $1", productID,
	).Scan(&remaining)
	require.NoError(t, err)
	require.Equal(t, int32(2), remaining, "5 - 3 = 2 должно остаться на складе")
}

func TestInventoryServer_ReserveStock_InsufficientStock(t *testing.T) {
	env := setupTestEnv(t)
	productID := uuid.NewString()
	insertTestProduct(t, env.pool, productID, 1)

	_, err := env.client.ReserveStock(context.Background(), &inventoryv1.ReserveStockRequest{
		ProductId: productID,
		OrderId:   uuid.NewString(),
		Quantity:  5,
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok, "ошибка должна быть gRPC-статусом, а не голым error")
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

type reserveResult struct {
	orderID string
	err     error
}

func TestInventoryServer_ReserveStock_ConcurrentRace(t *testing.T) {
	env := setupTestEnv(t)
	productID := uuid.NewString()
	insertTestProduct(t, env.pool, productID, 1)

	const attempts = 2
	results := make(chan reserveResult, attempts)

	var wg sync.WaitGroup
	for range attempts {
		orderID := uuid.NewString()
		wg.Go(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := env.client.ReserveStock(ctx, &inventoryv1.ReserveStockRequest{
				ProductId: productID,
				OrderId:   orderID,
				Quantity:  1,
			})
			results <- reserveResult{orderID: orderID, err: err}
		})
	}

	wg.Wait()
	close(results)

	var successCount, failureCount int
	for res := range results {
		switch {
		case res.err == nil:
			successCount++
		default:
			st, ok := status.FromError(res.err)
			require.True(t, ok)
			require.Equal(t, codes.FailedPrecondition, st.Code(),
				"неожиданный код ошибки для заказа %s: %v", res.orderID, res.err)
			failureCount++
		}
	}

	require.Equal(t, 1, successCount, "ровно один запрос должен быть успешным")
	require.Equal(t, 1, failureCount, "ровно один запрос должен получить отказ по нехватке остатка")

	var remaining int32
	err := env.pool.QueryRow(context.Background(),
		"SELECT stock_quantity FROM products WHERE id = $1", productID,
	).Scan(&remaining)
	require.NoError(t, err)
	require.Equal(t, int32(0), remaining, "остаток не должен уйти в минус и не должен остаться 1")

	var reservationCount int
	err = env.pool.QueryRow(context.Background(),
		"SELECT count(*) FROM stock_reservations WHERE product_id = $1", productID,
	).Scan(&reservationCount)
	require.NoError(t, err)
	require.Equal(t, 1, reservationCount, "должна быть создана ровно одна резервация, не две")
}
