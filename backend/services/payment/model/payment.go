package model

import "time"

type Payment struct {
	ID             uint `gorm:"primaryKey"`
	OrderID        uint
	UserID         uint
	Amount         float64
	Method         string
	Status         string
	IdempotencyKey string `gorm:"uniqueIndex"`
	GatewayTxID    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Wallet struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"uniqueIndex"`
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WalletTransaction struct {
	ID            uint `gorm:"primaryKey"`
	WalletID      uint
	Type          string
	Amount        float64
	BalanceBefore float64
	BalanceAfter  float64
	CreatedAt     time.Time
}

type Refund struct {
	ID        uint `gorm:"primaryKey"`
	PaymentID uint
	Amount    float64
	Reason    string
	Status    string
	CreatedAt time.Time
}
