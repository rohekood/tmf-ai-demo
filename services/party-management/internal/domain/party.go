package domain

import (
	"context"
	"time"
)

// PartyType defines the type of party (Individual or Organization)
type PartyType string

const (
	PartyTypeIndividual   PartyType = "Individual"
	PartyTypeOrganization PartyType = "Organization"
)

// Party is the base entity mapping to 'parties' table
type Party struct {
	ID        string    `gorm:"primaryKey"`
	Type      PartyType `gorm:"not null"`
	Href      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Individual represents a natural person.
// For GORM, we treats this as a separate structure that might be joined manually or via composition.
// To keep the domain model clean but usable with the repository pattern:
type Individual struct {
	Party      `gorm:"embedded"` // Embedding might flatten fields in GORM default, but we will handle split persistence in Repo.
	GivenName  string
	FamilyName string
}

// Organization represents a company.
type Organization struct {
	Party         `gorm:"embedded"`
	TradingName   string
	IsLegalEntity bool
}

// Repository defines the interface for Party storage.
// All methods require context.Context for timeout control, cancellation, and tracing.
type Repository interface {
	// GetParty retrieves any party by ID, regardless of type
	GetParty(ctx context.Context, id string) (*Party, error)

	CreateIndividual(ctx context.Context, ind *Individual) error
	GetIndividual(ctx context.Context, id string) (*Individual, error)
	UpdateIndividual(ctx context.Context, ind *Individual) error

	CreateOrganization(ctx context.Context, org *Organization) error
	GetOrganization(ctx context.Context, id string) (*Organization, error)
	UpdateOrganization(ctx context.Context, org *Organization) error

	DeleteParty(ctx context.Context, id string) error
	SearchParties(ctx context.Context, criteria map[string]interface{}) ([]Party, error)
}
