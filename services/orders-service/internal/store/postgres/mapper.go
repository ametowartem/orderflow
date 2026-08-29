package postgres

import "github.com/ametowartem/orderflow/orders-service/internal/domain"

func toOrderModel(o domain.Order) OrderModel {
	return OrderModel{
		ID:          o.ID,
		UserID:      o.UserID,
		Status:      string(o.Status),
		TotalAmount: o.TotalAmount,
		CreatedAt:   o.CreatedAt,
	}
}

func toDomainOrder(m OrderModel) domain.Order {
	items := make([]domain.OrderItem, 0, len(m.Items))
	for _, i := range m.Items {
		items = append(items, toDomainOrderItem(i))
	}

	return domain.Order{
		ID:          m.ID,
		UserID:      m.UserID,
		Status:      domain.OrderStatus(m.Status),
		TotalAmount: m.TotalAmount,
		Items:       items,
		CreatedAt:   m.CreatedAt,
	}
}

func toOrderItemModel(i domain.OrderItem, orderID string) OrderItemModel {
	return OrderItemModel{
		ID:           i.ID,
		OrderID:      orderID,
		ProductID:    i.ProductID,
		Quantity:     i.Quantity,
		PriceAtOrder: i.PriceAtOrder,
	}
}

func toDomainOrderItem(i OrderItemModel) domain.OrderItem {
	return domain.OrderItem{
		ID:           i.ID,
		ProductID:    i.ProductID,
		Quantity:     i.Quantity,
		PriceAtOrder: i.PriceAtOrder,
	}
}
