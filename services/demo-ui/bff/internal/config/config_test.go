package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Unset to test default values
	os.Unsetenv("RABBITMQ_URL")
	os.Unsetenv("AUTH0_DOMAIN")
	os.Unsetenv("AUTH0_AUDIENCE")
	os.Unsetenv("PORT")

	cfg := Load()

	if cfg.RabbitMQURL != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("expected amqp://guest:guest@localhost:5672/, got %s", cfg.RabbitMQURL)
	}
	if cfg.Auth0Domain != "rohekood.eu.auth0.com" {
		t.Errorf("expected rohekood.eu.auth0.com, got %s", cfg.Auth0Domain)
	}
	if cfg.Auth0Audience != "http://localhost/api" {
		t.Errorf("expected http://localhost/api, got %s", cfg.Auth0Audience)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected 8080, got %s", cfg.Port)
	}

	// Set env vars
	os.Setenv("RABBITMQ_URL", "amqp://test:5672")
	os.Setenv("AUTH0_DOMAIN", "test.auth0.com")
	os.Setenv("AUTH0_AUDIENCE", "test-audience")
	os.Setenv("PORT", "9090")

	cfg2 := Load()

	if cfg2.RabbitMQURL != "amqp://test:5672" {
		t.Errorf("expected amqp://test:5672, got %s", cfg2.RabbitMQURL)
	}
	if cfg2.Auth0Domain != "test.auth0.com" {
		t.Errorf("expected test.auth0.com, got %s", cfg2.Auth0Domain)
	}
	if cfg2.Auth0Audience != "test-audience" {
		t.Errorf("expected test-audience, got %s", cfg2.Auth0Audience)
	}
	if cfg2.Port != "9090" {
		t.Errorf("expected 9090, got %s", cfg2.Port)
	}
}
