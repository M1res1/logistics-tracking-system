package config

import "os"

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	DBSSLMode  string
	RedisURL   string
	Port       string
	JWTSecret  string
}

func LoadConfig() (*Config, error) {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "logistics"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		RedisURL:   getEnv("REDIS_URL", "redis://localhost:6379"),
		Port:       getEnv("PORT", "8082"),
		JWTSecret:  getEnv("JWT_SECRET", "supersecretkey"),
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
