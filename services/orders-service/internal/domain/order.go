package domain

import (
	"time"
)

type Order struct {
	ID          string
	UserID      string
	Status      OrderStatus
	TotalAmount float64
	CreatedAt   time.Time
}

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusCancelled OrderStatus = "cancelled"
)
