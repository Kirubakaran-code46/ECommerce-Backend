package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	FrontendURL string
	Env         string
}

func Load() *Config {
	// Load .env if it exists; ignore error if not found in production
	if err := godotenv.Load(); err != nil {
		// Also try loading from root .env if running from backend folder
		_ = godotenv.Load("../../.env")
		_ = godotenv.Load("../.env")
	}

	port := getEnv("PORT")
	dbURL := getEnv("DATABASE_URL")
	frontendURL := getEnv("FRONTEND_URL")
	env := getEnv("ENV")

	log.Printf("[Config] Loaded configuration: PORT=%s, ENV=%s, FRONTEND_URL=%s", port, env, frontendURL)

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		FrontendURL: frontendURL,
		Env:         env,
	}
}

func getEnv(key string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return ""
}
