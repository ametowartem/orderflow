package kafka

type PaymentEvent struct {
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
	Success   bool   `json:"success"`
}
