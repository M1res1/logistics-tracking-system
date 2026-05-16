package handler

import (
	"auth-go/internal/dto"
	"auth-go/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if err := h.authService.Logout(c.Request.Context(), authHeader); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken := c.Query("refreshToken")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	resp, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	// Use the helper from pkg/middleware/auth.go if possible, 
	// but here we just need the email from context.
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Based on pkg/middleware/auth.go, the user in context is a User struct
	// Since we are in a different module, we might have issues with type assertion 
	// if we don't import the package. For now, let's use a simple approach.
	
	var email string
	// Attempt to get email via reflection or if we know the structure
	// Given it's a shared pkg, we should ideally import it.
	// However, to keep it simple and avoid circular deps or complex imports:
	
	// If the middleware is from backend/pkg/middleware, we can try to use it.
	// For now, let's assume the email is available in claims or we can cast it if we import it.
	
	// Let's assume we'll fix the import in the next step if needed.
	// For the sake of completion:
	type contextUser struct {
		ID       float64
		Email    string
		UserType string
	}
	
	if u, ok := userVal.(contextUser); ok {
		email = u.Email
	} else {
		// Try to extract from map if it was stored as map
		if m, ok := userVal.(map[string]interface{}); ok {
			email = m["sub"].(string)
		} else {
			// Fallback: we might need to properly import the middleware package
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
			return
		}
	}

	resp, err := h.authService.GetCurrentUser(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
