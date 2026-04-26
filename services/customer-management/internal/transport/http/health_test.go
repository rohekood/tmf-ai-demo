package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	os.Unsetenv("XDG_RUNTIME_DIR")
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		Env:          map[string]string{"POSTGRES_USER": "test", "POSTGRES_PASSWORD": "password", "POSTGRES_DB": "testdb"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(30 * time.Second),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}
	defer postgresC.Terminate(ctx)

	host, _ := postgresC.Host(ctx)
	port, _ := postgresC.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("host=%s user=test password=password dbname=testdb port=%s sslmode=disable TimeZone=UTC", host, port.Port())

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	sharedDB = db

	code := m.Run()
	os.Exit(code)
}

func TestMetricsHandler(t *testing.T) {
	handler := MetricsHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthHandler_ServeHTTP(t *testing.T) {
	db, _ := gorm.Open(postgres.Open("host=invalid"), &gorm.Config{})

	handler := NewHealthHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var response map[string]any
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "DOWN", response["status"])
}

type failWriter struct{}

func (f *failWriter) Header() http.Header {
	return make(http.Header)
}
func (f *failWriter) Write([]byte) (int, error) {
	return 0, assert.AnError
}
func (f *failWriter) WriteHeader(statusCode int) {}

func TestHealthHandler_ServeHTTP_UP(t *testing.T) {
	// Trigger the JSON encoding error using a custom ResponseWriter
	db, _ := gorm.Open(postgres.Open("host=invalid"), &gorm.Config{})
	handler := NewHealthHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := &failWriter{}

	// This should trigger the slog.Error("failed to encode health response", ...) branch
	handler.ServeHTTP(rr, req)
}

func TestHealthHandler_ServeHTTP_DBConnected(t *testing.T) {
	handler := NewHealthHandler(sharedDB, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code) // Still DOWN because rabbitConn is nil

	var response map[string]any
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	details := response["details"].(map[string]any)
	assert.Equal(t, "connected", details["database"])
}
