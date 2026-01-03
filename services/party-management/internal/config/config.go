package config

import (
	"os"
)

type Config struct {
	RabbitMQURL string
	PostgresURL string
	HTTPPort    string
}

func Load() *Config {
	return &Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://postgres:password@localhost:5432/tmf_party_db?sslmode=disable"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
