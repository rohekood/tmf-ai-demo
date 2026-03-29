package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	keys := []string{"RABBITMQ_URL", "AUTH0_DOMAIN", "AUTH0_AUDIENCE", "PORT"}
	originalValues := make(map[string]string, len(keys))
	originalPresence := make(map[string]bool, len(keys))
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		originalPresence[key] = ok
		if ok {
			originalValues[key] = value
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
	t.Cleanup(func() {
		for _, key := range keys {
			if !originalPresence[key] {
				if err := os.Unsetenv(key); err != nil {
					t.Fatalf("cleanup unset %s: %v", key, err)
				}
				continue
			}
			if err := os.Setenv(key, originalValues[key]); err != nil {
				t.Fatalf("cleanup restore %s: %v", key, err)
			}
		}
	})

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

	for key, value := range map[string]string{
		"RABBITMQ_URL":   "amqp://test:5672",
		"AUTH0_DOMAIN":   "test.auth0.com",
		"AUTH0_AUDIENCE": "test-audience",
		"PORT":           "9090",
	} {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("set %s: %v", key, err)
		}
	}

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
