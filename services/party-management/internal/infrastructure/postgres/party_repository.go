package postgres

import (
	"context"
	"errors"
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

// Helper structs to map strictly to database tables
type PartyTable struct {
	ID        string `gorm:"primaryKey"`
	Type      string
	Href      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (PartyTable) TableName() string {
	return "parties"
}

func (r *PartyRepository) GetParty(ctx context.Context, id string) (*domain.Party, error) {
	var p domain.Party
	if err := r.db.WithContext(ctx).
		Preload("ContactMediums").
		Preload("Identifications").
		Preload("RelatedParties").
		First(&p, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &p, nil
}

func (r *PartyRepository) CreateIndividual(ctx context.Context, ind *domain.Individual) error {
	return r.withUser(ctx, func(tx *gorm.DB) error {
		// 1. Create Party (with sub-resources)
		if err := tx.Create(&ind.Party).Error; err != nil {
			return err
		}

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

func (r *PartyRepository) GetIndividual(ctx context.Context, id string) (*domain.Individual, error) {
	var ind domain.Individual

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Fetch Party with sub-resources
		if err := tx.Table("parties").
			Preload("ContactMediums").
			Preload("Identifications").
			Preload("RelatedParties").
			Where("id = ?", id).First(&ind.Party).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		// Fetch Individual specifics
		type IndSpecifics struct {
			GivenName  string
			FamilyName string
		}
		var specifics IndSpecifics
		if err := tx.Table("individuals").Where("id = ?", id).First(&specifics).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound // Should not happen if party exists, but for safety
			}
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

func (r *PartyRepository) CreateOrganization(ctx context.Context, org *domain.Organization) error {
	return r.withUser(ctx, func(tx *gorm.DB) error {
		// 1. Create Party (with sub-resources)
		if err := tx.Create(&org.Party).Error; err != nil {
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

func (r *PartyRepository) GetOrganization(ctx context.Context, id string) (*domain.Organization, error) {
	var org domain.Organization

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("parties").
			Preload("ContactMediums").
			Preload("Identifications").
			Preload("RelatedParties").
			Where("id = ?", id).First(&org.Party).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		type OrgSpecifics struct {
			TradingName   string
			IsLegalEntity bool
		}
		var specifics OrgSpecifics
		if err := tx.Table("organizations").Where("id = ?", id).First(&specifics).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
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

func (r *PartyRepository) UpdateIndividual(ctx context.Context, ind *domain.Individual) error {
	return r.withUser(ctx, func(tx *gorm.DB) error {
		// 1. Update Party (base)
		result := tx.Table("parties").Where("id = ?", ind.ID).Updates(&ind.Party)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
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

func (r *PartyRepository) UpdateOrganization(ctx context.Context, org *domain.Organization) error {
	return r.withUser(ctx, func(tx *gorm.DB) error {
		// 1. Update Party (base)
		result := tx.Table("parties").Where("id = ?", org.ID).Updates(&org.Party)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
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

func (r *PartyRepository) DeleteParty(ctx context.Context, id string) error {
	return r.withUser(ctx, func(tx *gorm.DB) error {
		var p domain.Party
		if err := tx.Table("parties").Where("id = ?", id).First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
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

func (r *PartyRepository) SearchParties(ctx context.Context, criteria map[string]interface{}) ([]domain.Party, error) {
	var parties []domain.Party
	query := r.db.WithContext(ctx).Table("parties")

	// Apply filters
	if val, ok := criteria["id"]; ok {
		query = query.Where("parties.id = ?", val)
	}
	if val, ok := criteria["type"]; ok {
		query = query.Where("parties.type = ?", val)
	}

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

// Helpers

func (r *PartyRepository) withUser(ctx context.Context, fn func(tx *gorm.DB) error) error {
	userID, _ := ctx.Value(domain.UserContextKey).(string)
	if userID == "" {
		userID = "unknown"
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Set local session variable for audit trigger
		if err := tx.Exec("SELECT set_config('app.current_user', ?, true)", userID).Error; err != nil {
			return err
		}
		return fn(tx)
	})
}
