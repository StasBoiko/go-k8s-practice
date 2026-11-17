// Package config provides application configuration loaded from environment variables.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration parameters for the application.
type Config struct {
	AppEnv      string
	AppVersion  string
	HTTPPort    string
	DatabaseURL string
}

// Load loads the configuration from environment variables and an optional .env file.
func Load() *Config {
	// Load .env file if it exists, ignore error if it doesn't
	_ = godotenv.Load()

	return &Config{
		AppEnv:      getEnv("APP_ENV", "local"),
		AppVersion:  getEnv("APP_VERSION", "0.1.0"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}
