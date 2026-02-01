package ports

import (
	"context"
)

// CatalogPricingClient defines the interface for querying product catalog for pricing
type CatalogPricingClient interface {
	// GetOffering retrieves a product offering with base price
	GetOffering(ctx context.Context, offeringID string) (*Offering, error)
}

// Offering represents a product offering from the catalog
type Offering struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	BasePrice float64 `json:"basePrice"`
	Currency  string  `json:"currency"`
}

// CustomerPricingClient defines the interface for querying customer information
type CustomerPricingClient interface {
	// GetCustomer retrieves customer information including tier
	GetCustomer(ctx context.Context, customerID string) (*Customer, error)
}

// Customer represents customer information
type Customer struct {
	ID      string `json:"id"`
	Tier    string `json:"tier"`    // "Standard", "Premium", "VIP"
	Segment string `json:"segment"` // "Residential", "Business"
	Status  string `json:"status"`  // "Active", "Suspended"
}
