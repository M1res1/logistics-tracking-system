package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret string

	KafkaBroker string

	AuthServicePort       string
	OrderServicePort      string
	DeliveryServicePort   string
	RestaurantServicePort string
	PaymentServicePort    string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "admin"),
		DBPassword: getEnv("DB_PASSWORD", "secret"),
		DBName:     getEnv("DB_NAME", "fooddelivery"),

		JWTSecret:   getEnv("JWT_SECRET", "change-me"),
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),

		AuthServicePort:       getEnv("AUTH_SERVICE_PORT", ":8080"),
		OrderServicePort:      getEnv("ORDER_SERVICE_PORT", ":8081"),
		DeliveryServicePort:   getEnv("DELIVERY_SERVICE_PORT", ":8082"),
		RestaurantServicePort: getEnv("RESTAURANT_SERVICE_PORT", ":8083"),
		PaymentServicePort:    getEnv("PAYMENT_SERVICE_PORT", ":8084"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
