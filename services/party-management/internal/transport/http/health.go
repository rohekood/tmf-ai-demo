package http

import (
	"encoding/json"
	"net/http"
	"tmf/services/party-management/internal/infrastructure/rabbitmq"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gorm.io/gorm"
)

type HealthHandler struct {
	db     *gorm.DB
	rabbit *rabbitmq.ConnectionManager
}

func NewHealthHandler(db *gorm.DB, rabbit *rabbitmq.ConnectionManager) *HealthHandler {
	return &HealthHandler{db: db, rabbit: rabbit}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := "UP"
	details := make(map[string]string)

	// Check DB
	sqlDB, err := h.db.DB()
	if err != nil {
		status = "DOWN"
		details["db"] = "failed to get sql.DB"
	} else if err := sqlDB.Ping(); err != nil {
		status = "DOWN"
		details["db"] = err.Error()
	} else {
		details["db"] = "OK"
	}

	// Check Rabbit
	if h.rabbit.GetConnection() == nil || h.rabbit.GetConnection().IsClosed() {
		status = "DOWN"
		details["rabbitmq"] = "connection closed"
	} else {
		details["rabbitmq"] = "OK"
	}

	w.Header().Set("Content-Type", "application/json")
	if status == "DOWN" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  status,
		"details": details,
	})
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
