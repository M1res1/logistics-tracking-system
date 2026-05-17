package dto

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone" binding:"omitempty"`
	UserType string `json:"user_type" binding:"required,oneof=CUSTOMER DRIVER ADMIN RESTAURANT"`
}

type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	DeviceInfo string `json:"device_info"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserProfile struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	UserType string `json:"user_type"`
}
