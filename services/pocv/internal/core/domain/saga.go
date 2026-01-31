package domain

import (
	"encoding/json"
	"time"
)

type SagaStatus string

const (
	SagaStatusPending      SagaStatus = "PENDING"
	SagaStatusInProgress   SagaStatus = "IN_PROGRESS"
	SagaStatusCompleted    SagaStatus = "COMPLETED"
	SagaStatusFailed       SagaStatus = "FAILED"
	SagaStatusCompensating SagaStatus = "COMPENSATING"
)

type SagaStep string

const (
	StepInit          SagaStep = "INIT"
	StepInventory     SagaStep = "INVENTORY"
	StepPayment       SagaStep = "PAYMENT"
	StepOrderCreation SagaStep = "ORDER_CREATION"
)

type SagaInstance struct {
	ID          string          `gorm:"primaryKey;type:uuid" json:"id"`
	CartID      string          `gorm:"not null;unique;type:uuid" json:"cartId"`
	CustomerID  *string         `gorm:"type:uuid" json:"customerId"` // Pointer for NULL support
	CurrentStep SagaStep        `gorm:"not null" json:"currentStep"`
	Status      SagaStatus      `gorm:"not null" json:"status"`
	Payload     json.RawMessage `gorm:"type:jsonb" json:"payload"`
	History     json.RawMessage `gorm:"type:jsonb" json:"history"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func (SagaInstance) TableName() string {
	return "saga_instances"
}

type OutboxEvent struct {
	ID          string     `gorm:"primaryKey;type:uuid" json:"id"`
	Topic       string     `gorm:"not null" json:"topic"`
	Payload     []byte     `gorm:"type:jsonb;not null" json:"payload"`
	Headers     []byte     `gorm:"type:jsonb" json:"headers"`
	Status      string     `gorm:"default:'PENDING'" json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}
