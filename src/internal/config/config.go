package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	JWTSecret       string
	AppEnv          string
	AuthCallbackURL string
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
		Port:            getEnv("PORT", "8080"),
		JWTSecret:       getEnv("JWT_SECRET_KEY", "defaultsecret"),
		AppEnv:          getEnv("APP_ENV", "local"),
		AuthCallbackURL: getEnv("AUTH_CALLBACK_URL", "http://localhost:8080/auth/callback"),
	}
	return Cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
