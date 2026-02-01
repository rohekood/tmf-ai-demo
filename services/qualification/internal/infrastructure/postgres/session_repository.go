package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/core/ports"

	"github.com/google/uuid"
)

type sessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new PostgreSQL-backed session repository
func NewSessionRepository(db *sql.DB) ports.SessionRepository {
	return &sessionRepository{db: db}
}

// Create stores a new qualification session and returns its ID
func (r *sessionRepository) Create(ctx context.Context, session *domain.QualificationSession) (string, error) {
	// Generate ID if not provided
	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	// Set timestamps
	session.CreatedAt = time.Now()
	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = time.Now().Add(24 * time.Hour)
	}

	// Marshal address and qualified offers to JSON
	addressJSON, err := json.Marshal(session.Address)
	if err != nil {
		return "", fmt.Errorf("failed to marshal address: %w", err)
	}

	offersJSON, err := json.Marshal(session.QualifiedOffers)
	if err != nil {
		return "", fmt.Errorf("failed to marshal qualified offers: %w", err)
	}

	query := `
		INSERT INTO qualification_sessions (
			id, customer_id, address, qualified_offers, status, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.ExecContext(ctx, query,
		session.ID,
		session.CustomerID,
		addressJSON,
		offersJSON,
		session.Status,
		session.CreatedAt,
		session.ExpiresAt,
	)

	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return session.ID, nil
}

// Get retrieves a qualification session by ID
func (r *sessionRepository) Get(ctx context.Context, sessionID string) (*domain.QualificationSession, error) {
	query := `
		SELECT id, customer_id, address, qualified_offers, status, created_at, expires_at
		FROM qualification_sessions
		WHERE id = $1
	`

	var session domain.QualificationSession
	var addressJSON, offersJSON []byte

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.CustomerID,
		&addressJSON,
		&offersJSON,
		&session.Status,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(addressJSON, &session.Address); err != nil {
		return nil, fmt.Errorf("failed to unmarshal address: %w", err)
	}

	if err := json.Unmarshal(offersJSON, &session.QualifiedOffers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal qualified offers: %w", err)
	}

	return &session, nil
}

// Update updates an existing qualification session
func (r *sessionRepository) Update(ctx context.Context, session *domain.QualificationSession) error {
	addressJSON, err := json.Marshal(session.Address)
	if err != nil {
		return fmt.Errorf("failed to marshal address: %w", err)
	}

	offersJSON, err := json.Marshal(session.QualifiedOffers)
	if err != nil {
		return fmt.Errorf("failed to marshal qualified offers: %w", err)
	}

	query := `
		UPDATE qualification_sessions
		SET customer_id = $2, address = $3, qualified_offers = $4, status = $5, expires_at = $6
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.CustomerID,
		addressJSON,
		offersJSON,
		session.Status,
		session.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	return nil
}

// Delete removes a qualification session
func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	query := `DELETE FROM qualification_sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

// FindExpired returns all sessions that have expired
func (r *sessionRepository) FindExpired(ctx context.Context) ([]*domain.QualificationSession, error) {
	query := `
		SELECT id, customer_id, address, qualified_offers, status, created_at, expires_at
		FROM qualification_sessions
		WHERE expires_at < NOW()
		ORDER BY expires_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.QualificationSession

	for rows.Next() {
		var session domain.QualificationSession
		var addressJSON, offersJSON []byte

		err := rows.Scan(
			&session.ID,
			&session.CustomerID,
			&addressJSON,
			&offersJSON,
			&session.Status,
			&session.CreatedAt,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		if err := json.Unmarshal(addressJSON, &session.Address); err != nil {
			return nil, fmt.Errorf("failed to unmarshal address: %w", err)
		}

		if err := json.Unmarshal(offersJSON, &session.QualifiedOffers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal qualified offers: %w", err)
		}

		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return sessions, nil
}
