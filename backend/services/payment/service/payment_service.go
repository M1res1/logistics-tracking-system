package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"logistics-tracking-system/services/payment/model"
	"logistics-tracking-system/services/payment/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PaymentService struct {
	db    *gorm.DB
	repo  *repository.PaymentRepository
	redis *redis.Client
}

func NewPaymentService(db *gorm.DB, repo *repository.PaymentRepository, redis *redis.Client) *PaymentService {
	return &PaymentService{db: db, repo: repo, redis: redis}
}

type ProcessPaymentRequest struct {
	OrderID        uint
	UserID         uint
	Amount         float64
	Method         string
	IdempotencyKey string
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*model.Payment, error) {
	cacheKey := "idempotency:" + req.IdempotencyKey

	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		var p model.Payment
		if err := json.Unmarshal([]byte(cached), &p); err == nil {
			return &p, nil
		}
	}

	if existing, err := s.repo.GetPaymentByIdempotencyKey(req.IdempotencyKey); err == nil {
		return existing, nil
	}

	var payment *model.Payment
	err := s.db.Transaction(func(tx *gorm.DB) error {
		localRepo := repository.NewPaymentRepository(tx)

		p := &model.Payment{
			OrderID:        req.OrderID,
			UserID:         req.UserID,
			Amount:         req.Amount,
			Method:         req.Method,
			Status:         "PENDING",
			IdempotencyKey: req.IdempotencyKey,
		}

		if err := localRepo.CreatePayment(p); err != nil {
			if existing, fetchErr := localRepo.GetPaymentByIdempotencyKey(req.IdempotencyKey); fetchErr == nil {
				payment = existing
				return nil
			}
			return err
		}

		// Mock gateway: always succeed for now
		p.Status = "SUCCEEDED"
		p.GatewayTxID = fmt.Sprintf("mock_tx_%d", time.Now().UnixNano())

		if req.Method == "WALLET" {
			w, err := localRepo.GetOrCreateWalletForUpdate(req.UserID)
			if err != nil {
				return err
			}

			if w.Balance < req.Amount {
				return errors.New("insufficient wallet balance")
			}

			before := w.Balance
			after := w.Balance - req.Amount

			if err := localRepo.UpdateWalletBalance(w.ID, after); err != nil {
				return err
			}

			txn := &model.WalletTransaction{
				WalletID:      w.ID,
				Type:          "DEBIT",
				Amount:        req.Amount,
				BalanceBefore: before,
				BalanceAfter:  after,
			}
			if err := localRepo.CreateWalletTransaction(txn); err != nil {
				return err
			}
		}

		if err := localRepo.UpdatePayment(p); err != nil {
			return err
		}

		payment = p
		return nil
	})
	if err != nil {
		return nil, err
	}

	b, _ := json.Marshal(payment)
	_ = s.redis.Set(ctx, cacheKey, string(b), 24*time.Hour).Err()

	return payment, nil
}

func (s *PaymentService) GetPayment(id uint) (*model.Payment, error) {
	return s.repo.GetPaymentByID(id)
}

func (s *PaymentService) RefundPayment(ctx context.Context, paymentID uint, amount float64, reason string) (*model.Refund, error) {
	var refund *model.Refund

	err := s.db.Transaction(func(tx *gorm.DB) error {
		localRepo := repository.NewPaymentRepository(tx)

		p, err := localRepo.GetPaymentByID(paymentID)
		if err != nil {
			return err
		}

		if p.Status != "SUCCEEDED" {
			return errors.New("payment not successful, cannot refund")
		}

		r := &model.Refund{
			PaymentID: paymentID,
			Amount:    amount,
			Reason:    reason,
			Status:    "REFUNDED",
		}
		if err := localRepo.CreateRefund(r); err != nil {
			return err
		}

		w, err := localRepo.GetOrCreateWalletForUpdate(p.UserID)
		if err != nil {
			return err
		}

		before := w.Balance
		after := w.Balance + amount

		if err := localRepo.UpdateWalletBalance(w.ID, after); err != nil {
			return err
		}

		txn := &model.WalletTransaction{
			WalletID:      w.ID,
			Type:          "CREDIT",
			Amount:        amount,
			BalanceBefore: before,
			BalanceAfter:  after,
		}
		if err := localRepo.CreateWalletTransaction(txn); err != nil {
			return err
		}

		refund = r
		return nil
	})
	if err != nil {
		return nil, err
	}

	return refund, nil
}

func (s *PaymentService) GetWallet(userID uint) (*model.Wallet, error) {
	return s.repo.GetOrCreateWallet(userID)
}

func (s *PaymentService) TopupWallet(userID uint, amount float64) (*model.Wallet, error) {
	var wallet *model.Wallet

	err := s.db.Transaction(func(tx *gorm.DB) error {
		localRepo := repository.NewPaymentRepository(tx)

		w, err := localRepo.GetOrCreateWalletForUpdate(userID)
		if err != nil {
			return err
		}

		before := w.Balance
		after := w.Balance + amount

		if err := localRepo.UpdateWalletBalance(w.ID, after); err != nil {
			return err
		}

		txn := &model.WalletTransaction{
			WalletID:      w.ID,
			Type:          "CREDIT",
			Amount:        amount,
			BalanceBefore: before,
			BalanceAfter:  after,
		}
		if err := localRepo.CreateWalletTransaction(txn); err != nil {
			return err
		}

		w.Balance = after
		wallet = w
		return nil
	})
	if err != nil {
		return nil, err
	}

	return wallet, nil
}
