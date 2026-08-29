package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ametowartem/orderflow/orders-service/internal/handler"
	"github.com/ametowartem/orderflow/orders-service/internal/service"
	"github.com/ametowartem/orderflow/orders-service/internal/store/postgres"
	"github.com/gin-gonic/gin"
)

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func main() {
	r := gin.Default()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=postgres user=orders_user password=orders_pass dbname=orders_db port=5432 sslmode=disable"
	}

	// memStore := store.NewMemoryStore()
	pgStore, err := postgres.NewPostgresStore(dsn)
	if err != nil {
		log.Fatal(err)
	}
	svc := service.NewOrderService(pgStore)
	handler := handler.NewOrderHandler(svc)

	r.GET("/health", health)
	r.GET("/orders", handler.GetOrders)
	r.GET("/orders/:id", handler.GetOrderById)
	r.POST("/orders", handler.Create)

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
