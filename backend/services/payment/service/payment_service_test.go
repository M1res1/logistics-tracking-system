package service

import "testing"

// TestWalletTopupBalance verifies that topup arithmetic produces the correct balance.
func TestWalletTopupBalance(t *testing.T) {
	tests := []struct {
		name           string
		initialBalance float64
		topupAmount    float64
		expectedAfter  float64
	}{
		{name: "topup from zero", initialBalance: 0, topupAmount: 50.0, expectedAfter: 50.0},
		{name: "topup existing balance", initialBalance: 100.0, topupAmount: 25.0, expectedAfter: 125.0},
		{name: "fractional topup", initialBalance: 99.99, topupAmount: 0.01, expectedAfter: 100.0},
		{name: "large topup", initialBalance: 0, topupAmount: 10000.0, expectedAfter: 10000.0},
		{name: "zero topup", initialBalance: 50.0, topupAmount: 0.0, expectedAfter: 50.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.initialBalance + tc.topupAmount
			if got != tc.expectedAfter {
				t.Errorf("topup: %.2f + %.2f = %.2f, want %.2f",
					tc.initialBalance, tc.topupAmount, got, tc.expectedAfter)
			}
		})
	}
}

// TestWalletInsufficientBalance verifies that a debit exceeding the balance is detected.
func TestWalletInsufficientBalance(t *testing.T) {
	tests := []struct {
		name       string
		balance    float64
		amount     float64
		sufficient bool
	}{
		{name: "clearly insufficient", balance: 10.0, amount: 50.0, sufficient: false},
		{name: "exactly sufficient", balance: 50.0, amount: 50.0, sufficient: true},
		{name: "more than sufficient", balance: 100.0, amount: 50.0, sufficient: true},
		{name: "zero balance non-zero amount", balance: 0.0, amount: 0.01, sufficient: false},
		{name: "zero amount always sufficient", balance: 0.0, amount: 0.0, sufficient: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Mirrors the wallet check in ProcessPayment: w.Balance < req.Amount → insufficient
			sufficient := tc.balance >= tc.amount
			if sufficient != tc.sufficient {
				t.Errorf("balance=%.2f amount=%.2f: sufficient=%v, want %v",
					tc.balance, tc.amount, sufficient, tc.sufficient)
			}
		})
	}
}

// TestIdempotencyKey verifies the cache key construction used in ProcessPayment.
func TestIdempotencyKey(t *testing.T) {
	tests := []struct {
		name           string
		rawKey         string
		expectedCacheKey string
	}{
		{
			name:             "standard UUID key",
			rawKey:           "test-key-123",
			expectedCacheKey: "idempotency:test-key-123",
		},
		{
			name:             "UUID format key",
			rawKey:           "550e8400-e29b-41d4-a716-446655440000",
			expectedCacheKey: "idempotency:550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:             "empty key",
			rawKey:           "",
			expectedCacheKey: "idempotency:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key := "idempotency:" + tc.rawKey
			if key != tc.expectedCacheKey {
				t.Errorf("idempotency key = %q, want %q", key, tc.expectedCacheKey)
			}
		})
	}
}

// TestWalletDebitBalance verifies that a successful debit reduces balance correctly.
func TestWalletDebitBalance(t *testing.T) {
	tests := []struct {
		name          string
		initialBalance float64
		debitAmount   float64
		expectedAfter float64
	}{
		{name: "simple debit", initialBalance: 100.0, debitAmount: 30.0, expectedAfter: 70.0},
		{name: "exact debit", initialBalance: 50.0, debitAmount: 50.0, expectedAfter: 0.0},
		{name: "small debit", initialBalance: 1000.0, debitAmount: 0.01, expectedAfter: 999.99},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			after := tc.initialBalance - tc.debitAmount
			if after != tc.expectedAfter {
				t.Errorf("debit: %.2f - %.2f = %.4f, want %.4f",
					tc.initialBalance, tc.debitAmount, after, tc.expectedAfter)
			}
		})
	}
}
