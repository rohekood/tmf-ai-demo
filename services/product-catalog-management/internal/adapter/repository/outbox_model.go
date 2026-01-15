package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxEventStatus string

const (
	StatusPending   OutboxEventStatus = "PENDING"
	StatusPublished OutboxEventStatus = "PUBLISHED"
	StatusFailed    OutboxEventStatus = "FAILED"
)

type OutboxEventModel struct {
	ID          string            `gorm:"primaryKey;type:uuid"`
	RoutingKey  string            `gorm:"not null"`
	Payload     []byte            `gorm:"type:jsonb;not null"`
	Headers     []byte            `gorm:"type:jsonb"`
	Status      OutboxEventStatus `gorm:"not null;index"`
	CreatedAt   time.Time         `gorm:"autoCreateTime"`
	ProcessedAt *time.Time
}

func (m *OutboxEventModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.Status == "" {
		m.Status = StatusPending
	}
	return
}

func (m *OutboxEventModel) TableName() string {
	return "outbox_events"
}
