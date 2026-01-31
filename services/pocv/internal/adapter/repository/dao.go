package repository

import (
	"time"
	"tmf/services/pocv/internal/core/domain"
)

// DAO Definitions
type SagaTable struct {
	ID          string  `gorm:"primaryKey;type:uuid"`
	CartID      string  `gorm:"not null;unique;type:uuid"`
	CustomerID  *string `gorm:"type:uuid"` // Pointer for NULL support
	CurrentStep string  `gorm:"not null"`
	Status      string  `gorm:"not null"`
	Payload     []byte  `gorm:"type:jsonb"`
	History     []byte  `gorm:"type:jsonb"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (SagaTable) TableName() string {
	return "saga_instances"
}

type OutboxTable struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	Topic       string `gorm:"not null"`
	Payload     []byte `gorm:"type:jsonb;not null"`
	Headers     []byte `gorm:"type:jsonb"`
	Status      string `gorm:"default:'PENDING'"`
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

func (OutboxTable) TableName() string {
	return "outbox_events"
}

// Mappers
func toDomainSaga(t *SagaTable) *domain.SagaInstance {
	return &domain.SagaInstance{
		ID:          t.ID,
		CartID:      t.CartID,
		CustomerID:  t.CustomerID,
		CurrentStep: domain.SagaStep(t.CurrentStep),
		Status:      domain.SagaStatus(t.Status),
		Payload:     t.Payload,
		History:     t.History,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toDAOSaga(d *domain.SagaInstance) *SagaTable {
	return &SagaTable{
		ID:          d.ID,
		CartID:      d.CartID,
		CustomerID:  d.CustomerID,
		CurrentStep: string(d.CurrentStep),
		Status:      string(d.Status),
		Payload:     d.Payload,
		History:     d.History,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func toDAOOutbox(events []domain.OutboxEvent) []OutboxTable {
	out := make([]OutboxTable, len(events))
	for i, e := range events {
		out[i] = OutboxTable{
			ID:        e.ID,
			Topic:     e.Topic,
			Payload:   e.Payload,
			Status:    e.Status,
			CreatedAt: e.CreatedAt,
		}
	}
	return out
}
