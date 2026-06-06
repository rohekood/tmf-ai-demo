package rabbitmq

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockRPCRequester struct {
	response []byte
	err      error
}

func (m *mockRPCRequester) RequestToExchange(_ context.Context, _, _ string, _ any) ([]byte, error) {
	return m.response, m.err
}

func TestCustomerCheckerRPC_HasCustomers_Found(t *testing.T) {
	checker := &CustomerCheckerRPC{
		rpcClient: &mockRPCRequester{response: []byte(`[{"id":"c1"}]`)},
	}
	has, err := checker.HasCustomers(context.Background(), "p1")
	assert.NoError(t, err)
	assert.True(t, has)
}

func TestCustomerCheckerRPC_HasCustomers_None(t *testing.T) {
	checker := &CustomerCheckerRPC{
		rpcClient: &mockRPCRequester{response: []byte(`[]`)},
	}
	has, err := checker.HasCustomers(context.Background(), "p1")
	assert.NoError(t, err)
	assert.False(t, has)
}

func TestCustomerCheckerRPC_HasCustomers_RPCError(t *testing.T) {
	checker := &CustomerCheckerRPC{
		rpcClient: &mockRPCRequester{err: errors.New("timeout")},
	}
	has, err := checker.HasCustomers(context.Background(), "p1")
	assert.Error(t, err)
	assert.False(t, has)
}

func TestCustomerCheckerRPC_HasCustomers_NonArrayResponse(t *testing.T) {
	// An error object from the backend is treated as no customers
	checker := &CustomerCheckerRPC{
		rpcClient: &mockRPCRequester{response: []byte(`{"error":"some error"}`)},
	}
	has, err := checker.HasCustomers(context.Background(), "p1")
	assert.NoError(t, err)
	assert.False(t, has)
}

func TestNewCustomerCheckerRPC_NotNil(t *testing.T) {
	// Constructor must return a non-nil checker even with a nil client.
	checker := NewCustomerCheckerRPC(nil)
	assert.NotNil(t, checker)
}
