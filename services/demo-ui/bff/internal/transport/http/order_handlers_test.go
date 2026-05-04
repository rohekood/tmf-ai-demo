package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOrderHandler_CheckQualification(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewOrderHandler(mockClient)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			if routingKey != cmdQualEligibilityCheck {
				t.Errorf("Unexpected routing key: %s", routingKey)
			}
			return []byte(`{"sessionId":"sess1", "status":"qualified"}`), nil
		}

		req := httptest.NewRequest("POST", "/api/qualification/check", strings.NewReader(`{"partyId":"p1"}`))
		w := httptest.NewRecorder()
		handler.CheckQualification(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/qualification/check", strings.NewReader(`invalid json`))
		w := httptest.NewRecorder()
		handler.CheckQualification(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status BadRequest, got %v", w.Code)
		}
	})

	t.Run("RPCError", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return nil, errors.New("rpc error")
		}
		req := httptest.NewRequest("POST", "/api/qualification/check", strings.NewReader(`{"partyId":"p1"}`))
		w := httptest.NewRecorder()
		handler.CheckQualification(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status InternalServerError, got %v", w.Code)
		}
	})
}

func TestOrderHandler_GetQualificationSession(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewOrderHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			if routingKey != queryQualSessionGet {
				t.Errorf("Unexpected routing key: %s", routingKey)
			}
			p, ok := payload.(map[string]string)
			if !ok {
				t.Errorf("Expected payload to be map[string]string")
			}
			if p["sessionId"] != "sess1" {
				t.Errorf("Expected payload sessionId=sess1, got %v", p)
			}
			return []byte(`{"sessionId":"sess1", "status":"qualified"}`), nil
		}

		req := httptest.NewRequest("GET", "/api/qualification/session/sess1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
	})

	t.Run("RPCError_ReturnsUnprocessableEntity", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return nil, errors.New("not found or expired")
		}
		req := httptest.NewRequest("GET", "/api/qualification/session/sess1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("Expected status UnprocessableEntity, got %v", w.Code)
		}
		if !strings.Contains(w.Body.String(), "SESSION_EXPIRED") {
			t.Errorf("Expected body to contain SESSION_EXPIRED, got %s", w.Body.String())
		}
	})

	t.Run("EmptySessionId_ReturnsBadRequest", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/qualification/session/", nil)
		w := httptest.NewRecorder()
		// Call the handler directly with an empty path value to test the validation
		handler.GetQualificationSession(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status BadRequest for empty session ID, got %v", w.Code)
		}
	})
}

func TestOrderHandler_AddCartItem(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewOrderHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return []byte(`{"cartId":"cart1", "items":[]}`), nil
		}

		req := httptest.NewRequest("POST", "/api/cart/items", strings.NewReader(`{"cartId":"cart1", "offerId":"o1"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/cart/items", strings.NewReader(`invalid json`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status BadRequest, got %v", w.Code)
		}
	})

	t.Run("RPCError", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return nil, errors.New("rpc error")
		}
		req := httptest.NewRequest("POST", "/api/cart/items", strings.NewReader(`{"offerId":"o1"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("Expected status UnprocessableEntity, got %v", w.Code)
		}
	})
}

func TestOrderHandler_GetCart(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewOrderHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return []byte(`{"id":"cart1"}`), nil
		}

		req := httptest.NewRequest("GET", "/api/cart/cart1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
	})

	t.Run("RPCError", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return nil, errors.New("rpc error")
		}
		req := httptest.NewRequest("GET", "/api/cart/cart1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status InternalServerError, got %v", w.Code)
		}
	})
}

func TestOrderHandler_RemoveCartItem(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewOrderHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return []byte(`{}`), nil
		}

		req := httptest.NewRequest("DELETE", "/api/cart/cart1/items/item1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status NoContent, got %v", w.Code)
		}
	})

	t.Run("RPCError", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return nil, errors.New("rpc error")
		}
		req := httptest.NewRequest("DELETE", "/api/cart/cart1/items/item1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status InternalServerError, got %v", w.Code)
		}
	})
}

func TestOrderHandler_Checkout(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewOrderHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return []byte(`{"sagaId":"saga1"}`), nil
		}

		req := httptest.NewRequest("POST", "/api/orders/checkout", strings.NewReader(`{"cartId":"cart1"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("Expected status Accepted, got %v", w.Code)
		}
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/orders/checkout", strings.NewReader(`invalid json`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status BadRequest, got %v", w.Code)
		}
	})

	t.Run("RPCError", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return nil, errors.New("rpc error")
		}
		req := httptest.NewRequest("POST", "/api/orders/checkout", strings.NewReader(`{"cartId":"cart1"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status InternalServerError, got %v", w.Code)
		}
	})
}

func TestOrderHandler_GetSagaStatus(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewOrderHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return []byte(`{"id":"saga1", "status":"COMPLETED"}`), nil
		}

		req := httptest.NewRequest("GET", "/api/orders/saga/saga1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
	})

	t.Run("RPCError", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			return nil, errors.New("rpc error")
		}
		req := httptest.NewRequest("GET", "/api/orders/saga/saga1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status InternalServerError, got %v", w.Code)
		}
	})
}
