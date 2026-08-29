package domain

import (
	"time"
)

type Order struct {
	ID          string
	UserID      string
	Status      OrderStatus
	TotalAmount float64
	Items       []OrderItem
	CreatedAt   time.Time
}

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ID           string
	ProductID    string
	Quantity     int
	PriceAtOrder float64
}
