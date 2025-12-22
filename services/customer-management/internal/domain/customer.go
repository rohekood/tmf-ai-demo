package domain

import (
	"context"
	"time"
)

// CustomerStatus represents the lifecycle state of a customer.
type CustomerStatus string

const (
	CustomerStatusActive    CustomerStatus = "Active"
	CustomerStatusSuspended CustomerStatus = "Suspended"
	CustomerStatusClosed    CustomerStatus = "Closed"
)

// Customer represents a party playing a customer role.
type Customer struct {
	ID             string         `gorm:"primaryKey"`
	Name           string         // Display name
	Status         CustomerStatus `gorm:"not null"`
	StatusReason   string
	ValidForStart  time.Time
	ValidForEnd    *time.Time
	PartyID        string `gorm:"not null;index"` // Reference to TMF632 Party
	PartyType      string // Individual or Organization
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Accounts       []CustomerAccount `gorm:"foreignKey:CustomerID"`
	CreditProfiles []CreditProfile   `gorm:"foreignKey:CustomerID"`
	ContactMediums []ContactMedium   `gorm:"foreignKey:CustomerID"`
}

// IsActive checks if the customer is active at the given time.
func (c *Customer) IsActive(t time.Time) bool {
	if c.Status != CustomerStatusActive {
		return false
	}
	if !c.ValidForStart.IsZero() && t.Before(c.ValidForStart) {
		return false
	}
	if c.ValidForEnd != nil && t.After(*c.ValidForEnd) {
		return false
	}
	return true
}

// CustomerAccount represents a billing or financial relationship.
type CustomerAccount struct {
	ID            string `gorm:"primaryKey"`
	CustomerID    string `gorm:"not null;index"`
	Name          string
	AccountStatus string
	AccountType   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CreditProfile provides information regarding the customer's creditworthiness.
type CreditProfile struct {
	ID                string `gorm:"primaryKey"`
	CustomerID        string `gorm:"not null;index"`
	CreditProfileDate time.Time
	CreditRiskScore   int
	CreditScore       int
	ValidForStart     time.Time
	ValidForEnd       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ContactMedium represents customer-specific contact information.
type ContactMedium struct {
	ID            string `gorm:"primaryKey"`
	CustomerID    string `gorm:"not null;index"`
	MediumType    string // e.g., Email, Phone
	Preferred     bool
	Value         string
	ValidForStart time.Time
	ValidForEnd   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Repository defines the interface for Customer storage.
type Repository interface {
	CreateCustomer(ctx context.Context, c *Customer) error
	GetCustomer(ctx context.Context, id string) (*Customer, error)
	UpdateCustomer(ctx context.Context, c *Customer) error
	PatchCustomer(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteCustomer(ctx context.Context, id string) error
	SearchCustomers(ctx context.Context, criteria map[string]interface{}) ([]Customer, error)
}
