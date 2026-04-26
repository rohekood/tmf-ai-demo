package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"

	"gorm.io/gorm"

	"tmf/pkg/rabbitmq"
	transportRabbit "tmf/services/customer-management/internal/transport/rabbitmq"
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

type mockConsumer struct{}

func (m *mockConsumer) Consume(handler func(msg amqp091.Delivery)) error {
	return nil
}

func (m *mockConsumer) Subscribe(routingKey string, handler rabbitmq.ConsumerHandler) error {
	return nil
}

func (m *mockConsumer) Close() error {
	return nil
}

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "../..")
	_ = os.Chdir(dir)
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "TEST_VALUE")
	defer os.Unsetenv("TEST_KEY")

	if val := getEnv("TEST_KEY", "fallback"); val != "TEST_VALUE" {
		t.Errorf("expected TEST_VALUE, got %s", val)
	}

	if val := getEnv("NON_EXISTENT_KEY", "fallback"); val != "fallback" {
		t.Errorf("expected fallback, got %s", val)
	}
}

func TestRun_DBError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	envFn := func(k, f string) string {
		if k == "POSTGRES_URL" {
			return "invalid_dsn"
		}
		return f
	}

	dbDialFn := func(dsn string) (*gorm.DB, error) {
		return nil, fmt.Errorf("mock db dial error")
	}

	rabbitDialFn := func(url string) (*amqp.Connection, error) {
		return nil, nil // unreached
	}

	runMigrationsFn := func(string) error {
		return fmt.Errorf("failed to create migration instance")
	}

	newPublisherFn := func(*amqp091.Connection) (rabbitmq.Publisher, error) { return nil, nil }
	newListenerFn := func(*amqp091.Connection) (*transportRabbit.Listener, error) { return nil, nil }
	newRpcConsumerFn := func(*amqp091.Connection) (rabbitmq.Consumer, error) { return nil, nil }

	err := run(ctx, envFn, dbDialFn, rabbitDialFn, runMigrationsFn, newPublisherFn, newListenerFn, newRpcConsumerFn, logger)
	if err == nil {
		t.Fatal("expected error from invalid DB URL, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create migration instance") && !strings.Contains(err.Error(), "failed to run migrations") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRun_DBOpenError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	envFn := func(k, f string) string {
		if k == "POSTGRES_URL" {
			// This bypasses migrate parse error but fails dialing
			return "postgres://invalid"
		}
		return f
	}

	dbDialFn := func(dsn string) (*gorm.DB, error) {
		return nil, fmt.Errorf("mock db connection error")
	}

	rabbitDialFn := func(url string) (*amqp.Connection, error) {
		return nil, nil // unreached
	}

	runMigrationsFn := func(string) error {
		return nil
	}

	newPublisherFn := func(*amqp091.Connection) (rabbitmq.Publisher, error) { return nil, nil }
	newListenerFn := func(*amqp091.Connection) (*transportRabbit.Listener, error) { return nil, nil }
	newRpcConsumerFn := func(*amqp091.Connection) (rabbitmq.Consumer, error) { return nil, nil }

	err := run(ctx, envFn, dbDialFn, rabbitDialFn, runMigrationsFn, newPublisherFn, newListenerFn, newRpcConsumerFn, logger)
	if err == nil {
		t.Fatal("expected error from DB dial, got nil")
	}
	if !strings.Contains(err.Error(), "mock db connection error") {
		t.Errorf("expected mock db connection error, got %v", err)
	}
}

func TestRunMigrations_Error(t *testing.T) {
	err := runMigrations("invalid_dsn")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_RabbitError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	envFn := func(k, f string) string {
		if k == "POSTGRES_URL" {
			// This URL will pass migration parsing but will fail to connect,
			// allowing the execution to reach the rabbitmq dial.
			return "postgres://invalid"
		}
		return f
	}

	dbDialFn := func(dsn string) (*gorm.DB, error) {
		return sharedDB, nil
	}

	rabbitDialFn := func(url string) (*amqp.Connection, error) {
		return nil, fmt.Errorf("mock rabbit connection error")
	}

	runMigrationsFn := func(string) error {
		return nil
	}

	newPublisherFn := func(*amqp091.Connection) (rabbitmq.Publisher, error) { return nil, nil }
	newListenerFn := func(*amqp091.Connection) (*transportRabbit.Listener, error) { return nil, nil }
	newRpcConsumerFn := func(*amqp091.Connection) (rabbitmq.Consumer, error) { return nil, nil }

	err := run(ctx, envFn, dbDialFn, rabbitDialFn, runMigrationsFn, newPublisherFn, newListenerFn, newRpcConsumerFn, logger)
	if err == nil {
		t.Fatal("expected error from RabbitMQ dial, got nil")
	}
	if !strings.Contains(err.Error(), "mock rabbit connection error") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRun_PublisherError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	envFn := func(k, f string) string { return f }
	dbDialFn := func(string) (*gorm.DB, error) {
		return sharedDB, nil
	}
	rabbitDialFn := func(string) (*amqp091.Connection, error) { return &amqp091.Connection{}, nil }
	runMigrationsFn := func(string) error { return nil }
	newPublisherFn := func(*amqp091.Connection) (rabbitmq.Publisher, error) {
		return nil, fmt.Errorf("mock publisher error")
	}
	newListenerFn := func(*amqp091.Connection) (*transportRabbit.Listener, error) { return nil, nil }
	newRpcConsumerFn := func(*amqp091.Connection) (rabbitmq.Consumer, error) { return nil, nil }

	err := run(ctx, envFn, dbDialFn, rabbitDialFn, runMigrationsFn, newPublisherFn, newListenerFn, newRpcConsumerFn, logger)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mock publisher error") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRun_ListenerError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	envFn := func(k, f string) string { return f }
	dbDialFn := func(string) (*gorm.DB, error) {
		return sharedDB, nil
	}
	rabbitDialFn := func(string) (*amqp091.Connection, error) { return &amqp091.Connection{}, nil }
	runMigrationsFn := func(string) error { return nil }
	newPublisherFn := func(*amqp091.Connection) (rabbitmq.Publisher, error) { return &mockPublisher{}, nil }
	newListenerFn := func(*amqp091.Connection) (*transportRabbit.Listener, error) {
		return nil, fmt.Errorf("mock listener error")
	}
	newRpcConsumerFn := func(*amqp091.Connection) (rabbitmq.Consumer, error) { return nil, nil }

	err := run(ctx, envFn, dbDialFn, rabbitDialFn, runMigrationsFn, newPublisherFn, newListenerFn, newRpcConsumerFn, logger)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mock listener error") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRun_RpcConsumerError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	envFn := func(k, f string) string { return f }
	dbDialFn := func(string) (*gorm.DB, error) {
		return sharedDB, nil
	}
	rabbitDialFn := func(string) (*amqp091.Connection, error) { return &amqp091.Connection{}, nil }
	runMigrationsFn := func(string) error { return nil }
	newPublisherFn := func(*amqp091.Connection) (rabbitmq.Publisher, error) { return &mockPublisher{}, nil }
	newListenerFn := func(*amqp091.Connection) (*transportRabbit.Listener, error) {
		return &transportRabbit.Listener{}, nil
	}
	newRpcConsumerFn := func(*amqp091.Connection) (rabbitmq.Consumer, error) {
		return nil, fmt.Errorf("mock rpc consumer error")
	}

	err := run(ctx, envFn, dbDialFn, rabbitDialFn, runMigrationsFn, newPublisherFn, newListenerFn, newRpcConsumerFn, logger)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mock rpc consumer error") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRun_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	envFn := func(k, f string) string { return "0" } // port 0 for random HTTP port
	dbDialFn := func(string) (*gorm.DB, error) {
		return sharedDB, nil
	}
	rabbitDialFn := func(string) (*amqp091.Connection, error) { return &amqp091.Connection{}, nil }
	runMigrationsFn := func(string) error { return nil }
	newPublisherFn := func(*amqp091.Connection) (rabbitmq.Publisher, error) { return &mockPublisher{}, nil }
	newListenerFn := func(*amqp091.Connection) (*transportRabbit.Listener, error) {
		return nil, nil
	}
	newRpcConsumerFn := func(*amqp091.Connection) (rabbitmq.Consumer, error) { return &mockConsumer{}, nil }

	err := run(ctx, envFn, dbDialFn, rabbitDialFn, runMigrationsFn, newPublisherFn, newListenerFn, newRpcConsumerFn, logger)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMain_Func(t *testing.T) {
	// Backup original os.Exit and restore it after
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	os.Setenv("POSTGRES_URL", "invalid_dsn")
	defer os.Unsetenv("POSTGRES_URL")

	main()

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

var sharedDB *gorm.DB = &gorm.DB{}
