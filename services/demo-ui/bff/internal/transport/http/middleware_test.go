package http

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChain(t *testing.T) {
	// Middleware that appends a string to a slice
	createMiddleware := func(name string, called *[]string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				*called = append(*called, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	var called []string
	m1 := createMiddleware("m1", &called)
	m2 := createMiddleware("m2", &called)
	m3 := createMiddleware("m3", &called)

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = append(called, "final")
	})

	// Chain(h, m1, m2, m3) means m1(m2(m3(h)))
	// Execution order: m1 -> m2 -> m3 -> final
	chained := Chain(finalHandler, m1, m2, m3)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	chained.ServeHTTP(w, req)

	expected := []string{"m1", "m2", "m3", "final"}
	if len(called) != len(expected) {
		t.Errorf("Expected %d calls, got %d", len(expected), len(called))
	}
	for i, v := range expected {
		if called[i] != v {
			t.Errorf("Expected call %d to be %s, got %s", i, v, called[i])
		}
	}
}

func TestLoggerMiddleware(t *testing.T) {
	// We can't easily assert on slog output without a custom handler,
	// but we can verify it calls the next handler and doesn't panic.

	// Optional: Capture log output
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	oldDefault := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldDefault)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
	})

	handler := LoggerMiddleware(next)

	req := httptest.NewRequest("GET", "/test-path", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !nextCalled {
		t.Error("Next handler was not called")
	}
	if w.Code != http.StatusTeapot {
		t.Errorf("Expected status %d, got %d", http.StatusTeapot, w.Code)
	}

	// Verify log contained key info
	logOutput := buf.String()
	if !contains(logOutput, "request completed") {
		t.Error("Log output missing 'request completed'")
	}
	if !contains(logOutput, "/test-path") {
		t.Error("Log output missing path")
	}
	if !contains(logOutput, "418") { // Teapot code
		t.Error("Log output missing status code")
	}
}

func TestRecovererMiddleware(t *testing.T) {
	// Optional: Capture log output to suppress noise during test
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	oldDefault := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldDefault)

	panickingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("oops")
	})

	handler := RecovererMiddleware(panickingHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Should not panic, but recover
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
	if w.Body.String() != "Internal Server Error\n" {
		t.Errorf("Unexpected body: %q", w.Body.String())
	}
}

func TestCORSMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware(next)

	t.Run("Allowed Origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		checkCORSHeaders(t, w, "http://localhost:5173")
	})

	t.Run("Another Allowed Origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "http://localhost")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		checkCORSHeaders(t, w, "http://localhost")
	})

	t.Run("Disallowed Origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", "http://evil.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if val := w.Header().Get("Access-Control-Allow-Origin"); val != "" {
			t.Errorf("Expected empty Allow-Origin for disallowed origin, got %s", val)
		}
	})

	t.Run("Preflight Request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
		}
		checkCORSHeaders(t, w, "http://localhost:5173")
	})
}

func checkCORSHeaders(t *testing.T, w *httptest.ResponseRecorder, expectedOrigin string) {
	headers := w.Header()
	if val := headers.Get("Access-Control-Allow-Origin"); val != expectedOrigin {
		t.Errorf("Expected Allow-Origin: %s, got %s", expectedOrigin, val)
	}
	if val := headers.Get("Access-Control-Allow-Credentials"); val != "true" {
		t.Errorf("Unexpected Allow-Credentials: %s", val)
	}
	// Check for 'user' in Allow-Headers as we added it
	if val := headers.Get("Access-Control-Allow-Headers"); !contains(val, "user") {
		t.Errorf("Allow-Headers missing 'user': %s", val)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || searchString(s, substr))
}

func searchString(s, substr string) bool {
	// Simple containment check
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
