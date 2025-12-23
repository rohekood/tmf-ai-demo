package config

import (
	"os"
)

type Config struct {
	RabbitMQURL string
	PostgresURL string
}

func Load() *Config {
	return &Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/party_db?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
