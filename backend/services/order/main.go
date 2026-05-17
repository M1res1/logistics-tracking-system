package main

import (
	"log"

	"logistics-tracking-system/pkg/config"
	"logistics-tracking-system/pkg/database"
	"logistics-tracking-system/pkg/middleware"
	"logistics-tracking-system/services/order/handler"
	"logistics-tracking-system/services/order/model"
	"logistics-tracking-system/services/order/repository"
	"logistics-tracking-system/services/order/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := database.AutoMigrate(db, &model.Order{}, &model.OrderItem{}); err != nil {
		log.Fatal(err)
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("invalid REDIS_URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)

	repo := repository.NewOrderRepository(db)
	svc := service.NewOrderService(repo)
	h := handler.NewOrderHandler(svc)

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	api.Use(middleware.Auth(redisClient))
	{
		api.POST("/orders", h.CreateOrder)
		api.GET("/orders/my", h.ListMyOrders)
		api.GET("/orders/:id", h.GetOrder)
		api.POST("/orders/:id/cancel", h.CancelOrder)
		api.PUT("/orders/:id/status", h.UpdateStatus)
	}

	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
