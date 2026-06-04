package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	rabbitmqpkg "tmf/pkg/rabbitmq"
	"tmf/services/demo-ui/bff/internal/auth"
	"tmf/services/demo-ui/bff/internal/config"
	"tmf/services/demo-ui/bff/internal/events"
	httpTransport "tmf/services/demo-ui/bff/internal/transport/http"
	"tmf/services/demo-ui/bff/internal/transport/rabbitmq"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

var (
	newClientFunc = rabbitmq.NewClient
	newConsumerFunc = func(url, exchange, queue string, opts ...rabbitmqpkg.ConsumerOption) (rabbitmqpkg.Consumer, error) {
		return rabbitmqpkg.NewConsumer(url, exchange, queue, append(opts, rabbitmqpkg.WithMessageTimeout(30*time.Second))...)
	}
	listenAndServeFunc = func(srv *http.Server) error { return srv.ListenAndServe() }
	newAuthValidatorFunc = auth.NewAuth0Validator
	startDebugConsumerFunc = func(dc *rabbitmq.DebugConsumer, exchange string) error { return dc.StartSubscribing(exchange) }
)

func main() {
	// 0. Load Config
	cfg := config.Load()

	// 1. Initialize RabbitMQ RPC Client
	rpcClient, err := newClientFunc(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rpcClient.Close()

	// 2. Initialize Mux
	mux := http.NewServeMux()

	// 3. Initialize Auth Validator (used by both WebSocket and API routes)
	authValidator, err := newAuthValidatorFunc(cfg.Auth0Domain, cfg.Auth0Audience)
	if err != nil {
		log.Fatalf("Failed to initialize auth validator: %v", err)
	}

	// 4. Initialize WebSocket Hub (Legacy Debug + New Event)
	hub := httpTransport.NewHub()
	hub.SetTokenValidator(authValidator)
	go hub.Run()

	// 4b. Initialize New Event Hub (Global)
	eventHub := events.NewHub()

	// 4c. Start RabbitMQ Consumer for Events
	eventConsumer, err := newConsumerFunc(cfg.RabbitMQURL, "catalog_events", "q.bff.events")
	if err != nil {
		log.Fatalf("Failed to create Event Consumer: %v", err)
	}
	go eventHub.StartConsumer(eventConsumer)

	// 5. Set broadcaster on RPC client for debug reply forwarding
	rpcClient.SetBroadcaster(hub)

	// 6. Initialize Debug Consumer
	debugConsumer := rabbitmq.NewDebugConsumer(rpcClient, hub)
	go func() {
		if err := startDebugConsumerFunc(debugConsumer, "tmf.events"); err != nil {
			log.Printf("Failed to start debug subscriber: %v", err)
		}
	}()

	// 7. Register WebSocket routes
	mux.HandleFunc("/ws/debug", func(w http.ResponseWriter, req *http.Request) {
		hub.ServeWs(w, req)
	})
	// NEW Route for specific event streaming
	mux.HandleFunc("/ws/events", func(w http.ResponseWriter, req *http.Request) {
		// Upgrade HTTP connection
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}

		// TODO: Extract proper CorrelationID or UserID from Token
		// For now, generate a temporary ID or use a query param
		clientID := req.URL.Query().Get("id")
		if clientID == "" {
			clientID = "unknown-" + time.Now().UTC().String()
		}

		// Register with Global Event Hub
		eventHub.Register(clientID, conn)

		// Keep connection alive (Hub handles writes, we handle read/Ping)
		// closeHandler/readLoop needed here ideally
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

	// 10. Start Server with Graceful Shutdown
	// Create a context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: finalHandler,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		log.Printf("BFF Server listening on port %s", cfg.Port)
		if err := listenAndServeFunc(srv); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
