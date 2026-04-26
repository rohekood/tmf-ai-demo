package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"

	"tmf/pkg/rabbitmq"
	rabbitTransport "tmf/services/party-management/internal/transport/rabbitmq"
)

type mockPublisher struct{}

func (m *mockPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	return nil
}
func (m *mockPublisher) Publish(ctx context.Context, exchange, routingKey string, body any) error {
	return nil
}
func (m *mockPublisher) PublishToQueue(ctx context.Context, exchange, queue string, body any) error {
	return nil
}
func (m *mockPublisher) Close() error {
	return nil
}

type mockConnManager struct {
	connectErr error
}

func (m *mockConnManager) Connect() error                  { return m.connectErr }
func (m *mockConnManager) Close() error                    { return nil }
func (m *mockConnManager) GetConnection() *amqp.Connection { return nil }

func TestRun_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	os.Setenv("HTTP_PORT", "0")
	defer os.Unsetenv("HTTP_PORT")
	envFn := func(s string) string { return "0" } // random port

	dbDialFn := func(dsn string) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	}

	rabbitConnFn := func(url string) rmqConnManager {
		return &mockConnManager{}
	}

	runMigrationsFn := func(dsn string) error {
		return nil
	}

	newPublisherFn := func(connMgr rmqConnManager) (rabbitmq.Publisher, error) {
		return &mockPublisher{}, nil
	}

	newListenerFn := func(connMgr rmqConnManager) (*rabbitTransport.Listener, error) {
		return nil, nil
	}

	err := run(ctx, envFn, dbDialFn, rabbitConnFn, runMigrationsFn, newPublisherFn, newListenerFn, logger)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// For coverage of listener error handling in go routine, we let it sleep a bit
	time.Sleep(20 * time.Millisecond)
}

func TestRun_DBError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	envFn := func(s string) string { return "" }

	dbDialFn := func(dsn string) (*gorm.DB, error) {
		return nil, fmt.Errorf("mock db error")
	}

	err := run(ctx, envFn, dbDialFn, nil, nil, nil, nil, logger)
	if err == nil || err.Error() != "failed to connect to database: mock db error" {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestRun_MigrationError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbDialFn := func(dsn string) (*gorm.DB, error) { return &gorm.DB{}, nil }
	runMigrationsFn := func(dsn string) error { return fmt.Errorf("mock migration error") }

	err := run(ctx, func(s string) string { return "" }, dbDialFn, nil, runMigrationsFn, nil, nil, logger)
	if err == nil || err.Error() != "failed to run migrations: mock migration error" {
		t.Fatalf("expected migration error, got %v", err)
	}
}

func TestRun_RabbitConnError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbDialFn := func(dsn string) (*gorm.DB, error) { return &gorm.DB{}, nil }
	runMigrationsFn := func(dsn string) error { return nil }

	rabbitConnFn := func(url string) rmqConnManager {
		return &mockConnManager{connectErr: fmt.Errorf("failed to dial: mock dial error")}
	}

	err := run(ctx, func(s string) string { return "" }, dbDialFn, rabbitConnFn, runMigrationsFn, nil, nil, logger)
	if err == nil || err.Error() != "failed to connect to rabbitmq: failed to dial: mock dial error" {
		t.Fatalf("expected rabbit error, got %v", err)
	}
}

func TestRun_PublisherError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbDialFn := func(dsn string) (*gorm.DB, error) { return &gorm.DB{}, nil }
	runMigrationsFn := func(dsn string) error { return nil }

	rabbitConnFn := func(url string) rmqConnManager {
		return &mockConnManager{}
	}

	newPublisherFn := func(connMgr rmqConnManager) (rabbitmq.Publisher, error) {
		return nil, fmt.Errorf("mock pub error")
	}

	err := run(ctx, func(s string) string { return "" }, dbDialFn, rabbitConnFn, runMigrationsFn, newPublisherFn, nil, logger)
	if err == nil || err.Error() != "failed to create publisher: mock pub error" {
		t.Fatalf("expected publisher error, got %v", err)
	}
}

func TestRun_ListenerError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbDialFn := func(dsn string) (*gorm.DB, error) { return &gorm.DB{}, nil }
	runMigrationsFn := func(dsn string) error { return nil }

	rabbitConnFn := func(url string) rmqConnManager {
		return &mockConnManager{}
	}

	newPublisherFn := func(connMgr rmqConnManager) (rabbitmq.Publisher, error) {
		return &mockPublisher{}, nil
	}

	newListenerFn := func(connMgr rmqConnManager) (*rabbitTransport.Listener, error) {
		return nil, fmt.Errorf("mock listener error")
	}

	err := run(ctx, func(s string) string { return "" }, dbDialFn, rabbitConnFn, runMigrationsFn, newPublisherFn, newListenerFn, logger)
	if err == nil || err.Error() != "failed to create listener: mock listener error" {
		t.Fatalf("expected listener error, got %v", err)
	}
}

type mockDeclareFailPublisher struct{}

func (m *mockDeclareFailPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	return fmt.Errorf("mock exchange declare error")
}
func (m *mockDeclareFailPublisher) Publish(ctx context.Context, exchange, routingKey string, body any) error {
	return nil
}
func (m *mockDeclareFailPublisher) PublishToQueue(ctx context.Context, exchange, queue string, body any) error {
	return nil
}
func (m *mockDeclareFailPublisher) Close() error {
	return nil
}

func TestRun_DeclareExchangeError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dbDialFn := func(dsn string) (*gorm.DB, error) { return &gorm.DB{}, nil }
	runMigrationsFn := func(dsn string) error { return nil }

	rabbitConnFn := func(url string) rmqConnManager {
		return &mockConnManager{}
	}

	newPublisherFn := func(connMgr rmqConnManager) (rabbitmq.Publisher, error) {
		return &mockDeclareFailPublisher{}, nil
	}

	err := run(ctx, func(s string) string { return "" }, dbDialFn, rabbitConnFn, runMigrationsFn, newPublisherFn, nil, logger)
	if err == nil || err.Error() != "failed to declare exchange (tmf.events): mock exchange declare error" {
		t.Fatalf("expected exchange declare error, got %v", err)
	}
}
