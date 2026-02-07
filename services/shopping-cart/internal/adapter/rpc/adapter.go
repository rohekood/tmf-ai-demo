package rpc

import (
	"context"

	"tmf/services/shopping-cart/internal/usecase"
)

// QualificationClientAdapter adapts the RPC client to the use case interface
type QualificationClientAdapter struct {
	client *QualificationClient
}

// NewQualificationClientAdapter creates an adapter
func NewQualificationClientAdapter(client *QualificationClient) *QualificationClientAdapter {
	return &QualificationClientAdapter{client: client}
}

// GetSession retrieves a session and returns it as the interface type
func (a *QualificationClientAdapter) GetSession(ctx context.Context, sessionID string) (usecase.QualificationSession, error) {
	return a.client.GetSession(ctx, sessionID)
}
