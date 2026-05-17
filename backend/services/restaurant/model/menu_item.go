package model

import "time"

type MenuItem struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	RestaurantID    uint      `json:"restaurant_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Price           float64   `json:"price"`
	Category        string    `json:"category"`
	IsAvailable     bool      `gorm:"default:true" json:"is_available"`
	PrepTimeMinutes int       `json:"prep_time_minutes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
