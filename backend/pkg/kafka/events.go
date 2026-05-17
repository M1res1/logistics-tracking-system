package kafka

const (
	TopicOrderCreated       = "order.created"
	TopicOrderStatusChanged = "order.status_changed"
)

type OrderCreatedEvent struct {
	OrderID      uint    `json:"order_id"`
	CustomerID   uint    `json:"customer_id"`
	RestaurantID uint    `json:"restaurant_id"`
	Total        float64 `json:"total"`
}

type OrderStatusChangedEvent struct {
	OrderID   uint   `json:"order_id"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}
