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

	repo := repository.NewDeliveryRepository(db)
	svc := service.NewDeliveryService(db, repo)
	dh := handler.NewDeliveryHandler(svc)

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	{
		api.POST("/deliveries/assign", dh.Assign)
		api.POST("/deliveries/:id/accept", dh.Accept)
		api.POST("/deliveries/:id/pickup", dh.Pickup)
		api.POST("/deliveries/:id/complete", dh.Complete)
	}

	if err := r.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
