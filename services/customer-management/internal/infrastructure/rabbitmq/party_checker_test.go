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

func TestPartyCheckerRPC_CheckParty_Active(t *testing.T) {
	checker := &PartyCheckerRPC{
		rpcClient: &mockRPCRequester{response: []byte(`{"id":"p1","status":"Active"}`)},
	}
	err := checker.CheckParty(context.Background(), "p1")
	assert.NoError(t, err)
}

func TestPartyCheckerRPC_CheckParty_Deleted(t *testing.T) {
	checker := &PartyCheckerRPC{
		rpcClient: &mockRPCRequester{response: []byte(`{"id":"p1","status":"Deleted"}`)},
	}
	err := checker.CheckParty(context.Background(), "p1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deleted")
}

func TestPartyCheckerRPC_CheckParty_DeletionPending(t *testing.T) {
	checker := &PartyCheckerRPC{
		rpcClient: &mockRPCRequester{response: []byte(`{"id":"p1","status":"DeletionPending"}`)},
	}
	err := checker.CheckParty(context.Background(), "p1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pending deletion")
}

func TestPartyCheckerRPC_CheckParty_NotFound(t *testing.T) {
	checker := &PartyCheckerRPC{
		rpcClient: &mockRPCRequester{response: []byte(`{"status":""}`)},
	}
	err := checker.CheckParty(context.Background(), "p1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPartyCheckerRPC_CheckParty_RPCError(t *testing.T) {
	checker := &PartyCheckerRPC{
		rpcClient: &mockRPCRequester{err: errors.New("connection refused")},
	}
	err := checker.CheckParty(context.Background(), "p1")
	assert.Error(t, err)
}

func TestPartyCheckerRPC_CheckParty_InvalidJSON(t *testing.T) {
	checker := &PartyCheckerRPC{
		rpcClient: &mockRPCRequester{response: []byte(`not json`)},
	}
	err := checker.CheckParty(context.Background(), "p1")
	assert.Error(t, err)
}

func TestNewPartyCheckerRPC_NotNil(t *testing.T) {
	// Constructor must return a non-nil checker even with a nil client.
	checker := NewPartyCheckerRPC(nil)
	assert.NotNil(t, checker)
}
