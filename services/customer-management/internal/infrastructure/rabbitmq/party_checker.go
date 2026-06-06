package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pkgrmq "tmf/pkg/rabbitmq"
)

const (
	partyExchange   = "tmf.party"
	queryPartyGet   = "query.party.get"
	partyRPCTimeout = 5 * time.Second
)

// rpcRequester is a narrow interface satisfied by *pkgrmq.RPCClient.
type rpcRequester interface {
	RequestToExchange(ctx context.Context, exchange, routingKey string, payload any) ([]byte, error)
}

// PartyCheckerRPC implements PartyChecker via a synchronous RPC call to party-management.
type PartyCheckerRPC struct {
	rpcClient rpcRequester
}

func NewPartyCheckerRPC(rpcClient *pkgrmq.RPCClient) *PartyCheckerRPC {
	return &PartyCheckerRPC{rpcClient: rpcClient}
}

// CheckParty returns an error if the party does not exist or is not in Active status.
func (c *PartyCheckerRPC) CheckParty(ctx context.Context, partyID string) error {
	ctx, cancel := context.WithTimeout(ctx, partyRPCTimeout)
	defer cancel()

	payload := map[string]string{"id": partyID}
	responseBytes, err := c.rpcClient.RequestToExchange(ctx, partyExchange, queryPartyGet, payload)
	if err != nil {
		return fmt.Errorf("party lookup failed: %w", err)
	}

	var party struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(responseBytes, &party); err != nil {
		return fmt.Errorf("invalid party response: %w", err)
	}

	switch party.Status {
	case "Deleted":
		return fmt.Errorf("party %s has been deleted and cannot be used for onboarding", partyID)
	case "DeletionPending":
		return fmt.Errorf("party %s is pending deletion and cannot be used for onboarding", partyID)
	case "":
		return fmt.Errorf("party %s not found", partyID)
	}

	return nil
}
