package main

import (
	"log"

	"logistics-tracking-system/pkg/config"
	"logistics-tracking-system/pkg/database"
	"logistics-tracking-system/pkg/middleware"
	"logistics-tracking-system/services/delivery/handler"
	"logistics-tracking-system/services/delivery/model"
	"logistics-tracking-system/services/delivery/repository"
	"logistics-tracking-system/services/delivery/service"

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

	if err := database.AutoMigrate(db, &model.DeliveryAssignment{}, &model.DeliveryLocation{}, &model.DriverStatus{}); err != nil {
		log.Fatal(err)
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("invalid REDIS_URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)

	repo := repository.NewDeliveryRepository(db)
	svc := service.NewDeliveryService(db, repo)
	dh := handler.NewDeliveryHandler(svc)
	th := handler.NewTrackingHandler(svc)

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")

	// Internal route — no auth required
	api.POST("/deliveries/assign", dh.Assign)

	// Auth-protected routes
	protected := api.Group("")
	protected.Use(middleware.Auth(redisClient))
	{
		protected.POST("/deliveries/:id/accept", dh.Accept)
		protected.POST("/deliveries/:id/pickup", dh.Pickup)
		protected.POST("/deliveries/:id/complete", dh.Complete)
		protected.PUT("/deliveries/:id/location", th.UpdateLocation)
		protected.GET("/deliveries/:id/location", th.GetLocation)
		protected.GET("/drivers/available", th.AvailableDrivers)
	}

	if err := r.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
