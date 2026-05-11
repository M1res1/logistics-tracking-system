package model

import "time"

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusConfirmed Status = "CONFIRMED"
	StatusPreparing Status = "PREPARING"
	StatusReady     Status = "READY"
	StatusAssigned  Status = "ASSIGNED"
	StatusInTransit Status = "IN_TRANSIT"
	StatusDelivered Status = "DELIVERED"
	StatusCancelled Status = "CANCELLED"
)

// maps each status to where it's allowed to go next
var allowedTransitions = map[Status][]Status{
	StatusPending:   {StatusConfirmed, StatusCancelled},
	StatusConfirmed: {StatusPreparing, StatusCancelled},
	StatusPreparing: {StatusReady},
	StatusReady:     {StatusAssigned},
	StatusAssigned:  {StatusInTransit},
	StatusInTransit: {StatusDelivered},
	StatusDelivered: {},
	StatusCancelled: {},
}

func CanTransition(from, to Status) bool {
	targets, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

type Order struct {
	ID              uint        `gorm:"primaryKey" json:"id"`
	CustomerID      uint        `gorm:"not null;index" json:"customer_id"`
	RestaurantID    uint        `gorm:"not null" json:"restaurant_id"`
	DriverID        *uint       `json:"driver_id,omitempty"`
	Status          Status      `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	Subtotal        float64     `json:"subtotal"`
	Tax             float64     `json:"tax"`
	DeliveryFee     float64     `json:"delivery_fee"`
	Total           float64     `json:"total"`
	DeliveryAddress string      `gorm:"not null" json:"delivery_address"`
	DeliveryLat     float64     `json:"delivery_lat"`
	DeliveryLng     float64     `json:"delivery_lng"`
	Items           []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID                  uint    `gorm:"primaryKey" json:"id"`
	OrderID             uint    `gorm:"not null;index" json:"order_id"`
	MenuItemID          uint    `json:"menu_item_id"`
	Quantity            int     `json:"quantity"`
	UnitPrice           float64 `json:"unit_price"`
	SpecialInstructions string  `json:"special_instructions,omitempty"`
}
