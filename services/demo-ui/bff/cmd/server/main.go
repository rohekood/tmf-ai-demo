package main

import (
	"log"
	"net/http"
	"os"
	"tmf/services/demo-ui/bff/internal/auth"
	httpTransport "tmf/services/demo-ui/bff/internal/transport/http"
	"tmf/services/demo-ui/bff/internal/transport/rabbitmq"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// 1. Initialize RabbitMQ RPC Client
	rpcClient, err := rabbitmq.NewClient(os.Getenv("RABBITMQ_URL"))
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
	// TODO: Configure with real Okta credentials from env
	r.Use(auth.Middleware)

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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("BFF Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
