package model

import "time"

type Restaurant struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	OwnerID      uint      `json:"owner_id"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	Lat          float64   `json:"lat"`
	Lng          float64   `json:"lng"`
	CuisineTypes string    `json:"cuisine_types"` // comma-separated
	Rating       float64   `json:"rating"`
	IsActive     bool      `json:"is_active"`
	OpeningTime  string    `json:"opening_time"` // "08:00"
	ClosingTime  string    `json:"closing_time"` // "22:00"
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
