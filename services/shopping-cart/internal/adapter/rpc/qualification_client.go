package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tmf/pkg/rabbitmq"
)

// RPCRequester interface for making RPC requests (for testing)
type RPCRequester interface {
	RequestWithHeaders(ctx context.Context, exchange, routingKey string, request interface{}, headers map[string]interface{}) ([]byte, error)
}

// QualificationClient handles RPC calls to the qualification service
type QualificationClient struct {
	rpcClient RPCRequester
}

// NewQualificationClient creates a new qualification RPC client
func NewQualificationClient(rpcClient *rabbitmq.RPCClient) *QualificationClient {
	return &QualificationClient{
		rpcClient: rpcClient,
	}
}

// NewQualificationClientForTest creates a client with a custom RPCRequester (for testing)
func NewQualificationClientForTest(rpcRequester RPCRequester) *QualificationClient {
	return &QualificationClient{
		rpcClient: rpcRequester,
	}
}

// QualificationSession represents a qualification session with pricing
type QualificationSession struct {
	ID                string                       `json:"id"`
	CustomerID        string                       `json:"customerId"`
	QualifiedOffering []QualifiedOfferingWithPrice `json:"qualifiedOffers"`
	ExpiresAt         time.Time                    `json:"expiresAt"`
	Status            string                       `json:"status"`
}

// QualifiedOfferingWithPrice represents an offering with its calculated price
type QualifiedOfferingWithPrice struct {
	OfferingID string    `json:"offeringId"`
	PriceInfo  PriceInfo `json:"price"`
	Eligible   string    `json:"eligibility"` // Changed to string to match "QUALIFIED"
}

type PriceInfo struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// GetSession retrieves a qualification session by ID
func (c *QualificationClient) GetSession(ctx context.Context, sessionID string) (*QualificationSession, error) {
	request := map[string]string{
		"sessionId": sessionID,
	}

	// Call RPC: query.qual.session.get
	respBytes, err := c.rpcClient.RequestWithHeaders(
		ctx,
		"ex.domain.market",
		"query.qual.session.get",
		request,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get qualification session: %w", err)
	}

	var session QualificationSession
	if err := json.Unmarshal(respBytes, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session response: %w", err)
	}

	// Validate session is not expired
	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, fmt.Errorf("qualification session has expired")
	}

	return &session, nil
}

// GetOfferingPrice implements the interface method for extracting offering price
func (s *QualificationSession) GetOfferingPrice(offeringID string) (price float64, currency string, eligible bool, found bool) {
	for _, offering := range s.QualifiedOffering {
		if offering.OfferingID == offeringID {
			isEligible := offering.Eligible == "QUALIFIED" || offering.Eligible == "true"
			return offering.PriceInfo.Amount, offering.PriceInfo.Currency, isEligible, true
		}
	}
	return 0, "", false, false
}

// GetPriceForOfferingWrapper is a wrapper method that returns an interface type
func (c *QualificationClient) GetPriceForOfferingWrapper(sessionInterface interface{}, offeringID string) (interface{}, error) {
	price, currency, err := c.GetPriceForOffering(sessionInterface, offeringID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"price":    price,
		"currency": currency,
	}, nil
}

// GetPriceForOffering extracts the price for a specific offering from the session
// Deprecated: Use GetOfferingPrice on the session directly
func (c *QualificationClient) GetPriceForOffering(sessionInterface interface{}, offeringID string) (float64, string, error) {
	session, ok := sessionInterface.(*QualificationSession)
	if !ok {
		return 0, "", fmt.Errorf("invalid session type")
	}

	for _, offering := range session.QualifiedOffering {
		if offering.OfferingID == offeringID {
			if offering.Eligible != "QUALIFIED" && offering.Eligible != "true" {
				return 0, "", fmt.Errorf("offering %s is not eligible in this session", offeringID)
			}
			return offering.PriceInfo.Amount, offering.PriceInfo.Currency, nil
		}
	}
	return 0, "", fmt.Errorf("offering %s not found in qualification session", offeringID)
}
