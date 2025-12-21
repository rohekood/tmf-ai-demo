package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tmf/services/party-management/internal/config"
	"tmf/services/party-management/internal/domain"
	"tmf/services/party-management/internal/infrastructure/postgres"
	rabbitTransport "tmf/services/party-management/internal/transport/rabbitmq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	amqp "github.com/rabbitmq/amqp091-go"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Party Management Service...")

	// 1. Load Config
	cfg := config.Load()

	// 2. Initialize GORM
	db, err := gorm.Open(gormPostgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database using GORM: %v", err)
	}

	// 3. Run Migrations
	log.Println("Running database migrations...")
	// Note: We need a valid connection string for migrate, ensuring parameters like sslmode are correct for the driver
	// Config DSN: "postgres://postgres:postgres@localhost:5432/party_db?sslmode=disable"
	m, err := migrate.New(
		"file://internal/infrastructure/postgres/migrations",
		cfg.PostgresDSN,
	)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations applied successfully.")

	// 4. Initialize Repository
	repo := postgres.NewPartyRepository(db)

	// 5. Connect to RabbitMQ
	var conn *amqp.Connection
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(cfg.RabbitMQURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ, retrying in 2s... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after retries: %v", err)
	}
	defer conn.Close()

	// 6. Setup Listener
	listener, err := rabbitTransport.NewListener(conn)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}

	// Handler for creating individual
	createIndividualHandler := func(d amqp.Delivery) error {
		log.Printf("Received Create Individual command: %s", string(d.Body))
		var ind domain.Individual
		if err := json.Unmarshal(d.Body, &ind); err != nil {
			return err
		}
		// Set defaults if needed
		if ind.ID == "" {
			ind.ID = "generated-id-" + time.Now().Format(time.RFC3339Nano)
		}
		ind.Type = domain.PartyTypeIndividual
		ind.CreatedAt = time.Now()
		ind.UpdatedAt = time.Now()

		if err := repo.CreateIndividual(&ind); err != nil {
			log.Printf("Failed to create individual: %v", err)
			return err
		}
		log.Printf("Successfully created individual: %s", ind.ID)
		return nil
	}

	// Start Listening
	go func() {
		if err := listener.Listen("cmd.party.create", createIndividualHandler); err != nil {
			log.Printf("Failed to start listening: %v", err)
		}
	}()

	log.Println("Party Management Service is running. Press Ctrl+C to exit.")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")
}
