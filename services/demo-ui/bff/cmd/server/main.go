package main

import (
	"log"
	"net/http"
	"tmf/services/demo-ui/bff/internal/auth"
	"tmf/services/demo-ui/bff/internal/config"
	httpTransport "tmf/services/demo-ui/bff/internal/transport/http"
	"tmf/services/demo-ui/bff/internal/transport/rabbitmq"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// 0. Load Config
	cfg := config.Load()

	// 1. Initialize RabbitMQ RPC Client
	rpcClient, err := rabbitmq.NewClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rpcClient.Close()

	// 2. Initialize Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 3. CORS Configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:80", "http://localhost"}, // Allow UI
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 4. Auth Middleware
	// Strictly enforces Valid JWT
	authValidator, err := auth.NewAuth0Validator(cfg.Auth0Domain, cfg.Auth0Audience)
	if err != nil {
		log.Fatalf("Failed to initialize auth validator: %v", err)
	}
	r.Use(auth.EnsureValidToken(authValidator, cfg.Auth0Domain, cfg.Auth0Audience))

	// 5. Initialize WebSocket Hub
	hub := httpTransport.NewHub()
	go hub.Run()

	// 6. Initialize Debug Consumer
	debugConsumer := rabbitmq.NewDebugConsumer(rpcClient, hub)
	go func() {
		if err := debugConsumer.StartSubscribing("tmf.events"); err != nil {
			log.Printf("Failed to start debug subscriber: %v", err)
		}
	}()

	// 7. Register Routes
	handler := httpTransport.NewHandler(rpcClient, hub)
	handler.RegisterRoutes(r)

	// 6. Start Server
	log.Printf("BFF Server listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
