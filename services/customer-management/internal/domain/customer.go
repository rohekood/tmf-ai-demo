package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type contextKey string

const UserContextKey contextKey = "user_id"

// CustomerStatus represents the lifecycle state of a customer.
type CustomerStatus string

const (
	CustomerStatusActive    CustomerStatus = "Active"
	CustomerStatusSuspended CustomerStatus = "Suspended"
	CustomerStatusClosed    CustomerStatus = "Closed"
)

// Customer represents a party playing a customer role.
type Customer struct {
	ID                   string                   `gorm:"primaryKey" json:"id"`
	Name                 string                   `json:"name"` // Display name
	Status               CustomerStatus           `gorm:"not null" json:"status"`
	StatusReason         string                   `json:"statusReason,omitempty"`
	ValidForStart        time.Time                `json:"validForStart,omitempty"`
	ValidForEnd          *time.Time               `json:"validForEnd,omitempty"`
	PartyID              string                   `gorm:"not null;index" json:"partyId"`
	PartyType            string                   `json:"partyType"` // Individual or Organization
	CreatedAt            time.Time                `json:"createdAt"`
	UpdatedAt            time.Time                `json:"updatedAt"`
	Accounts             []CustomerAccount        `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE" json:"accounts,omitempty"`
	CreditProfiles       []CreditProfile          `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE" json:"creditProfiles,omitempty"`
	ContactMediums       []ContactMedium          `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE" json:"contactMediums,omitempty"`
	Characteristics      []CustomerCharacteristic `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE" json:"characteristics,omitempty"`
	TaxExemptions        []TaxExemption           `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE" json:"taxExemptions,omitempty"`
	PrivacyConsents      []PrivacyConsent         `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE" json:"privacyConsents,omitempty"`
	RelatedParties       []RelatedParty           `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE" json:"relatedParties,omitempty"`
	PaymentMethods       []PaymentMethod          `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE" json:"paymentMethods,omitempty"`
	MarketSegments       []MarketSegment          `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE" json:"marketSegments,omitempty"`
	CustomerInteractions []CustomerInteraction    `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE" json:"customerInteractions,omitempty"`
}

func (Customer) TableName() string {
	return "customers"
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
	ID            string    `gorm:"primaryKey" json:"id"`
	CustomerID    string    `gorm:"not null;index" json:"customerId"`
	Name          string    `json:"name"`
	AccountStatus string    `json:"accountStatus"`
	AccountType   string    `json:"accountType"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	BillFormat    string    `json:"billFormat"`
	BillingCycle  string    `json:"billingCycle"`
}

func (CustomerAccount) TableName() string {
	return "customer_accounts"
}

func (c *CustomerAccount) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// CreditProfile provides information regarding the customer's creditworthiness.
type CreditProfile struct {
	ID                string     `gorm:"primaryKey" json:"id"`
	CustomerID        string     `gorm:"not null;index" json:"customerId"`
	CreditProfileDate time.Time  `json:"creditProfileDate"`
	CreditRiskScore   int        `json:"creditRiskScore"`
	CreditScore       int        `json:"creditScore"`
	ValidForStart     time.Time  `json:"validForStart,omitempty"`
	ValidForEnd       *time.Time `json:"validForEnd,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

func (CreditProfile) TableName() string {
	return "credit_profiles"
}

func (c *CreditProfile) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// ContactMedium represents customer-specific contact information.
type ContactMedium struct {
	ID              string     `gorm:"primaryKey" json:"id"`
	CustomerID      string     `gorm:"not null;index" json:"customerId"`
	MediumType      string     `json:"mediumType"` // e.g., Email, Phone
	Preferred       bool       `json:"preferred"`
	Value           string     `json:"value"`
	Street1         string     `json:"street1,omitempty"`
	Street2         string     `json:"street2,omitempty"`
	City            string     `json:"city,omitempty"`
	StateOrProvince string     `json:"stateOrProvince,omitempty"`
	Postcode        string     `json:"postcode,omitempty"`
	Country         string     `json:"country,omitempty"`
	ValidForStart   time.Time  `json:"validForStart,omitempty"`
	ValidForEnd     *time.Time `json:"validForEnd,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (ContactMedium) TableName() string {
	return "contact_mediums"
}

func (c *ContactMedium) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// CustomerCharacteristic represents a dynamic attribute of a customer.
type CustomerCharacteristic struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	CustomerID string    `gorm:"not null;index" json:"customerId"`
	Name       string    `json:"name"`
	Value      string    `json:"value"`
	ValueType  string    `json:"valueType,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (CustomerCharacteristic) TableName() string {
	return "customer_characteristics"
}

func (c *CustomerCharacteristic) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

type TaxExemption struct {
	ID                  string     `gorm:"primaryKey" json:"id"`
	CustomerID          string     `gorm:"not null;index" json:"customerId"`
	CertificateNumber   string     `json:"certificateNumber"`
	IssuingJurisdiction string     `json:"issuingJurisdiction"`
	ValidForStart       time.Time  `json:"validForStart,omitempty"`
	ValidForEnd         *time.Time `json:"validForEnd,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

func (TaxExemption) TableName() string {
	return "tax_exemptions"
}

func (t *TaxExemption) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

type PrivacyConsent struct {
	ID            string     `gorm:"primaryKey" json:"id"`
	CustomerID    string     `gorm:"not null;index" json:"customerId"`
	ConsentType   string     `json:"consentType"`
	Status        string     `json:"status"`
	ValidForStart time.Time  `json:"validForStart,omitempty"`
	ValidForEnd   *time.Time `json:"validForEnd,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func (PrivacyConsent) TableName() string {
	return "privacy_consents"
}

func (p *PrivacyConsent) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

type RelatedParty struct {
	ID             string     `gorm:"primaryKey" json:"id"`
	CustomerID     string     `gorm:"not null;index" json:"customerId"`
	RelatedPartyID string     `json:"relatedPartyId"`
	Role           string     `json:"role"`
	Name           string     `json:"name"`
	ValidForStart  time.Time  `json:"validForStart"`
	ValidForEnd    *time.Time `json:"validForEnd,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (RelatedParty) TableName() string {
	return "related_parties"
}

func (r *RelatedParty) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

type PaymentMethod struct {
	ID            string          `gorm:"primaryKey" json:"id"`
	CustomerID    string          `gorm:"not null;index" json:"customerId"`
	Type          string          `json:"type"`
	Token         string          `json:"-"` // Never expose via default JSON
	Details       json.RawMessage `gorm:"type:jsonb" json:"details"`
	IsDefault     bool            `json:"isDefault"`
	ValidForStart time.Time       `json:"validForStart"`
	ValidForEnd   *time.Time      `json:"validForEnd,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (PaymentMethod) TableName() string {
	return "payment_methods"
}

func (p *PaymentMethod) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

type MarketSegment struct {
	ID         string `gorm:"primaryKey" json:"id"`
	CustomerID string `gorm:"not null;index" json:"customerId"`
	Name       string `json:"name"`
	Category   string `json:"category"`
}

func (MarketSegment) TableName() string {
	return "market_segments"
}

func (m *MarketSegment) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

type CustomerInteraction struct {
	ID              string    `gorm:"primaryKey" json:"id"`
	CustomerID      string    `gorm:"not null;index" json:"customerId"`
	InteractionDate time.Time `json:"interactionDate"`
	Channel         string    `json:"channel"`
	Type            string    `json:"type"`
	Description     string    `json:"description"`
	AgentID         string    `json:"agentId"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (CustomerInteraction) TableName() string {
	return "customer_interactions"
}

func (c *CustomerInteraction) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// Repository defines the interface for Customer storage.
type Repository interface {
	CreateCustomer(ctx context.Context, c *Customer) error
	GetCustomer(ctx context.Context, id string) (*Customer, error)
	UpdateCustomer(ctx context.Context, c *Customer) error
	PatchCustomer(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteCustomer(ctx context.Context, id string) error
	SearchCustomers(ctx context.Context, criteria map[string]interface{}) ([]Customer, error)
	AddInteraction(ctx context.Context, interaction *CustomerInteraction) error
}
