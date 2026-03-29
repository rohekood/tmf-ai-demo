package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tmf/pkg/rabbitmq"
)

func TestHandler_SearchCustomers(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			if exchange != "tmf.customer" || routingKey != "query.customer.search" {
				t.Errorf("Unexpected exchange or routing key: %s, %s", exchange, routingKey)
			}
			return []byte(`[{"id":"1", "name":"Customer 1"}]`), nil
		}

		req := httptest.NewRequest("GET", "/api/customers?name=test", nil)
		w := httptest.NewRecorder()

		handler.SearchCustomers(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Customer 1") {
			t.Errorf("Expected response to contain 'Customer 1', got %s", w.Body.String())
		}
	})

	t.Run("RPC Error", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return nil, errors.New("RPC failed")
		}

		req := httptest.NewRequest("GET", "/api/customers", nil)
		w := httptest.NewRecorder()

		handler.SearchCustomers(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status InternalServerError, got %v", w.Code)
		}
	})
}

func TestHandler_GetCustomer(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/customers/{id}", handler.GetCustomer)

	t.Run("Success", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			return []byte(`{"id":"1", "name":"Customer 1"}`), nil
		}

		req := httptest.NewRequest("GET", "/api/customers/123", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}
	})
}

func TestHandler_CreateCustomer_DerivesName(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)

	mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
		if routingKey == "query.party.get" {
			return []byte(`{"id":"p1", "@type":"Individual", "givenName":"John", "familyName":"Doe"}`), nil
		}
		if routingKey == "cmd.customer.onboard" {
			p := payload.(map[string]interface{})
			if p["name"] != "John Doe" {
				t.Errorf("Expected name to be derived as 'John Doe', got %v", p["name"])
			}
			return []byte(`{"id":"c1", "name":"John Doe"}`), nil
		}
		return nil, errors.New("unexpected call")
	}

	body := `{"partyId":"p1"}` // No name provided
	req := httptest.NewRequest("POST", "/api/customers", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateCustomer(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status Created, got %v", w.Code)
	}
}

func TestHandler_CreateCustomer_RespectsProvidedName(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)

	mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
		if routingKey == "cmd.customer.onboard" {
			p := payload.(map[string]interface{})
			if p["name"] != "Custom Name" {
				t.Errorf("Expected name to be 'Custom Name', got %v", p["name"])
			}
			return []byte(`{"id":"c1", "name":"Custom Name"}`), nil
		}
		return nil, errors.New("unexpected call")
	}

	body := `{"partyId":"p1", "name":"Custom Name"}`
	req := httptest.NewRequest("POST", "/api/customers", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateCustomer(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status Created, got %v", w.Code)
	}
}

func TestHandler_GetCustomer_EnrichesPartyDetails(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)

	mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
		if routingKey == "query.customer.get" {
			return []byte(`{"id":"c1", "name":"Cust", "partyId":"p1"}`), nil
		}
		if routingKey == "query.party.get" {
			return []byte(`{"id":"p1", "@type":"Individual", "givenName":"John", "familyName":"Doe"}`), nil
		}
		return nil, errors.New("unexpected call: " + routingKey)
	}

	req := httptest.NewRequest("GET", "/api/customers/c1", nil)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/customers/{id}", handler.GetCustomer)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["partyName"] != "John Doe" {
		t.Errorf("Expected partyName 'John Doe', got %v", response["partyName"])
	}
	if response["partyType"] != "Individual" {
		t.Errorf("Expected partyType 'Individual', got %v", response["partyType"])
	}
}

func TestCustomerHandler_PartyEnrichmentErrors(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("GetCustomer_PartyErr", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			if routingKey == "query.customer.get" {
				return []byte(`{"id":"c1", "partyId":"p1"}`), nil
			}
			if routingKey == "query.party.get" {
				return nil, errors.New("party rpc error")
			}
			return nil, nil
		}
		req := httptest.NewRequest("GET", "/api/customers/c1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200 got %d", w.Code)
		}
	})

	t.Run("CreateCustomer_PartyErr", func(t *testing.T) {
		mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
			if routingKey == "cmd.customer.onboard" {
				return []byte(`{"id":"c1", "partyId":"p1"}`), nil
			}
			if routingKey == "query.party.get" {
				return nil, errors.New("party rpc error")
			}
			return nil, nil
		}
		req := httptest.NewRequest("POST", "/api/customers", strings.NewReader(`{"partyId":"p1"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201 got %d", w.Code)
		}
	})
}

func TestHandler_GetHeaders_UsesTypedUserKeys(t *testing.T) {
	t.Run("context user key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer token")
		req = req.WithContext(context.WithValue(req.Context(), rabbitmq.ContextKeyUser, "typed-user"))

		headers := getHeaders(req)
		if headers["Authorization"] != "Bearer token" {
			t.Fatalf("expected Authorization header to be forwarded")
		}
		if headers["user"] != "typed-user" {
			t.Fatalf("expected typed user to be forwarded, got %v", headers["user"])
		}
	})

	t.Run("header user key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), rabbitmq.Key(rabbitmq.HeaderUser), "header-user"))

		headers := getHeaders(req)
		if headers["user"] != "header-user" {
			t.Fatalf("expected header user to be forwarded, got %v", headers["user"])
		}
	})
}

func TestHandler_GetHeaders_WithoutContextValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	headers := getHeaders(req)
	if len(headers) != 0 {
		t.Fatalf("expected no forwarded headers, got %v", headers)
	}
}

func TestHandler_CustomerIDRequired(t *testing.T) {
	handler := NewHandler(&MockRPCClient{}, nil)

	tests := []struct {
		name string
		call func(http.ResponseWriter, *http.Request)
	}{
		{name: "GetCustomer", call: handler.GetCustomer},
		{name: "UpdateCustomer", call: handler.UpdateCustomer},
		{name: "DeleteCustomer", call: handler.DeleteCustomer},
		{name: "LogInteraction", call: handler.LogInteraction},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			tc.call(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestHandler_CustomerMutationSuccess(t *testing.T) {
	mockClient := &MockRPCClient{}
	handler := NewHandler(mockClient, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	mockClient.CallRPCFunc = func(ctx context.Context, exchange, routingKey string, payload interface{}, headers map[string]interface{}) ([]byte, error) {
		if exchange != customerExchange {
			return nil, errors.New("unexpected exchange")
		}
		switch routingKey {
		case cmdCustomerUpdate:
			if payload.(map[string]interface{})["id"] != "c1" {
				t.Fatalf("expected update payload to include id c1")
			}
			if headers["user"] != "user-1" {
				t.Fatalf("expected user header to be forwarded")
			}
			return []byte(`{"id":"c1","status":"updated"}`), nil
		case cmdCustomerDelete:
			if payload.(map[string]string)["id"] != "c1" {
				t.Fatalf("expected delete payload to include id c1")
			}
			return []byte(`{}`), nil
		case cmdCustomerLogInteraction:
			if payload.(map[string]interface{})["customerId"] != "c1" {
				t.Fatalf("expected interaction payload to include customerId c1")
			}
			return []byte(`{"id":"i1"}`), nil
		default:
			return nil, errors.New("unexpected routing key: " + routingKey)
		}
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/customers/c1", strings.NewReader(`{"status":"active"}`))
	updateReq = updateReq.WithContext(context.WithValue(updateReq.Context(), rabbitmq.ContextKeyUser, "user-1"))
	updateResp := httptest.NewRecorder()
	mux.ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d", updateResp.Code)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/customers/c1", nil)
	deleteResp := httptest.NewRecorder()
	mux.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d", deleteResp.Code)
	}

	interactionReq := httptest.NewRequest(http.MethodPost, "/api/customers/c1/interactions", strings.NewReader(`{"type":"note"}`))
	interactionResp := httptest.NewRecorder()
	mux.ServeHTTP(interactionResp, interactionReq)
	if interactionResp.Code != http.StatusCreated {
		t.Fatalf("expected interaction status 201, got %d", interactionResp.Code)
	}
}
