package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// User holds the authenticated user's identity extracted from the JWT.
type User struct {
	ID       uint
	Email    string
	UserType string
}

type contextKey string

const userContextKey contextKey = "user"

// Auth returns a gin middleware that validates a Bearer JWT token and optionally
// checks the Redis blacklist before attaching the parsed User to the context.
func Auth(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "supersecretkey"
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		// Check Redis blacklist
		blacklistKey := "blacklist:" + tokenStr
		val, err := redisClient.Get(context.Background(), blacklistKey).Result()
		if err == nil && val != "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}

		var userID uint
		if raw, ok := claims["user_id"]; ok {
			if f, ok := raw.(float64); ok {
				userID = uint(f)
			}
		}

		email, _ := claims["sub"].(string)
		userType, _ := claims["user_type"].(string)

		user := User{
			ID:       userID,
			Email:    email,
			UserType: userType,
		}

		c.Set("user", user)
		c.Next()
	}
}

// RequireRole returns a middleware that aborts with 403 if the authenticated
// user's UserType is not in the provided roles list.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		user, ok := raw.(User)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		for _, role := range roles {
			if user.UserType == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

// GetUserFromCtx extracts the User attached by the Auth middleware from a gin context.
func GetUserFromCtx(c *gin.Context) (User, bool) {
	raw, exists := c.Get("user")
	if !exists {
		return User{}, false
	}
	user, ok := raw.(User)
	return user, ok
}
