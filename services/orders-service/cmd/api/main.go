package main

import (
	"log"
	"net/http"

	"github.com/ametowartem/orderflow/orders-service/internal/handler"
	"github.com/ametowartem/orderflow/orders-service/internal/service"
	"github.com/ametowartem/orderflow/orders-service/internal/store"
	"github.com/gin-gonic/gin"
)

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func main() {
	r := gin.Default()

	memStore := store.NewMemoryStore()
	svc := service.NewOrderService(memStore)
	handler := handler.NewOrderHandler(svc)

	r.GET("/health", health)
	r.GET("/orders", handler.GetOrders)
	r.GET("/orders/:id", handler.GetOrderById)
	r.POST("/orders", handler.Create)

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
