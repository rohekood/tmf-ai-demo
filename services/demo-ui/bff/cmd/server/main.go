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

	// 4. Initialize Auth Validator (used by both WebSocket and API routes)
	authValidator, err := auth.NewAuth0Validator(cfg.Auth0Domain, cfg.Auth0Audience)
	if err != nil {
		log.Fatalf("Failed to initialize auth validator: %v", err)
	}

	// 5. Initialize WebSocket Hub with token validator for authentication
	hub := httpTransport.NewHub()
	hub.SetTokenValidator(authValidator)
	go hub.Run()

	// 5a. Set broadcaster on RPC client for debug reply forwarding
	rpcClient.SetBroadcaster(hub)

	// 6. Initialize Debug Consumer
	debugConsumer := rabbitmq.NewDebugConsumer(rpcClient, hub)
	go func() {
		if err := debugConsumer.StartSubscribing("tmf.events"); err != nil {
			log.Printf("Failed to start debug subscriber: %v", err)
		}
	}()

	// 7. Register WebSocket route BEFORE auth group (handles its own auth via Sec-WebSocket-Protocol)
	r.Get("/ws/debug", func(w http.ResponseWriter, req *http.Request) {
		hub.ServeWs(w, req)
	})

	// 8. Auth-protected routes group
	r.Group(func(r chi.Router) {
		// Apply auth middleware only to this group
		r.Use(auth.EnsureValidToken(authValidator, cfg.Auth0Domain, cfg.Auth0Audience))

		// Register API Routes
		handler := httpTransport.NewHandler(rpcClient, hub)
		handler.RegisterRoutes(r)
	})

	// 6. Start Server
	log.Printf("BFF Server listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
