package domain

import (
	"context"
	"time"
)

type contextKey string

const UserContextKey contextKey = "user_id"

// PartyType defines the type of party (Individual or Organization)
type PartyType string

const (
	PartyTypeIndividual   PartyType = "Individual"
	PartyTypeOrganization PartyType = "Organization"
)

// Party is the base entity mapping to 'parties' table
type Party struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Type      PartyType `gorm:"not null" json:"@type"`
	Href      string    `json:"href,omitempty"`
	Status    string    `json:"status,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	ContactMediums  []ContactMedium       `gorm:"foreignKey:PartyID" json:"contactMediums,omitempty"`
	Identifications []Identification      `gorm:"foreignKey:PartyID" json:"identifications,omitempty"`
	RelatedParties  []RelatedParty        `gorm:"foreignKey:PartyID" json:"relatedParties,omitempty"`
	Characteristics []PartyCharacteristic `gorm:"foreignKey:PartyID" json:"characteristics,omitempty"`
}

func (Party) TableName() string {
	return "parties"
}

type ContactMedium struct {
	ID              string     `gorm:"primaryKey" json:"id"`
	PartyID         string     `gorm:"not null;index" json:"partyId"`
	MediumType      string     `json:"mediumType"`
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
	return "party_contact_mediums"
}

type Identification struct {
	ID                 string     `gorm:"primaryKey" json:"id"`
	PartyID            string     `gorm:"not null;index" json:"partyId"`
	IdentificationType string     `json:"identificationType"`
	IdentificationID   string     `json:"identificationId"`
	IssuingAuthority   string     `json:"issuingAuthority,omitempty"`
	IssuingDate        time.Time  `json:"issuingDate,omitempty"`
	ValidForStart      time.Time  `json:"validForStart,omitempty"`
	ValidForEnd        *time.Time `json:"validForEnd,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

func (Identification) TableName() string {
	return "identifications"
}

type RelatedParty struct {
	ID               string     `gorm:"primaryKey" json:"id"`
	PartyID          string     `gorm:"not null;index" json:"partyId"`
	RelatedPartyID   string     `json:"relatedPartyId"`
	RelatedPartyName string     `json:"relatedPartyName"`
	Role             string     `json:"role"`
	ValidForStart    time.Time  `json:"validForStart,omitempty"`
	ValidForEnd      *time.Time `json:"validForEnd,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func (RelatedParty) TableName() string {
	return "related_parties"
}

type PartyCharacteristic struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	PartyID   string    `gorm:"not null;index" json:"partyId"`
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	ValueType string    `json:"valueType,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (PartyCharacteristic) TableName() string {
	return "party_characteristics"
}

// Individual represents a natural person.
// For GORM, we treats this as a separate structure that might be joined manually or via composition.
// To keep the domain model clean but usable with the repository pattern:
type Individual struct {
	Party      `gorm:"embedded"` // Embedding might flatten fields in GORM default, but we will handle split persistence in Repo.
	GivenName  string            `json:"givenName"`
	FamilyName string            `json:"familyName"`
	MiddleName string            `json:"middleName,omitempty"`
	BirthDate  string            `json:"birthDate,omitempty"`
	Gender     string            `json:"gender,omitempty"`
}

// Organization represents a company.
type Organization struct {
	Party            `gorm:"embedded"`
	TradingName      string `json:"tradingName"`
	IsLegalEntity    bool   `json:"isLegalEntity"`
	OrganizationType string `json:"organizationType,omitempty"`
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
