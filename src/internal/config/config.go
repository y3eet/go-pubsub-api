package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	JWTSecret   string
	AppEnv      string
	FrontendURL string
}

var Cfg *Config

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	if Cfg != nil {
		return Cfg
	}

	Cfg = &Config{
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET_KEY", "defaultsecret"),
		AppEnv:      getEnv("APP_ENV", "local"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
	return Cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
