package domain

import "time"

// QualificationSession represents a TMF679 qualification session
// with customer-specific pricing calculated at qualification time
type QualificationSession struct {
	ID              string           `json:"id"`
	CustomerID      string           `json:"customerId"`
	Address         Address          `json:"address"`
	QualifiedOffers []QualifiedOffer `json:"qualifiedOffers"`
	Status          string           `json:"status"` // "QUALIFIED", "UNQUALIFIED", "EXPIRED"
	CreatedAt       time.Time        `json:"createdAt"`
	ExpiresAt       time.Time        `json:"expiresAt"`
}

// QualifiedOffer represents an offering that passed qualification
// with customer-specific pricing
type QualifiedOffer struct {
	OfferingID   string   `json:"offeringId"`
	OfferingName string   `json:"offeringName"`
	Price        Price    `json:"price"`
	Eligibility  string   `json:"eligibility"` // "QUALIFIED", "NOT_AVAILABLE"
	Constraints  []string `json:"constraints,omitempty"`
}

// Price represents customer-specific pricing
type Price struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	TaxIncluded bool    `json:"taxIncluded"`
}
