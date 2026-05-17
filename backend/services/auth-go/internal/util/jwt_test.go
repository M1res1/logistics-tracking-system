package util

import (
	"testing"
	"time"
)

func newTestJwtUtil(accessExpMs, refreshExpMs int64) *JwtUtil {
	return NewJwtUtil("test-secret-key", accessExpMs, refreshExpMs)
}

func TestGenerateAndVerifyToken(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		userType string
		userID   uint
	}{
		{
			name:     "customer token",
			email:    "customer@example.com",
			userType: "CUSTOMER",
			userID:   1,
		},
		{
			name:     "driver token",
			email:    "driver@example.com",
			userType: "DRIVER",
			userID:   42,
		},
		{
			name:     "admin token",
			email:    "admin@example.com",
			userType: "ADMIN",
			userID:   100,
		},
	}

	jwtUtil := newTestJwtUtil(3600000, 86400000) // 1h access, 24h refresh

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtUtil.GenerateToken(tt.email, tt.userType, tt.userID)
			if err != nil {
				t.Fatalf("GenerateToken() error = %v", err)
			}
			if token == "" {
				t.Fatal("GenerateToken() returned empty token")
			}

			got, err := jwtUtil.ExtractUsername(token)
			if err != nil {
				t.Fatalf("ExtractUsername() error = %v", err)
			}
			if got != tt.email {
				t.Errorf("ExtractUsername() = %q, want %q", got, tt.email)
			}
		})
	}
}

func TestTokenExpiry(t *testing.T) {
	// Use 1ms expiry so the token expires almost immediately
	jwtUtil := newTestJwtUtil(1, 1)

	token, err := jwtUtil.GenerateToken("user@example.com", "CUSTOMER", 1)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Sleep to let the token expire (JWT exp is Unix seconds, so we need to
	// cross a second boundary; sleep 1100ms to be safe)
	time.Sleep(1100 * time.Millisecond)

	_, err = jwtUtil.ExtractUsername(token)
	if err == nil {
		t.Error("ExtractUsername() expected error for expired token, got nil")
	}
}

func TestRefreshToken(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "standard email",
			email: "refresh@example.com",
		},
		{
			name:  "email with plus",
			email: "user+tag@example.com",
		},
	}

	jwtUtil := newTestJwtUtil(3600000, 86400000)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtUtil.GenerateRefreshToken(tt.email)
			if err != nil {
				t.Fatalf("GenerateRefreshToken() error = %v", err)
			}
			if token == "" {
				t.Fatal("GenerateRefreshToken() returned empty token")
			}

			got, err := jwtUtil.ExtractUsername(token)
			if err != nil {
				t.Fatalf("ExtractUsername() error = %v", err)
			}
			if got != tt.email {
				t.Errorf("ExtractUsername() = %q, want %q", got, tt.email)
			}
		})
	}
}

func TestInvalidToken(t *testing.T) {
	jwtUtil := newTestJwtUtil(3600000, 86400000)

	tests := []struct {
		name        string
		tokenString string
	}{
		{
			name:        "empty string",
			tokenString: "",
		},
		{
			name:        "random garbage",
			tokenString: "not.a.token",
		},
		{
			name:        "truncated token",
			tokenString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyQGV4YW1wbGUuY29tIn0",
		},
		{
			name:        "wrong signature",
			tokenString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyQGV4YW1wbGUuY29tIiwiZXhwIjo5OTk5OTk5OTk5fQ.wrongsignature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jwtUtil.ExtractUsername(tt.tokenString)
			if err == nil {
				t.Errorf("ExtractUsername(%q) expected error, got nil", tt.tokenString)
			}
		})
	}
}
