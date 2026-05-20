package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	BaseURL  string
	DBURL    string
	RedisURL string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}
	return &Config{
		Port:     getEnv("PORT", "8080"),
		BaseURL:  getEnv("BASE_URL", "http://localhost:8080"),
		DBURL:    mustEnv("DB_URL"),
		RedisURL: mustEnv("REDIS_URL"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}
