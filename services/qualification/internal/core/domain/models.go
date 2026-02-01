package domain

import "errors"

var (
	ErrInvalidAddress     = errors.New("invalid address")
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
	ErrBackendFailure     = errors.New("backend dependency failed")
)

type Address struct {
	Street string `json:"street"`
	Number string `json:"number"`
	City   string `json:"city"`
	Zip    string `json:"zip"`
}

type CheckEligibilityCommand struct {
	Address        Address  `json:"address"`
	CustomerID     string   `json:"customerId"` // NEW: For pricing calculation
	CategoryFilter []string `json:"categoryFilter"`
	CorrelationID  string   `json:"correlationId"`
	ReplyTo        string   `json:"replyTo"`
}

type EligibilityResult struct {
	QualificationID      string              `json:"qualificationId"`
	SessionID            string              `json:"sessionId,omitempty"` // NEW: Session ID for reuse
	CorrelationID        string              `json:"correlationId"`
	Status               QualificationStatus `json:"status"`
	EligibleCategories   []EligibleCategory  `json:"eligibleCategories"`
	QualifiedOffers      []QualifiedOffer    `json:"qualifiedOffers,omitempty"` // NEW: Offers with prices
	UnavailabilityReason string              `json:"unavailabilityReason,omitempty"`
}

type QualificationStatus string

const (
	StatusQualified   QualificationStatus = "Qualified"
	StatusUnqualified QualificationStatus = "Unqualified"
	StatusError       QualificationStatus = "Error"
)

type EligibleCategory struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Characteristics map[string]string `json:"characteristics"` // Dynamic constraints
}
