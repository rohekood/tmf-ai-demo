package ports

import (
	"context"
	"tmf/services/qualification/internal/core/domain"
)

// SessionRepository defines the interface for persisting qualification sessions
type SessionRepository interface {
	// Create stores a new qualification session and returns its ID
	Create(ctx context.Context, session *domain.QualificationSession) (string, error)

	// Get retrieves a qualification session by ID
	Get(ctx context.Context, sessionID string) (*domain.QualificationSession, error)

	// Update updates an existing qualification session
	Update(ctx context.Context, session *domain.QualificationSession) error

	// Delete removes a qualification session
	Delete(ctx context.Context, sessionID string) error

	// FindExpired returns all sessions that have expired
	FindExpired(ctx context.Context) ([]*domain.QualificationSession, error)
}
