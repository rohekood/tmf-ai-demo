package config

import (
	"os"
)

type Config struct {
	RabbitMQURL string
	PostgresURL string
	HTTPPort    string
	// AutoDeclareTopology controls whether the service creates its RabbitMQ
	// exchanges/queues on startup. Enabled for local/dev (RABBITMQ_AUTO_DECLARE=true);
	// disabled in production, where the topology is provisioned out-of-band
	// (e.g. RabbitMQ Operator CRDs) and services must only bind/consume.
	AutoDeclareTopology bool
}

func Load() *Config {
	return &Config{
		RabbitMQURL:         getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		PostgresURL:         getEnv("POSTGRES_URL", "postgres://postgres:password@localhost:5432/tmf_party_db?sslmode=disable"),
		HTTPPort:            getEnv("HTTP_PORT", "8080"),
		AutoDeclareTopology: getEnv("RABBITMQ_AUTO_DECLARE", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
