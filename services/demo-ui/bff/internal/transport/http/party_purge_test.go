package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPartyHandler_PurgeParty(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("Success_NoContent", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			if routingKey != cmdPartyPurge {
				t.Errorf("expected routing key %q, got %q", cmdPartyPurge, routingKey)
			}
			return []byte(`{"status":"purged","id":"p1"}`), nil
		}
		req := httptest.NewRequest("DELETE", "/api/parties/p1/purge", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", w.Code)
		}
	})

	t.Run("RPCError_500", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return nil, errors.New("connection lost")
		}
		req := httptest.NewRequest("DELETE", "/api/parties/p2/purge", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("BackendError_422", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return []byte(`{"error":"party must be in Deleted status"}`), nil
		}
		req := httptest.NewRequest("DELETE", "/api/parties/p3/purge", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected 422, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Deleted status") {
			t.Errorf("expected error message in body, got %q", w.Body.String())
		}
	})
}

func TestPartyHandler_DeleteParty_ErrorBodyReturns409(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
		return []byte(`{"error":"party has linked customers and cannot be deleted"}`), nil
	}
	req := httptest.NewRequest("DELETE", "/api/parties/p1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "linked customers") {
		t.Errorf("expected error message in body, got %q", w.Body.String())
	}
}

func TestPartyHandler_PurgeParty_EmptyID(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewPartyHandler(mockClient)
	req := httptest.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	handler.PurgeParty(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPartyHandler_DeleteParty_EmptyID(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewPartyHandler(mockClient)
	req := httptest.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	handler.DeleteParty(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPartyHandler_SearchParties_StatusParam(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	var capturedPayload map[string]*string
	mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
		if m, ok := payload.(map[string]*string); ok {
			capturedPayload = m
		}
		return []byte(`[]`), nil
	}
	req := httptest.NewRequest("GET", "/api/parties?status=Active", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedPayload["status"] == nil || *capturedPayload["status"] != "Active" {
		t.Errorf("expected status=Active in payload, got %v", capturedPayload)
	}
}
