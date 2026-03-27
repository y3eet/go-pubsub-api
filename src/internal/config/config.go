package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	JWTSecret         string
	AppEnv            string
	GoPubSubMasterKey string
	AuthCallbackURL   string
	AllowedOrigins    []string
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
		Port:              getEnv("PORT", "8080"),
		JWTSecret:         getEnv("JWT_SECRET_KEY", "defaultsecret"),
		AppEnv:            getEnv("APP_ENV", "local"),
		GoPubSubMasterKey: getEnv("GO_PUB_SUB_MASTER_KEY", "defaultmasterkey"),
		AuthCallbackURL:   getEnv("AUTH_CALLBACK_URL", "http://localhost:8080/auth/callback"),
		AllowedOrigins:    getEnvArray("ALLOWED_ORIGINS"),
	}
	return Cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvArray(key string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return []string{}
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	return result
}
