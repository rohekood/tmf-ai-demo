package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db         *gorm.DB
	rabbitConn *amqp.Connection
}

func NewHealthHandler(db *gorm.DB, rabbitConn *amqp.Connection) *HealthHandler {
	return &HealthHandler{
		db:         db,
		rabbitConn: rabbitConn,
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := "UP"
	details := make(map[string]string)

	// Check DB
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		status = "DOWN"
		details["database"] = "unreachable"
	} else {
		details["database"] = "connected"
	}

	// Check RabbitMQ
	if h.rabbitConn == nil || h.rabbitConn.IsClosed() {
		status = "DOWN"
		details["rabbitmq"] = "disconnected"
	} else {
		details["rabbitmq"] = "connected"
	}

	w.Header().Set("Content-Type", "application/json")
	if status == "DOWN" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  status,
		"details": details,
	}); err != nil {
		slog.Error("failed to encode health response", "error", err)
	}
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
