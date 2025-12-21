package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tmf/services/party-management/internal/config"
	"tmf/services/party-management/internal/infrastructure/postgres"
	infraRabbit "tmf/services/party-management/internal/infrastructure/rabbitmq"
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

	// 6. Setup Publisher
	publisher, err := infraRabbit.NewPublisher(conn)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// Declare the events exchange
	ch, _ := publisher.GetChannel()
	err = ch.ExchangeDeclare(
		rabbitTransport.EventExchange, // name
		"topic",                       // type
		true,                          // durable
		false,                         // auto-deleted
		false,                         // internal
		false,                         // no-wait
		nil,                           // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	// 7. Setup Listener
	listener, err := rabbitTransport.NewListener(conn)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}

	// 8. Setup Handlers
	handlers := rabbitTransport.NewHandlers(repo, publisher)

	// 9. Register Command Listeners
	go func() {
		if err := listener.Listen(rabbitTransport.CmdPartyCreate, handlers.HandleCreateParty); err != nil {
			log.Printf("Failed to listen on %s: %v", rabbitTransport.CmdPartyCreate, err)
		}
	}()
	go func() {
		if err := listener.Listen(rabbitTransport.CmdPartyUpdate, handlers.HandleUpdateParty); err != nil {
			log.Printf("Failed to listen on %s: %v", rabbitTransport.CmdPartyUpdate, err)
		}
	}()
	go func() {
		if err := listener.Listen(rabbitTransport.CmdPartyPatch, handlers.HandlePatchParty); err != nil {
			log.Printf("Failed to listen on %s: %v", rabbitTransport.CmdPartyPatch, err)
		}
	}()
	go func() {
		if err := listener.Listen(rabbitTransport.CmdPartyDelete, handlers.HandleDeleteParty); err != nil {
			log.Printf("Failed to listen on %s: %v", rabbitTransport.CmdPartyDelete, err)
		}
	}()

	// 10. Register Query Listeners
	go func() {
		if err := listener.Listen(rabbitTransport.QueryPartyGet, handlers.HandleGetParty); err != nil {
			log.Printf("Failed to listen on %s: %v", rabbitTransport.QueryPartyGet, err)
		}
	}()
	go func() {
		if err := listener.Listen(rabbitTransport.QueryPartySearch, handlers.HandleSearchParty); err != nil {
			log.Printf("Failed to listen on %s: %v", rabbitTransport.QueryPartySearch, err)
		}
	}()

	log.Println("Party Management Service is running. Press Ctrl+C to exit.")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")
}
