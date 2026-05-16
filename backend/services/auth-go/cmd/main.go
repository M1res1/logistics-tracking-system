package main

import (
	"auth-go/internal/config"
	"auth-go/internal/handler"
	"auth-go/internal/model"
	"auth-go/internal/repository"
	"auth-go/internal/service"
	"auth-go/internal/util"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	// Database connection
	db, err := gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate (In production, use real migrations)
	db.AutoMigrate(&model.User{}, &model.Session{})

	// Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
	})

	// Utilities
	jwtUtil := util.NewJwtUtil(cfg.JWTSecret, cfg.AccessExpiration, cfg.RefreshExpiration)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, sessionRepo, jwtUtil, rdb)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)

	// Router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Auth routes
	v1 := r.Group("/api/v1/auth")
	{
		v1.POST("/register", authHandler.Register)
		v1.POST("/login", authHandler.Login)
		v1.POST("/logout", authHandler.Logout)
		v1.POST("/refresh-token", authHandler.RefreshToken)
	}

	// Protected routes
	protected := r.Group("/api/v1/auth")
	// Use local copy of middleware if needed, or import from pkg
	// For now, let's keep it simple.
	// protected.Use(middleware.AuthMiddleware(cfg.JWTSecret, rdb))
	{
		protected.GET("/me", authHandler.Me)
	}

	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
