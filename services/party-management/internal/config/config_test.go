package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	os.Clearenv()

	cfg := Load()
	if cfg.RabbitMQURL != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("Expected default RabbitMQURL, got %s", cfg.RabbitMQURL)
	}
	if cfg.PostgresURL != "postgres://postgres:password@localhost:5432/tmf_party_db?sslmode=disable" {
		t.Errorf("Expected default PostgresURL, got %s", cfg.PostgresURL)
	}
	if cfg.HTTPPort != "8080" {
		t.Errorf("Expected default HTTPPort, got %s", cfg.HTTPPort)
	}

	os.Setenv("RABBITMQ_URL", "amqp://test")
	os.Setenv("POSTGRES_URL", "postgres://test")
	os.Setenv("HTTP_PORT", "9090")

	cfg = Load()
	if cfg.RabbitMQURL != "amqp://test" {
		t.Errorf("Expected amqp://test, got %s", cfg.RabbitMQURL)
	}
	if cfg.PostgresURL != "postgres://test" {
		t.Errorf("Expected postgres://test, got %s", cfg.PostgresURL)
	}
	if cfg.HTTPPort != "9090" {
		t.Errorf("Expected 9090, got %s", cfg.HTTPPort)
	}
}
