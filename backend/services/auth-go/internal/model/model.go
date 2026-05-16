package model

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"unique;not null" json:"email"`
	Phone        string    `json:"phone"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	UserType     string    `gorm:"column:user_type;not null" json:"user_type"`
	IsActive     bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Session struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"-"`
	Token      string    `gorm:"not null" json:"token"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	DeviceInfo string    `json:"device_info"`
	CreatedAt  time.Time `json:"created_at"`
}
