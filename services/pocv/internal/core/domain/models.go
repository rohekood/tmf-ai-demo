package domain

import (
	"time"

	"github.com/google/uuid"
)

// Order Statuses
const (
	OrderStatusNew       = "NEW"
	OrderStatusPending   = "PENDING"
	OrderStatusReserved  = "INVENTORY_RESERVED"
	OrderStatusPaid      = "PAID"
	OrderStatusCompleted = "COMPLETED"
	OrderStatusFailed    = "FAILED"
	OrderStatusCancelled = "CANCELLED"
)

type Order struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	CustomerID string    `gorm:"not null"`
	Status     string    `gorm:"not null"`
	TotalPrice float64
	Items      []OrderItem `gorm:"foreignKey:OrderID"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderItem struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	OrderID           uuid.UUID `gorm:"type:uuid;not null"`
	ProductOfferingID string
	Quantity          int
	Price             float64
}

// Commands
type SubmitOrderCommand struct {
	CartID     string         `json:"cartId"`
	CustomerID string         `json:"customerId"`
	Items      []OrderItemDTO `json:"items"`
}

type OrderItemDTO struct {
	ProductOfferingID string  `json:"productOfferingId"`
	Quantity          int     `json:"quantity"`
	Price             float64 `json:"price"`
}

// Integration Commands (emitted by Saga)
type ReserveInventoryCommand struct {
	OrderID string         `json:"orderId"`
	Items   []OrderItemDTO `json:"items"`
}

// Events
type OrderCreatedEvent struct {
	OrderID    string    `json:"orderId"`
	CustomerID string    `json:"customerId"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
}

type OrderCompletedEvent struct {
	OrderID string `json:"orderId"`
}

type OrderFailedEvent struct {
	OrderID string `json:"orderId"`
	Reason  string `json:"reason"`
}
