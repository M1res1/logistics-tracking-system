package main

import (
    "log"

    "logistics-tracking-system/pkg/config"
    "logistics-tracking-system/pkg/database"
    "logistics-tracking-system/pkg/middleware"
    "logistics-tracking-system/services/payment/handler"
    "logistics-tracking-system/services/payment/repository"
    "logistics-tracking-system/services/payment/service"

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

    opt, err := redis.ParseURL(cfg.RedisURL)
    if err != nil {
        log.Fatal(err)
    }
    redisClient := redis.NewClient(opt)

    repo := repository.NewPaymentRepository(db)
    svc := service.NewPaymentService(db, repo, redisClient)
    h := handler.NewPaymentHandler(svc)

    r := gin.New()
    r.Use(middleware.CORS())
    r.Use(middleware.Logger())
    r.Use(gin.Recovery())

    api := r.Group("/api/v1")
    {
        api.POST("/payments/process", h.ProcessPayment)
        api.GET("/payments/:id", h.GetPayment)
        api.POST("/payments/:id/refund", h.RefundPayment)

        api.GET("/wallet/:userId", h.GetWallet)
        api.POST("/wallet/:userId/topup", h.TopupWallet)
    }

    if err := r.Run(":8084"); err != nil {
        log.Fatal(err)
    }
}