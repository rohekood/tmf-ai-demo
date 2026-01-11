package domain

import (
	"encoding/json"
	"time"
)

type OutboxEventStatus string

const (
	StatusPending   OutboxEventStatus = "PENDING"
	StatusPublished OutboxEventStatus = "PUBLISHED"
	StatusFailed    OutboxEventStatus = "FAILED"
)

type OutboxEvent struct {
	ID          string            `json:"id" gorm:"primaryKey;type:uuid"`
	RoutingKey  string            `json:"routing_key" gorm:"not null"`
	Payload     json.RawMessage   `json:"payload" gorm:"type:jsonb;not null"`
	Headers     json.RawMessage   `json:"headers" gorm:"type:jsonb"` // To store metadata like user context
	Status      OutboxEventStatus `json:"status" gorm:"not null;index"`
	CreatedAt   time.Time         `json:"created_at" gorm:"autoCreateTime"`
	ProcessedAt *time.Time        `json:"processed_at"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}
