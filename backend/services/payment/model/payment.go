package model

import "time"

type Payment struct {
	ID             uint      `gorm:"primaryKey"       json:"id"`
	OrderID        uint      `                        json:"order_id"`
	UserID         uint      `                        json:"user_id"`
	Amount         float64   `                        json:"amount"`
	Method         string    `                        json:"method"`
	Status         string    `                        json:"status"`
	IdempotencyKey string    `gorm:"uniqueIndex"      json:"idempotency_key"`
	GatewayTxID    string    `                        json:"gateway_tx_id"`
	CreatedAt      time.Time `                        json:"created_at"`
	UpdatedAt      time.Time `                        json:"updated_at"`
}

type Wallet struct {
	ID        uint      `gorm:"primaryKey"  json:"id"`
	UserID    uint      `gorm:"uniqueIndex" json:"user_id"`
	Balance   float64   `                   json:"balance"`
	CreatedAt time.Time `                   json:"created_at"`
	UpdatedAt time.Time `                   json:"updated_at"`
}

type WalletTransaction struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	WalletID      uint      `                  json:"wallet_id"`
	Type          string    `                  json:"type"`
	Amount        float64   `                  json:"amount"`
	BalanceBefore float64   `                  json:"balance_before"`
	BalanceAfter  float64   `                  json:"balance_after"`
	CreatedAt     time.Time `                  json:"created_at"`
}

type Refund struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PaymentID uint      `                  json:"payment_id"`
	Amount    float64   `                  json:"amount"`
	Reason    string    `                  json:"reason"`
	Status    string    `                  json:"status"`
	CreatedAt time.Time `                  json:"created_at"`
}
