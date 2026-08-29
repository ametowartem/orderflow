package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/ametowartem/orderflow/orders-service/internal/domain"
	"github.com/ametowartem/orderflow/orders-service/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type itemRequest struct {
	ProductId    string  `json:"productId" binding:"required,uuid"`
	Quantity     int     `json:"quantity" binding:"required,gt=0"`
	PriceAtOrder float64 `json:"priceAtOrder" binding:"required,gt=0"`
}

type createOrderRequest struct {
	UserID string        `json:"userId" binding:"required,uuid"`
	Items  []itemRequest `json:"items"`
	// Amount float64 `json:"amount" binding:"required,gt=0"`
}

type getOrderByIdRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type orderResponse struct {
	ID          string              `json:"id"`
	UserID      string              `json:"userId"`
	Status      string              `json:"status"`
	TotalAmount float64             `json:"totalAmount"`
	Items       []orderItemResponse `json:"items"`
	CreatedAt   string              `json:"createdAt"`
}

type orderItemResponse struct {
	ID           string  `json:"id"`
	ProductId    string  `json:"productId"`
	Quantity     int     `json:"quantity"`
	PriceAtOrder float64 `json:"priceAtOrder"`
}

func toOrderResponse(o domain.Order) orderResponse {
	orderItems := make([]orderItemResponse, 0, len(o.Items))
	for _, i := range o.Items {
		orderItems = append(orderItems, toOrderItemReponse(i))
	}

	return orderResponse{
		ID:          o.ID,
		UserID:      o.UserID,
		Status:      string(o.Status),
		TotalAmount: o.TotalAmount,
		Items:       orderItems,
		CreatedAt:   o.CreatedAt.Format(time.RFC3339),
	}
}

func toDomainOrderItem(i itemRequest) domain.OrderItem {
	return domain.OrderItem{
		ID:           "",
		ProductID:    i.ProductId,
		Quantity:     i.Quantity,
		PriceAtOrder: i.PriceAtOrder,
	}
}

func toOrderItemReponse(i domain.OrderItem) orderItemResponse {
	return orderItemResponse{
		ID:           i.ID,
		ProductId:    i.ProductID,
		Quantity:     i.Quantity,
		PriceAtOrder: i.PriceAtOrder,
	}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var orderRequest createOrderRequest
	if err := c.ShouldBindJSON(&orderRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]domain.OrderItem, 0, len(orderRequest.Items))
	for _, i := range orderRequest.Items {
		items = append(items, toDomainOrderItem(i))
	}

	order, err := h.svc.CreateOrder(orderRequest.UserID, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, toOrderResponse(order))
}

func (h *OrderHandler) GetOrderById(c *gin.Context) {
	var orderRequest getOrderByIdRequest
	if err := c.ShouldBindUri(&orderRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.svc.GetOrderById(orderRequest.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, toOrderResponse(order))
}

func (h *OrderHandler) GetOrders(c *gin.Context) {

	orders, err := h.svc.GetOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	responseOrders := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		responseOrders = append(responseOrders, toOrderResponse(o))
	}

	c.JSON(http.StatusOK, responseOrders)
}
