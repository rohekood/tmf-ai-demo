package config

import (
	"os"
)

type Config struct {
	RabbitMQURL   string
	Auth0Domain   string
	Auth0Audience string
	Port          string
}

func Load() *Config {
	return &Config{
		RabbitMQURL:   getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Auth0Domain:   getEnv("AUTH0_DOMAIN", "rohekood.eu.auth0.com"),
		Auth0Audience: getEnv("AUTH0_AUDIENCE", "http://localhost/api"),
		Port:          getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
