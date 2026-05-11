package main

import (
	"log"

	"food-delivery/pkg/config"
	"food-delivery/pkg/database"
	"food-delivery/pkg/kafka"
	"food-delivery/pkg/middleware"
	"food-delivery/services/order/handler"
	"food-delivery/services/order/model"
	"food-delivery/services/order/repository"
	"food-delivery/services/order/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)

	// auto-migrate so we don't have to run sql files manually every time
	if err := db.AutoMigrate(&model.Order{}, &model.OrderItem{}); err != nil {
		log.Fatal("migrate failed:", err)
	}

	orderProducer := kafka.NewProducer(cfg.KafkaBroker, "order.created")
	statusProducer := kafka.NewProducer(cfg.KafkaBroker, "order.status_changed")
	defer orderProducer.Close()
	defer statusProducer.Close()

	repo := repository.NewOrderRepo(db)
	svc := service.NewOrderService(repo, orderProducer, statusProducer)
	h := handler.NewOrderHandler(svc)

	r := gin.Default()
	r.Use(middleware.Logger())
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api/v1")
	api.Use(middleware.RequireAuth(cfg.JWTSecret))
	{
		api.POST("/orders", h.CreateOrder)
		api.GET("/orders/my", h.ListMyOrders) // must be before /:id
		api.GET("/orders/:id", h.GetOrder)
		api.POST("/orders/:id/cancel", h.CancelOrder)
		api.PUT("/orders/:id/status", h.UpdateStatus)
	}

	log.Println("order service starting on", cfg.OrderServicePort)
	if err := r.Run(cfg.OrderServicePort); err != nil {
		log.Fatal(err)
	}
}
