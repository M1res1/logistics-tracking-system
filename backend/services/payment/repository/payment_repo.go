package repository

import (
	"errors"

	"logistics-tracking-system/services/payment/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) CreatePayment(p *model.Payment) error {
	return r.db.Create(p).Error
}

func (r *PaymentRepository) UpdatePayment(p *model.Payment) error {
	return r.db.Save(p).Error
}

func (r *PaymentRepository) GetPaymentByID(id uint) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) GetPaymentByIdempotencyKey(key string) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.Where("idempotency_key = ?", key).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) GetOrCreateWallet(userID uint) (*model.Wallet, error) {
	var w model.Wallet
	err := r.db.Where("user_id = ?", userID).First(&w).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		w = model.Wallet{UserID: userID, Balance: 0}
		if err := r.db.Create(&w).Error; err != nil {
			return nil, err
		}
		return &w, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *PaymentRepository) GetOrCreateWalletForUpdate(userID uint) (*model.Wallet, error) {
	var w model.Wallet
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&w).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		w = model.Wallet{UserID: userID, Balance: 0}
		if err := r.db.Create(&w).Error; err != nil {
			return nil, err
		}
		return &w, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *PaymentRepository) UpdateWalletBalance(walletID uint, newBalance float64) error {
	return r.db.Model(&model.Wallet{}).Where("id = ?", walletID).Update("balance", newBalance).Error
}

func (r *PaymentRepository) CreateWalletTransaction(t *model.WalletTransaction) error {
	return r.db.Create(t).Error
}

func (r *PaymentRepository) CreateRefund(ref *model.Refund) error {
	return r.db.Create(ref).Error
}
