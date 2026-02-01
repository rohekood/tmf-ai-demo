package usecase

import (
	"context"
	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/core/ports"
)

// PricingCalculator calculates customer-specific prices
type PricingCalculator struct {
	catalogClient  ports.CatalogPricingClient
	customerClient ports.CustomerPricingClient
}

// NewPricingCalculator creates a new pricing calculator
func NewPricingCalculator(catalogClient ports.CatalogPricingClient, customerClient ports.CustomerPricingClient) *PricingCalculator {
	return &PricingCalculator{
		catalogClient:  catalogClient,
		customerClient: customerClient,
	}
}

// CalculatePrice calculates customer-specific price for an offering
func (p *PricingCalculator) CalculatePrice(
	ctx context.Context,
	offeringID string,
	customerID string,
) (*domain.Price, error) {
	// 1. Get offering base price from catalog
	offering, err := p.catalogClient.GetOffering(ctx, offeringID)
	if err != nil {
		return nil, err
	}

	// 2. Get customer tier
	customer, err := p.customerClient.GetCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// 3. Apply tier discount
	discount := getTierDiscount(customer.Tier)
	finalPrice := offering.BasePrice * (1 - discount)

	return &domain.Price{
		Amount:      finalPrice,
		Currency:    offering.Currency,
		TaxIncluded: false, // TODO: Add tax calculation
	}, nil
}

// getTierDiscount returns the discount percentage for a customer tier
func getTierDiscount(tier string) float64 {
	switch tier {
	case "VIP":
		return 0.20 // 20% discount
	case "Premium":
		return 0.10 // 10% discount
	case "Standard":
		return 0.00 // No discount
	default:
		return 0.00
	}
}
