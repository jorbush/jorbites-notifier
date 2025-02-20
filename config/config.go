package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port            string
	SMTPHost        string
	SMTPPort        int
	SMTPUser        string
	SMTPPassword    string
	WorkerCount     int
	RetryAttempts   int
	RetryDelayInSec int
}

func GetConfig() *Config {
	return &Config{
		Port:            getEnvOrDefault("PORT", "8080"),
		SMTPHost:        getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:        getEnvAsIntOrDefault("SMTP_PORT", 587),
		SMTPUser:        os.Getenv("SMTP_USER"),
		SMTPPassword:    os.Getenv("SMTP_PASSWORD"),
		WorkerCount:     getEnvAsIntOrDefault("WORKER_COUNT", 3),
		RetryAttempts:   getEnvAsIntOrDefault("RETRY_ATTEMPTS", 3),
		RetryDelayInSec: getEnvAsIntOrDefault("RETRY_DELAY", 60),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
