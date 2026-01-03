package handler_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tmf/services/product-catalog-management/internal/adapter/repository"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	sharedDB        *gorm.DB
	rabbitConn      *amqp.Connection
	amqpURL         string
	rabbitContainer *rabbitmq.RabbitMQContainer
	pgContainer     *tcpostgres.PostgresContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Start Postgres
	pg, err := tcpostgres.Run(ctx,
		"postgres:15",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start postgres: %v", err)
	}
	pgContainer = pg

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres connection: %v", err)
	}

	sharedDB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	err = sharedDB.AutoMigrate(
		&repository.CatalogModel{},
		&repository.CategoryModel{},
		&repository.ProductSpecificationModel{},
		&repository.ProductOfferingModel{},
	)
	if err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	// 2. Start RabbitMQ
	rmq, err := rabbitmq.Run(ctx,
		"rabbitmq:3.12-management",
		rabbitmq.WithAdminPassword("guest"),
		rabbitmq.WithAdminUsername("guest"),
	)
	if err != nil {
		log.Fatalf("failed to start rabbitmq: %v", err)
	}
	rabbitContainer = rmq

	amqpURL, err = rmq.AmqpURL(ctx)
	if err != nil {
		log.Fatalf("failed to get amqp url: %v", err)
	}

	rabbitConn, err = amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	// 3. Run Tests
	code := m.Run()

	// 4. Cleanup
	// 4. Cleanup
	if err := rabbitConn.Close(); err != nil {
		log.Printf("failed to close rabbitConn: %v", err)
	}
	if err := rmq.Terminate(ctx); err != nil {
		log.Printf("failed to terminate rabbitmq: %v", err)
	}
	if err := pg.Terminate(ctx); err != nil {
		log.Printf("failed to terminate postgres: %v", err)
	}

	os.Exit(code)
}
