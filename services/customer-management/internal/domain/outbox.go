package domain

import (
	"encoding/json"
	"time"
)

type OutboxEvent struct {
	ID          string          `gorm:"primaryKey" json:"id"`
	RoutingKey  string          `json:"routing_key"`
	Payload     json.RawMessage `gorm:"type:jsonb" json:"payload"`
	Headers     json.RawMessage `gorm:"type:jsonb" json:"headers"`
	Status      string          `json:"status"` // PENDING, PUBLISHED, FAILED
	CreatedAt   time.Time       `json:"createdAt"`
	ProcessedAt *time.Time      `json:"processedAt,omitempty"`
}
