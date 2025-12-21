package postgres

import (
	"time"
	"tmf/services/party-management/internal/domain"

	"gorm.io/gorm"
)

type PartyRepository struct {
	db *gorm.DB
}

func NewPartyRepository(db *gorm.DB) *PartyRepository {
	return &PartyRepository{db: db}
}

// Helper structs to map strictly to database tables if we want clean separation
type PartyTable struct {
	ID        string `gorm:"primaryKey"`
	Type      string
	Href      string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (PartyTable) TableName() string {
	return "parties"
}

type IndividualTable struct {
	ID         string `gorm:"primaryKey"`
	GivenName  string
	FamilyName string
}

func (IndividualTable) TableName() string {
	return "individuals"
}

type OrganizationTable struct {
	ID            string `gorm:"primaryKey"`
	TradingName   string
	IsLegalEntity bool
}

func (OrganizationTable) TableName() string {
	return "organizations"
}

func (r *PartyRepository) CreateIndividual(ind *domain.Individual) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create Party
		if err := tx.Table("parties").Create(&ind.Party).Error; err != nil {
			return err
		}
		// 2. Create Individual specific part
		// We can't pass *domain.Individual directly if it has embedded Party fields that don't belong to 'individuals' table
		// unless we use specific struct or map. using map for simplicity or struct with tags.
		// Since domain.Individual has embedded Party, GORM might try to insert Party fields into 'individuals' table if we just pass &ind.
		// Let's use a restricted map or struct.

		individualSpecifics := map[string]interface{}{
			"id":          ind.ID,
			"given_name":  ind.GivenName,
			"family_name": ind.FamilyName,
		}

		if err := tx.Table("individuals").Create(individualSpecifics).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *PartyRepository) GetIndividual(id string) (*domain.Individual, error) {
	var ind domain.Individual

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Fetch Party
		if err := tx.Table("parties").Where("id = ?", id).First(&ind.Party).Error; err != nil {
			return err
		}

		// Fetch Individual
		// We need to scan into a struct that matches columns
		type IndSpecifics struct {
			GivenName  string
			FamilyName string
		}
		var specifics IndSpecifics
		// Note: We already have ID in ind.Party.ID
		if err := tx.Table("individuals").Where("id = ?", id).First(&specifics).Error; err != nil {
			return err
		}

		ind.GivenName = specifics.GivenName
		ind.FamilyName = specifics.FamilyName
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &ind, nil
}

func (r *PartyRepository) CreateOrganization(org *domain.Organization) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("parties").Create(&org.Party).Error; err != nil {
			return err
		}

		orgSpecifics := map[string]interface{}{
			"id":              org.ID,
			"trading_name":    org.TradingName,
			"is_legal_entity": org.IsLegalEntity,
		}

		if err := tx.Table("organizations").Create(orgSpecifics).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *PartyRepository) GetOrganization(id string) (*domain.Organization, error) {
	var org domain.Organization

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("parties").Where("id = ?", id).First(&org.Party).Error; err != nil {
			return err
		}

		type OrgSpecifics struct {
			TradingName   string
			IsLegalEntity bool
		}
		var specifics OrgSpecifics
		if err := tx.Table("organizations").Where("id = ?", id).First(&specifics).Error; err != nil {
			return err
		}

		org.TradingName = specifics.TradingName
		org.IsLegalEntity = specifics.IsLegalEntity
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *PartyRepository) UpdateIndividual(ind *domain.Individual) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Update Party (base)
		if err := tx.Table("parties").Where("id = ?", ind.ID).Updates(&ind.Party).Error; err != nil {
			return err
		}

		// 2. Update Individual specifics
		individualSpecifics := map[string]interface{}{
			"given_name":  ind.GivenName,
			"family_name": ind.FamilyName,
		}

		if err := tx.Table("individuals").Where("id = ?", ind.ID).Updates(individualSpecifics).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *PartyRepository) UpdateOrganization(org *domain.Organization) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Update Party (base)
		if err := tx.Table("parties").Where("id = ?", org.ID).Updates(&org.Party).Error; err != nil {
			return err
		}

		// 2. Update Organization specifics
		orgSpecifics := map[string]interface{}{
			"trading_name":    org.TradingName,
			"is_legal_entity": org.IsLegalEntity,
		}

		if err := tx.Table("organizations").Where("id = ?", org.ID).Updates(orgSpecifics).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *PartyRepository) DeleteParty(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var p domain.Party
		if err := tx.Table("parties").Where("id = ?", id).First(&p).Error; err != nil {
			return err
		}

		switch p.Type {
		case domain.PartyTypeIndividual:
			if err := tx.Table("individuals").Where("id = ?", id).Delete(nil).Error; err != nil {
				return err
			}
		case domain.PartyTypeOrganization:
			if err := tx.Table("organizations").Where("id = ?", id).Delete(nil).Error; err != nil {
				return err
			}
		}

		if err := tx.Table("parties").Where("id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *PartyRepository) SearchParties(criteria map[string]interface{}) ([]domain.Party, error) {
	var parties []domain.Party
	query := r.db.Table("parties")

	// Apply filters
	// 1. Filter by ID or Type inside 'parties' table
	if val, ok := criteria["id"]; ok {
		query = query.Where("parties.id = ?", val)
	}
	if val, ok := criteria["type"]; ok {
		query = query.Where("parties.type = ?", val)
	}

	// 2. Filter by Name (requires Join)
	// Strategy: If 'name' is present, try to match against both Individual.Given/Family and Org.TradingName
	// OR if strict keys like 'givenName' are used, join specific table.
	// For simplicity, let's assuming explicit keys: 'given_name', 'family_name', 'trading_name'

	joinedIndividual := false
	joinedOrganization := false

	if val, ok := criteria["given_name"]; ok {
		query = query.Joins("JOIN individuals ON individuals.id = parties.id").Where("individuals.given_name = ?", val)
		joinedIndividual = true
	}
	if val, ok := criteria["family_name"]; ok {
		if !joinedIndividual {
			query = query.Joins("JOIN individuals ON individuals.id = parties.id")
			joinedIndividual = true
		}
		query = query.Where("individuals.family_name = ?", val)
	}

	if val, ok := criteria["trading_name"]; ok {
		query = query.Joins("JOIN organizations ON organizations.id = parties.id").Where("organizations.trading_name = ?", val)
		joinedOrganization = true
	}

	if val, ok := criteria["is_legal_entity"]; ok {
		if !joinedOrganization {
			query = query.Joins("JOIN organizations ON organizations.id = parties.id")
			joinedOrganization = true
		}
		query = query.Where("organizations.is_legal_entity = ?", val)
	}

	if err := query.Find(&parties).Error; err != nil {
		return nil, err
	}

	return parties, nil
}
