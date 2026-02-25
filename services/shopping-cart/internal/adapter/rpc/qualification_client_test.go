package rpc_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"tmf/services/shopping-cart/internal/adapter/rpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRPCClient mocks the RPCRequester interface
type MockRPCClient struct {
	mock.Mock
}

func (m *MockRPCClient) RequestWithHeaders(ctx context.Context, exchange, routingKey string, request interface{}, headers map[string]interface{}) ([]byte, error) {
	args := m.Called(ctx, exchange, routingKey, request, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestQualificationClient_GetSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Should retrieve valid session successfully", func(t *testing.T) {
		// Arrange
		mockRPC := new(MockRPCClient)
		client := &rpc.QualificationClient{}
		// Use reflection or create a test constructor
		client = &rpc.QualificationClient{}
		// Directly set the private field for testing (not ideal but works)
		// Better: create NewQualificationClientWithRequester for testing

		// Actually, let's use a better approach - create test helper
		client = rpc.NewQualificationClientForTest(mockRPC)

		sessionID := "session-123"
		expectedSession := rpc.QualificationSession{
			ID:         sessionID,
			CustomerID: "customer-456",
			QualifiedOffering: []rpc.QualifiedOfferingWithPrice{
				{
					OfferingID: "offering-1",
					PriceInfo: rpc.PriceInfo{Amount: 80.0, Currency: "EUR"},
					
					Eligible:   "QUALIFIED",
				},
			},
			ExpiresAt: time.Now().Add(1 * time.Hour),
			Status:    "active",
		}

		responseBytes, _ := json.Marshal(expectedSession)
		mockRPC.On("RequestWithHeaders", ctx, "ex.domain.market", "query.qual.session.get",
			map[string]string{"sessionId": sessionID}, mock.Anything).
			Return(responseBytes, nil)

		// Act
		session, err := client.GetSession(ctx, sessionID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, sessionID, session.ID)
		assert.Equal(t, "customer-456", session.CustomerID)
		assert.Len(t, session.QualifiedOffering, 1)
		mockRPC.AssertExpectations(t)
	})

	t.Run("Should fail when RPC call fails", func(t *testing.T) {
		// Arrange
		mockRPC := new(MockRPCClient)
		client := rpc.NewQualificationClientForTest(mockRPC)

		sessionID := "session-123"
		mockRPC.On("RequestWithHeaders", ctx, "ex.domain.market", "query.qual.session.get",
			mock.Anything, mock.Anything).
			Return(nil, errors.New("RPC timeout"))

		// Act
		session, err := client.GetSession(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "failed to get qualification session")
		mockRPC.AssertExpectations(t)
	})

	t.Run("Should fail when session is expired", func(t *testing.T) {
		// Arrange
		mockRPC := new(MockRPCClient)
		client := rpc.NewQualificationClientForTest(mockRPC)

		sessionID := "session-123"
		expiredSession := rpc.QualificationSession{
			ID:        sessionID,
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
			Status:    "active",
		}

		responseBytes, _ := json.Marshal(expiredSession)
		mockRPC.On("RequestWithHeaders", ctx, "ex.domain.market", "query.qual.session.get",
			mock.Anything, mock.Anything).
			Return(responseBytes, nil)

		// Act
		session, err := client.GetSession(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "expired")
		mockRPC.AssertExpectations(t)
	})

	t.Run("Should fail when response is invalid JSON", func(t *testing.T) {
		// Arrange
		mockRPC := new(MockRPCClient)
		client := rpc.NewQualificationClientForTest(mockRPC)

		sessionID := "session-123"
		mockRPC.On("RequestWithHeaders", ctx, "ex.domain.market", "query.qual.session.get",
			mock.Anything, mock.Anything).
			Return([]byte("invalid json"), nil)

		// Act
		session, err := client.GetSession(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "failed to unmarshal")
		mockRPC.AssertExpectations(t)
	})
}

func TestQualificationSession_GetOfferingPrice(t *testing.T) {
	t.Run("Should return price for eligible offering", func(t *testing.T) {
		// Arrange
		session := &rpc.QualificationSession{
			QualifiedOffering: []rpc.QualifiedOfferingWithPrice{
				{
					OfferingID: "offering-1",
					PriceInfo: rpc.PriceInfo{Amount: 80.0, Currency: "EUR"},
					
					Eligible:   "QUALIFIED",
				},
				{
					OfferingID: "offering-2",
					PriceInfo: rpc.PriceInfo{Amount: 90.0, Currency: "EUR"},
					
					Eligible:   "QUALIFIED",
				},
			},
		}

		// Act
		price, currency, eligible, found := session.GetOfferingPrice("offering-1")

		// Assert
		assert.True(t, found)
		assert.True(t, eligible)
		assert.Equal(t, 80.0, price)
		assert.Equal(t, "EUR", currency)
	})

	t.Run("Should return not eligible for ineligible offering", func(t *testing.T) {
		// Arrange
		session := &rpc.QualificationSession{
			QualifiedOffering: []rpc.QualifiedOfferingWithPrice{
				{
					OfferingID: "offering-1",
					PriceInfo: rpc.PriceInfo{Amount: 100.0, Currency: "EUR"},
					
					Eligible:   "FALSE",
				},
			},
		}

		// Act
		price, currency, eligible, found := session.GetOfferingPrice("offering-1")

		// Assert
		assert.True(t, found)
		assert.False(t, eligible)
		assert.Equal(t, 100.0, price)
		assert.Equal(t, "EUR", currency)
	})

	t.Run("Should return not found for missing offering", func(t *testing.T) {
		// Arrange
		session := &rpc.QualificationSession{
			QualifiedOffering: []rpc.QualifiedOfferingWithPrice{
				{
					OfferingID: "offering-1",
					PriceInfo: rpc.PriceInfo{Amount: 80.0, Currency: "EUR"},
					
					Eligible:   "QUALIFIED",
				},
			},
		}

		// Act
		price, currency, eligible, found := session.GetOfferingPrice("offering-999")

		// Assert
		assert.False(t, found)
		assert.False(t, eligible)
		assert.Equal(t, 0.0, price)
		assert.Equal(t, "", currency)
	})
}

func TestQualificationClientAdapter(t *testing.T) {
	ctx := context.Background()

	t.Run("Should adapt GetSession correctly", func(t *testing.T) {
		// Arrange
		mockRPC := new(MockRPCClient)
		client := rpc.NewQualificationClientForTest(mockRPC)
		adapter := rpc.NewQualificationClientAdapter(client)

		sessionID := "session-123"
		expectedSession := rpc.QualificationSession{
			ID:         sessionID,
			CustomerID: "customer-456",
			ExpiresAt:  time.Now().Add(1 * time.Hour),
		}

		responseBytes, _ := json.Marshal(expectedSession)
		mockRPC.On("RequestWithHeaders", ctx, "ex.domain.market", "query.qual.session.get",
			mock.Anything, mock.Anything).
			Return(responseBytes, nil)

		// Act
		session, err := adapter.GetSession(ctx, sessionID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, session)

		// Verify it implements the interface
		price, currency, eligible, found := session.GetOfferingPrice("test")
		assert.False(t, found) // No offerings in this session
		assert.Equal(t, 0.0, price)
		assert.Equal(t, "", currency)
		assert.False(t, eligible)

		mockRPC.AssertExpectations(t)
	})
}

func TestQualificationClient_GetPriceForOffering(t *testing.T) {
	client := rpc.NewQualificationClientForTest(nil)
	
	t.Run("Should get price for offering", func(t *testing.T) {
		session := &rpc.QualificationSession{
			QualifiedOffering: []rpc.QualifiedOfferingWithPrice{
				{
					OfferingID: "off-1",
					PriceInfo: rpc.PriceInfo{Amount: 50.0, Currency: "USD"},
					Eligible: "QUALIFIED",
				},
			},
		}

		price, cur, err := client.GetPriceForOffering(session, "off-1")
		assert.NoError(t, err)
		assert.Equal(t, 50.0, price)
		assert.Equal(t, "USD", cur)
	})

	t.Run("Should error on invalid session type", func(t *testing.T) {
		_, _, err := client.GetPriceForOffering("invalid", "off-1")
		assert.Error(t, err)
	})

	t.Run("Should error on not found", func(t *testing.T) {
		session := &rpc.QualificationSession{}
		_, _, err := client.GetPriceForOffering(session, "off-1")
		assert.Error(t, err)
	})

	t.Run("Should error on ineligible", func(t *testing.T) {
		session := &rpc.QualificationSession{
			QualifiedOffering: []rpc.QualifiedOfferingWithPrice{
				{OfferingID: "off-1", Eligible: "FALSE"},
			},
		}
		_, _, err := client.GetPriceForOffering(session, "off-1")
		assert.Error(t, err)
	})
}

func TestQualificationClient_GetPriceForOfferingWrapper(t *testing.T) {
	client := rpc.NewQualificationClientForTest(nil)
	
	t.Run("Should return map on success", func(t *testing.T) {
		session := &rpc.QualificationSession{
			QualifiedOffering: []rpc.QualifiedOfferingWithPrice{
				{
					OfferingID: "off-1",
					PriceInfo: rpc.PriceInfo{Amount: 50.0, Currency: "USD"},
					Eligible: "QUALIFIED",
				},
			},
		}

		res, err := client.GetPriceForOfferingWrapper(session, "off-1")
		assert.NoError(t, err)
		
		resMap := res.(map[string]interface{})
		assert.Equal(t, 50.0, resMap["price"])
		assert.Equal(t, "USD", resMap["currency"])
	})

	t.Run("Should propagate error", func(t *testing.T) {
		_, err := client.GetPriceForOfferingWrapper("invalid", "off-1")
		assert.Error(t, err)
	})
}

func TestQualificationClient_NewQualificationClient(t *testing.T) {
	client := rpc.NewQualificationClient(nil)
	assert.NotNil(t, client)
}
