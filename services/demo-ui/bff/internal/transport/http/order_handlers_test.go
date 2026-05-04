package http

import (
	"context"
	"encoding/json"
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

	t.Run("Success_Returns202WithSagaIdAndCartId", func(t *testing.T) {
		var publishedRoutingKey string
		var publishedPayload any
		mockClient.PublishCommandFunc = func(ctx context.Context, exchange, routingKey string, payload any) error {
			publishedRoutingKey = routingKey
			publishedPayload = payload
			return nil
		}

		req := httptest.NewRequest("POST", "/api/orders/checkout", strings.NewReader(`{"cartId":"cart1","customerId":"cust1"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("Expected status Accepted, got %v", w.Code)
		}
		if publishedRoutingKey != cmdOrderCheckoutSubmit {
			t.Errorf("Expected routing key %s, got %s", cmdOrderCheckoutSubmit, publishedRoutingKey)
		}

		// Response must contain sagaId, status="PENDING", cartId
		var resp map[string]any
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if resp["status"] != "PENDING" {
			t.Errorf("Expected status PENDING, got %v", resp["status"])
		}
		if resp["cartId"] != "cart1" {
			t.Errorf("Expected cartId cart1, got %v", resp["cartId"])
		}
		sagaID, ok := resp["sagaId"].(string)
		if !ok || sagaID == "" {
			t.Errorf("Expected non-empty sagaId, got %v", resp["sagaId"])
		}

		// Verify sagaId was forwarded in the published payload
		payloadMap, ok := publishedPayload.(map[string]any)
		if !ok {
			t.Fatalf("Published payload is not a map: %T", publishedPayload)
		}
		if payloadMap["sagaId"] != sagaID {
			t.Errorf("Expected sagaId %s in published payload, got %v", sagaID, payloadMap["sagaId"])
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

	t.Run("MissingCartId_Returns400", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/orders/checkout", strings.NewReader(`{"customerId":"cust1"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status BadRequest, got %v", w.Code)
		}
	})

	t.Run("PublishError_Returns500", func(t *testing.T) {
		mockClient.PublishCommandFunc = func(ctx context.Context, exchange, routingKey string, payload any) error {
			return errors.New("publish error")
		}
		req := httptest.NewRequest("POST", "/api/orders/checkout", strings.NewReader(`{"cartId":"cart1","customerId":"cust1"}`))
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

	t.Run("Success_ReturnsSagaStatus", func(t *testing.T) {
		var capturedRoutingKey string
		var capturedPayload any
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
			capturedRoutingKey = routingKey
			capturedPayload = payload
			return []byte(`{"id":"saga1","status":"COMPLETED","currentStep":"ORDER_CREATION"}`), nil
		}

		req := httptest.NewRequest("GET", "/api/orders/saga/saga1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
		if capturedRoutingKey != queryPocvSagaGet {
			t.Errorf("Expected routing key %s, got %s", queryPocvSagaGet, capturedRoutingKey)
		}

		// Verify payload only sends id (not duplicate sagaId key)
		payloadMap, ok := capturedPayload.(map[string]string)
		if !ok {
			t.Fatalf("Unexpected payload type: %T", capturedPayload)
		}
		if payloadMap["id"] != "saga1" {
			t.Errorf("Expected id saga1, got %v", payloadMap["id"])
		}
		if _, hasSagaId := payloadMap["sagaId"]; hasSagaId {
			t.Error("Payload should not contain duplicate sagaId key")
		}
	})

	t.Run("RPCError_Returns500", func(t *testing.T) {
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
