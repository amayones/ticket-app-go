package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}
	RequireEnv("JWT_SECRET")
	RequireEnv("DB_HOST")
	RequireEnv("DB_PORT")
	RequireEnv("DB_DATABASE")
	RequireEnv("DB_USERNAME")
	RequireEnv("DB_PASSWORD")
}

func GetEnv(key string) string {
	return os.Getenv(key)
}

func GetEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func RequireEnv(key string) {
	if os.Getenv(key) == "" {
		log.Fatalf("Missing required env: %s", key)
	}
}
