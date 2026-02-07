package repository

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending   = "PENDING"
	StatusPublished = "PUBLISHED"
	StatusFailed    = "FAILED"
)

type OutboxEventModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoutingKey  string    `gorm:"not null"`
	Payload     []byte    `gorm:"type:bytea;not null"`
	Headers     []byte    `gorm:"type:bytea"` // JSON map[string]string
	Status      string    `gorm:"index;default:'PENDING'"`
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

func (OutboxEventModel) TableName() string {
	return "outbox_events"
}

// Convert Map Headers to JSON Byte Array
func NewOutboxEvent(routingKey string, payload interface{}, headers map[string]string) (*OutboxEventModel, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	headerBytes, err := json.Marshal(headers)
	if err != nil {
		return nil, err
	}

	return &OutboxEventModel{
		ID:         uuid.New(),
		RoutingKey: routingKey,
		Payload:    data,
		Headers:    headerBytes,
		Status:     StatusPending,
		CreatedAt:  time.Now().UTC(),
	}, nil
}
