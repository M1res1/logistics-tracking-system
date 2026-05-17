package main

import (
	"log"

	"logistics-tracking-system/pkg/config"
	"logistics-tracking-system/pkg/database"
	"logistics-tracking-system/pkg/middleware"
	"logistics-tracking-system/services/restaurant/handler"
	"logistics-tracking-system/services/restaurant/repository"
	"logistics-tracking-system/services/restaurant/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, _ := config.LoadConfig()

	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("invalid REDIS_URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)

	repo := repository.NewRestaurantRepository(db)
	svc := service.NewRestaurantService(repo)
	rh := handler.NewRestaurantHandler(svc, db)
	mh := handler.NewMenuHandler(svc)

	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	api.Use(middleware.Auth(redisClient))
	{
		api.POST("/restaurants", rh.Create)
		api.GET("/restaurants", rh.List)
		api.GET("/restaurants/:id", rh.Get)
		api.PUT("/restaurants/:id", rh.Update)
		api.PUT("/restaurants/:id/toggle", rh.Toggle)

		api.GET("/restaurants/:id/orders", rh.ListOrders)
		api.POST("/restaurants/:id/orders/:orderId/accept", rh.AcceptOrder)
		api.POST("/restaurants/:id/orders/:orderId/ready", rh.ReadyOrder)
		api.POST("/restaurants/:id/orders/:orderId/reject", rh.RejectOrder)

		api.GET("/restaurants/:id/menu", mh.GetMenu)
		api.POST("/restaurants/:id/menu-items", mh.AddItem)
		api.PUT("/restaurants/:id/menu-items/:itemId", mh.UpdateItem)
		api.DELETE("/restaurants/:id/menu-items/:itemId", mh.DeleteItem)
	}

	if err := r.Run(":8083"); err != nil {
		log.Fatal(err)
	}
}
