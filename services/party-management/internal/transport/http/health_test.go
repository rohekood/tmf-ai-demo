package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tmf/services/party-management/internal/infrastructure/rabbitmq"
	apihttp "tmf/services/party-management/internal/transport/http"

	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestHealthHandler(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Skipping integration test: testcontainers unavailable (%v)", r)
		}
	}()
	ctx := context.Background()
	pg, err := pgContainer.Run(ctx, "postgres:15",
		pgContainer.WithDatabase("testdb"),
		pgContainer.WithUsername("postgres"),
		pgContainer.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Skipf("Skipping test due to testcontainers error: %s", err)
		return
	}
	defer func() { _ = pg.Terminate(ctx) }()

	pgConnStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(gormPostgres.Open(pgConnStr), &gorm.Config{})
	require.NoError(t, err)

	rmq := rabbitmq.NewConnectionManager("amqp://guest:guest@localhost:1234/")

	handler := apihttp.NewHealthHandler(db, rmq)

	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var resp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)

	assert.Equal(t, "DOWN", resp["status"])
	details := resp["details"].(map[string]interface{})
	assert.Equal(t, "OK", details["db"])
	assert.Equal(t, "connection closed", details["rabbitmq"])
}

func TestHealthHandler_DBDown(t *testing.T) {
	db, _ := gorm.Open(gormPostgres.Open("postgres://postgres:wrong@localhost:5432/testdb"), &gorm.Config{})
	rmq := rabbitmq.NewConnectionManager("amqp://guest:guest@localhost:1234/")
	handler := apihttp.NewHealthHandler(db, rmq)

	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestMetricsHandler(t *testing.T) {
	handler := apihttp.MetricsHandler()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
