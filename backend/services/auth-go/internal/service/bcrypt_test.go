package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHashAndCompare(t *testing.T) {
	password := "securepassword123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	// correct password
	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		t.Errorf("expected valid password to match, got: %v", err)
	}

	// wrong password
	if err := bcrypt.CompareHashAndPassword(hash, []byte("wrongpassword")); err == nil {
		t.Error("expected wrong password to fail")
	}
}

func TestBcryptDifferentPasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{name: "short password", password: "abc"},
		{name: "long password", password: "this-is-a-very-long-password-that-exceeds-normal-length-123456"},
		{name: "special characters", password: "p@$$w0rd!#%^&*()"},
		{name: "unicode", password: "пароль123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := bcrypt.GenerateFromPassword([]byte(tt.password), bcrypt.MinCost)
			if err != nil {
				t.Fatalf("GenerateFromPassword() error = %v", err)
			}

			if err := bcrypt.CompareHashAndPassword(hash, []byte(tt.password)); err != nil {
				t.Errorf("correct password should match hash, got: %v", err)
			}

			if err := bcrypt.CompareHashAndPassword(hash, []byte(tt.password+"x")); err == nil {
				t.Error("modified password should not match hash")
			}
		})
	}
}

func TestBcryptHashesAreUnique(t *testing.T) {
	password := "samepassword"

	hash1, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("first GenerateFromPassword() error = %v", err)
	}

	hash2, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("second GenerateFromPassword() error = %v", err)
	}

	if string(hash1) == string(hash2) {
		t.Error("bcrypt should produce different hashes for the same password due to random salts")
	}

	// Both hashes must still validate the same password
	if err := bcrypt.CompareHashAndPassword(hash1, []byte(password)); err != nil {
		t.Errorf("hash1 should still match original password: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hash2, []byte(password)); err != nil {
		t.Errorf("hash2 should still match original password: %v", err)
	}
}
