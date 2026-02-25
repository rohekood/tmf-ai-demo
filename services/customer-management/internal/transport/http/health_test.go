package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMetricsHandler(t *testing.T) {
	handler := MetricsHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthHandler_ServeHTTP(t *testing.T) {
	// Create a DB with invalid DSN
	db, _ := gorm.Open(postgres.Open("invalid dsn"), &gorm.Config{})

	handler := NewHealthHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "DOWN", response["status"])

	details := response["details"].(map[string]interface{})
	assert.Equal(t, "unreachable", details["database"])
	assert.Equal(t, "disconnected", details["rabbitmq"])
}
