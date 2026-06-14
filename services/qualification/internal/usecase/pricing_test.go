package usecase

import (
	"context"
	"errors"
	"testing"

	"tmf/services/qualification/internal/core/ports"

	"github.com/stretchr/testify/assert"
)

// Mock implementations for pricing tests
type mockCatalogPricingClient struct {
	offering *ports.Offering
	err      error
}

func (m *mockCatalogPricingClient) GetOffering(ctx context.Context, offeringID string) (*ports.Offering, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.offering, nil
}

type mockCustomerPricingClient struct {
	customer *ports.Customer
	err      error
}

func (m *mockCustomerPricingClient) GetCustomer(ctx context.Context, customerID string) (*ports.Customer, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.customer, nil
}

func TestPricingCalculator_CalculatePrice(t *testing.T) {
	ctx := context.Background()

	t.Run("VIP customer gets 20% discount", func(t *testing.T) {
		catalogClient := &mockCatalogPricingClient{
			offering: &ports.Offering{
				ID:        "offer-1",
				Name:      "Fiber 100",
				BasePrice: 100.0,
				Currency:  "EUR",
			},
		}
		customerClient := &mockCustomerPricingClient{
			customer: &ports.Customer{
				ID:      "customer-1",
				Tier:    "VIP",
				Segment: "Residential",
				Status:  "Active",
			},
		}

		calc := NewPricingCalculator(catalogClient, customerClient)
		price, err := calc.CalculatePrice(ctx, "offer-1", "customer-1")

		assert.NoError(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, 80.0, price.Amount) // 100 - 20% = 80
		assert.Equal(t, "EUR", price.Currency)
		assert.False(t, price.TaxIncluded)
	})

	t.Run("Premium customer gets 10% discount", func(t *testing.T) {
		catalogClient := &mockCatalogPricingClient{
			offering: &ports.Offering{
				ID:        "offer-1",
				Name:      "Fiber 100",
				BasePrice: 100.0,
				Currency:  "EUR",
			},
		}
		customerClient := &mockCustomerPricingClient{
			customer: &ports.Customer{
				ID:      "customer-2",
				Tier:    "Premium",
				Segment: "Residential",
				Status:  "Active",
			},
		}

		calc := NewPricingCalculator(catalogClient, customerClient)
		price, err := calc.CalculatePrice(ctx, "offer-1", "customer-2")

		assert.NoError(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, 90.0, price.Amount) // 100 - 10% = 90
		assert.Equal(t, "EUR", price.Currency)
	})

	t.Run("Standard customer gets no discount", func(t *testing.T) {
		catalogClient := &mockCatalogPricingClient{
			offering: &ports.Offering{
				ID:        "offer-1",
				Name:      "Fiber 100",
				BasePrice: 100.0,
				Currency:  "EUR",
			},
		}
		customerClient := &mockCustomerPricingClient{
			customer: &ports.Customer{
				ID:      "customer-3",
				Tier:    "Standard",
				Segment: "Residential",
				Status:  "Active",
			},
		}

		calc := NewPricingCalculator(catalogClient, customerClient)
		price, err := calc.CalculatePrice(ctx, "offer-1", "customer-3")

		assert.NoError(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, 100.0, price.Amount) // No discount
		assert.Equal(t, "EUR", price.Currency)
	})

	t.Run("Unknown tier defaults to no discount", func(t *testing.T) {
		catalogClient := &mockCatalogPricingClient{
			offering: &ports.Offering{
				ID:        "offer-1",
				Name:      "Fiber 100",
				BasePrice: 100.0,
				Currency:  "EUR",
			},
		}
		customerClient := &mockCustomerPricingClient{
			customer: &ports.Customer{
				ID:      "customer-4",
				Tier:    "Unknown",
				Segment: "Residential",
				Status:  "Active",
			},
		}

		calc := NewPricingCalculator(catalogClient, customerClient)
		price, err := calc.CalculatePrice(ctx, "offer-1", "customer-4")

		assert.NoError(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, 100.0, price.Amount) // Default to no discount
	})

	t.Run("Catalog client error is propagated", func(t *testing.T) {
		catalogClient := &mockCatalogPricingClient{
			err: errors.New("catalog service unavailable"),
		}
		customerClient := &mockCustomerPricingClient{
			customer: &ports.Customer{
				ID:   "customer-1",
				Tier: "VIP",
			},
		}

		calc := NewPricingCalculator(catalogClient, customerClient)
		price, err := calc.CalculatePrice(ctx, "offer-1", "customer-1")

		assert.Error(t, err)
		assert.Nil(t, price)
		assert.Contains(t, err.Error(), "catalog service unavailable")
	})

	t.Run("Customer client error is propagated", func(t *testing.T) {
		catalogClient := &mockCatalogPricingClient{
			offering: &ports.Offering{
				ID:        "offer-1",
				BasePrice: 100.0,
				Currency:  "EUR",
			},
		}
		customerClient := &mockCustomerPricingClient{
			err: errors.New("customer not found"),
		}

		calc := NewPricingCalculator(catalogClient, customerClient)
		price, err := calc.CalculatePrice(ctx, "offer-1", "customer-1")

		assert.Error(t, err)
		assert.Nil(t, price)
		assert.Contains(t, err.Error(), "customer not found")
	})
}

func TestPricingCalculator_CalculateGenericPrice(t *testing.T) {
	ctx := context.Background()

	t.Run("returns base price without customer lookup", func(t *testing.T) {
		catalogClient := &mockCatalogPricingClient{
			offering: &ports.Offering{
				ID:        "offer-1",
				Name:      "Fiber 100",
				BasePrice: 100.0,
				Currency:  "EUR",
			},
		}
		// Nil customer client proves no customer lookup happens on this path.
		calc := NewPricingCalculator(catalogClient, nil)
		price, err := calc.CalculateGenericPrice(ctx, "offer-1")

		assert.NoError(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, 100.0, price.Amount) // base price, no discount
		assert.Equal(t, "EUR", price.Currency)
		assert.False(t, price.TaxIncluded)
	})

	t.Run("catalog client error is propagated", func(t *testing.T) {
		catalogClient := &mockCatalogPricingClient{
			err: errors.New("catalog service unavailable"),
		}
		calc := NewPricingCalculator(catalogClient, nil)
		price, err := calc.CalculateGenericPrice(ctx, "offer-1")

		assert.Error(t, err)
		assert.Nil(t, price)
		assert.Contains(t, err.Error(), "catalog service unavailable")
	})
}

func TestGetTierDiscount(t *testing.T) {
	tests := []struct {
		name     string
		tier     string
		expected float64
	}{
		{"VIP tier", "VIP", 0.20},
		{"Premium tier", "Premium", 0.10},
		{"Standard tier", "Standard", 0.00},
		{"Unknown tier", "Unknown", 0.00},
		{"Empty tier", "", 0.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discount := getTierDiscount(tt.tier)
			assert.Equal(t, tt.expected, discount)
		})
	}
}
