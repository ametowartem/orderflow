package postgres

import "time"

type OrderModel struct {
	ID          string           `gorm:"primaryKey;type:uuid"`
	UserID      string           `gorm:"index;not null;type:uuid"`
	Status      string           `gorm:"not null"`
	TotalAmount float64          `gorm:"not null"`
	Items       []OrderItemModel `gorm:"foreignKey:OrderID"`
	CreatedAt   time.Time
}

func (OrderModel) TableName() string { return "orders" }

type OrderItemModel struct {
	ID           string  `gorm:"primaryKey;type:uuid"`
	OrderID      string  `gorm:"index;not null"`
	ProductID    string  `gorm:"not null"`
	Quantity     int     `gorm:"not null"`
	PriceAtOrder float64 `gorm:"not null"`
}

func (OrderItemModel) TableName() string { return "order_items" }
