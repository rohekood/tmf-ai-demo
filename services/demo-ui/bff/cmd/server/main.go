package main

import (
	"log"
	"net/http"
	"tmf/services/demo-ui/bff/internal/auth"
	"tmf/services/demo-ui/bff/internal/config"
	httpTransport "tmf/services/demo-ui/bff/internal/transport/http"
	"tmf/services/demo-ui/bff/internal/transport/rabbitmq"
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

	// 2. Initialize Mux
	mux := http.NewServeMux()

	// 3. Initialize Auth Validator (used by both WebSocket and API routes)
	authValidator, err := auth.NewAuth0Validator(cfg.Auth0Domain, cfg.Auth0Audience)
	if err != nil {
		log.Fatalf("Failed to initialize auth validator: %v", err)
	}

	// 4. Initialize WebSocket Hub with token validator for authentication
	hub := httpTransport.NewHub()
	hub.SetTokenValidator(authValidator)
	go hub.Run()

	// 5. Set broadcaster on RPC client for debug reply forwarding
	rpcClient.SetBroadcaster(hub)

	// 6. Initialize Debug Consumer
	debugConsumer := rabbitmq.NewDebugConsumer(rpcClient, hub)
	go func() {
		if err := debugConsumer.StartSubscribing("tmf.events"); err != nil {
			log.Printf("Failed to start debug subscriber: %v", err)
		}
	}()

	// 7. Register WebSocket route
	mux.HandleFunc("/ws/debug", func(w http.ResponseWriter, req *http.Request) {
		hub.ServeWs(w, req)
	})

	// 8. Register API Routes
	handler := httpTransport.NewHandler(rpcClient, hub)
	handler.RegisterRoutes(mux)

	// 9. Construct Middleware Chain
	// Conditional Auth Middleware: Only applies to routes starting with /api/
	authMiddleware := auth.EnsureValidToken(authValidator, cfg.Auth0Domain, cfg.Auth0Audience)

	// Wrap mux with Auth middleware selectively
	var authenticatedHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple prefix check: if path starts with /api, require auth
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			authMiddleware(mux).ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	})

	// Apply Global Middlewares (executed in reverse order of chaining)
	// Recoverer -> Logger -> CORS -> AuthenticatedHandler -> Mux
	finalHandler := httpTransport.Chain(
		authenticatedHandler,
		httpTransport.CORSMiddleware,
		httpTransport.LoggerMiddleware,
		httpTransport.RecovererMiddleware,
	)

	// 10. Start Server
	log.Printf("BFF Server listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, finalHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
