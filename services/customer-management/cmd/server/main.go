package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tmf/services/customer-management/internal/infrastructure/postgres"
	infraRabbit "tmf/services/customer-management/internal/infrastructure/rabbitmq"
	"tmf/services/customer-management/internal/transport/rabbitmq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	amqp "github.com/rabbitmq/amqp091-go"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Configuration (using defaults or env vars)
	dbURL := getEnv("DB_URL", "postgres://postgres:password@localhost:5432/tmf_customer_db?sslmode=disable")
	rabbitURL := getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/")

	// 1. Database Migrations
	runMigrations(dbURL)

	// 2. Database Connection
	db, err := gorm.Open(gormPostgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	repo := postgres.NewCustomerRepository(db)

	// 3. RabbitMQ Connection
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	publisher, err := infraRabbit.NewPublisher(conn)
	if err != nil {
		log.Fatalf("failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// 4. Handlers & Listener
	handlers := rabbitmq.NewHandlers(repo, publisher)
	listener, err := rabbitmq.NewListener(conn)
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}

	// 5. Start Service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := listener.Start(ctx, handlers); err != nil {
			log.Printf("listener stopped: %v", err)
			cancel()
		}
	}()

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down Customer Management service...")
	cancel()
	time.Sleep(1 * time.Second) // Give some time for cleanup
}

func runMigrations(dbURL string) {
	m, err := migrate.New(
		"file://internal/infrastructure/postgres/migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("failed to create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully.")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
