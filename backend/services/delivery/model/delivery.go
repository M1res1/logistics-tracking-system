package model

import "time"

type DeliveryAssignment struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	OrderID      uint       `json:"order_id"`
	DriverID     uint       `json:"driver_id"`
	Status       string     `json:"status"`
	PickupTime   *time.Time `json:"pickup_time,omitempty"`
	DeliveryTime *time.Time `json:"delivery_time,omitempty"`
	DistanceKm   float64    `json:"distance_km"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type DeliveryLocation struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DeliveryID uint      `json:"delivery_id"`
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	CreatedAt  time.Time `json:"created_at"`
}

type DriverStatus struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DriverID    uint      `gorm:"uniqueIndex" json:"driver_id"`
	IsOnline    bool      `json:"is_online"`
	IsAvailable bool      `json:"is_available"`
	LastLat     float64   `json:"last_lat"`
	LastLng     float64   `json:"last_lng"`
	UpdatedAt   time.Time `json:"updated_at"`
}
