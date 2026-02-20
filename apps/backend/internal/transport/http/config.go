package httptransport

import (
	"os"
	"time"
)

type Config struct {
	Addr        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
}

func LoadConfig() Config {
	ttl := 24 * time.Hour
	return Config{
		Addr:        getEnv("APP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/elevare-auction?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:      ttl,
	}
}

func getEnv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
