package config

import (
	"os"
)

type Config struct {
	DBURL             string
	RedisHost         string
	RedisPort         string
	JWTSecret         string
	AccessExpiration  int64 // milliseconds
	RefreshExpiration int64 // milliseconds
	Port              string
}

func Load() *Config {
	return &Config{
		DBURL:             getEnv("DB_URL", "host=localhost user=postgres password=postgres dbname=postgres port=5435 sslmode=disable"),
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         getEnv("REDIS_PORT", "6379"),
		JWTSecret:         getEnv("JWT_SECRET", "404E635266556A586E3272357538782F413F4428472B4B6250645367566B5970"),
		AccessExpiration:  3600000,  // 1 hour
		RefreshExpiration: 86400000, // 24 hours
		Port:              getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
