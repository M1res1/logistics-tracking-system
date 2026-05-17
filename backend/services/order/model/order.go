package model

import "time"

// OrderStatus represents the lifecycle state of an order.
type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusPreparing OrderStatus = "PREPARING"
	StatusReady     OrderStatus = "READY"
	StatusAssigned  OrderStatus = "ASSIGNED"
	StatusInTransit OrderStatus = "IN_TRANSIT"
	StatusDelivered OrderStatus = "DELIVERED"
	StatusCancelled OrderStatus = "CANCELLED"
)

// transitions defines the valid state machine edges.
var transitions = map[OrderStatus][]OrderStatus{
	StatusPending:   {StatusConfirmed, StatusCancelled},
	StatusConfirmed: {StatusPreparing, StatusCancelled},
	StatusPreparing: {StatusReady},
	StatusReady:     {StatusAssigned},
	StatusAssigned:  {StatusInTransit},
	StatusInTransit: {StatusDelivered},
	// StatusDelivered and StatusCancelled are terminal — no outgoing edges.
}

// CanTransition reports whether transitioning from 'from' to 'to' is allowed.
func CanTransition(from, to OrderStatus) bool {
	allowed, ok := transitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// Order is the aggregate root for an order in the system.
type Order struct {
	ID              uint        `gorm:"primaryKey"         json:"id"`
	CustomerID      uint        `                           json:"customer_id"`
	RestaurantID    uint        `                           json:"restaurant_id"`
	DriverID        *uint       `                           json:"driver_id,omitempty"`
	Status          OrderStatus `gorm:"default:PENDING"    json:"status"`
	Total           float64     `                           json:"total"`
	Tax             float64     `                           json:"tax"`
	DeliveryFee     float64     `                           json:"delivery_fee"`
	DeliveryAddress string      `                           json:"delivery_address"`
	DeliveryLat     float64     `                           json:"delivery_lat"`
	DeliveryLng     float64     `                           json:"delivery_lng"`
	CreatedAt       time.Time   `                           json:"created_at"`
	UpdatedAt       time.Time   `                           json:"updated_at"`
	Items           []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

// OrderItem is a line item belonging to an Order.
type OrderItem struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	OrderID             uint      `                   json:"order_id"`
	MenuItemID          uint      `                   json:"menu_item_id"`
	Quantity            int       `                   json:"quantity"`
	UnitPrice           float64   `                   json:"unit_price"`
	SpecialInstructions string    `                   json:"special_instructions,omitempty"`
	CreatedAt           time.Time `                   json:"created_at"`
}
