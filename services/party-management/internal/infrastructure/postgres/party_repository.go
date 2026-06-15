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
	if err := GetTx(ctx, r.db).
		Preload("ContactMediums").
		Preload("Identifications").
		Preload("RelatedParties").
		Preload("Characteristics").
		Preload("ExternalReferences").
		Preload("TaxExemptions").
		Preload("Attachments").
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
		// 1. Create Party (base) - Use a copy to avoid GORM clearing associations
		partyBase := ind.Party
		// Ensure associations are nil in the copy so GORM doesn't try to save them
		partyBase.ContactMediums = nil
		partyBase.Identifications = nil
		partyBase.RelatedParties = nil
		partyBase.Characteristics = nil
		partyBase.ExternalReferences = nil
		partyBase.TaxExemptions = nil
		partyBase.Attachments = nil

		if err := tx.Create(&partyBase).Error; err != nil {
			return err
		}

		// Sync back generated fields (CreatedAt, UpdatedAt)
		ind.CreatedAt = partyBase.CreatedAt
		ind.UpdatedAt = partyBase.UpdatedAt

		individualSpecifics := map[string]any{
			"id":          ind.ID,
			"given_name":  ind.GivenName,
			"family_name": ind.FamilyName,
			"middle_name": ind.MiddleName,
			"birth_date":  ind.BirthDate,
			"gender":      ind.Gender,
		}

		if err := tx.Table("individuals").Create(individualSpecifics).Error; err != nil {
			return err
		}

		// 3. Create Sub-resources (using original ind with data)
		if err := r.updateSubResources(tx, ind.ID, &ind.Party); err != nil {
			return err
		}

		return nil
	})
}

func (r *PartyRepository) GetIndividual(ctx context.Context, id string) (*domain.Individual, error) {
	var ind domain.Individual

	err := GetTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// Fetch Party with sub-resources
		if err := tx.Table("parties").
			Preload("ContactMediums").
			Preload("Identifications").
			Preload("RelatedParties").
			Preload("Characteristics").
			Preload("ExternalReferences").
			Preload("TaxExemptions").
			Preload("Attachments").
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
			MiddleName string
			BirthDate  string
			Gender     string
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
		ind.MiddleName = specifics.MiddleName
		ind.BirthDate = specifics.BirthDate
		ind.Gender = specifics.Gender
		return nil
	})

	if err != nil {
		return nil, err
	}

	_ = r.loadAttachmentContents(ctx, &ind.Party)

	return &ind, nil
}

func (r *PartyRepository) CreateOrganization(ctx context.Context, org *domain.Organization) error {
	return r.withUser(ctx, func(tx *gorm.DB) error {
		// 1. Create Party (base) - Use a copy
		partyBase := org.Party
		partyBase.ContactMediums = nil
		partyBase.Identifications = nil
		partyBase.RelatedParties = nil
		partyBase.Characteristics = nil
		partyBase.ExternalReferences = nil
		partyBase.TaxExemptions = nil
		partyBase.Attachments = nil

		if err := tx.Create(&partyBase).Error; err != nil {
			return err
		}

		// Sync back
		org.CreatedAt = partyBase.CreatedAt
		org.UpdatedAt = partyBase.UpdatedAt

		orgSpecifics := map[string]any{
			"id":                org.ID,
			"trading_name":      org.TradingName,
			"is_legal_entity":   org.IsLegalEntity,
			"organization_type": org.OrganizationType,
		}

		if err := tx.Table("organizations").Create(orgSpecifics).Error; err != nil {
			return err
		}

		// 3. Create Sub-resources
		if err := r.updateSubResources(tx, org.ID, &org.Party); err != nil {
			return err
		}

		return nil
	})
}

func (r *PartyRepository) GetOrganization(ctx context.Context, id string) (*domain.Organization, error) {
	var org domain.Organization

	err := GetTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("parties").
			Preload("ContactMediums").
			Preload("Identifications").
			Preload("RelatedParties").
			Preload("Characteristics").
			Preload("ExternalReferences").
			Preload("TaxExemptions").
			Preload("Attachments").
			Where("id = ?", id).First(&org.Party).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		type OrgSpecifics struct {
			TradingName      string
			IsLegalEntity    bool
			OrganizationType string
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
		org.OrganizationType = specifics.OrganizationType
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

		// 2. Update/Create Individual specifics
		individualSpecifics := map[string]any{
			"id":          ind.ID,
			"given_name":  ind.GivenName,
			"family_name": ind.FamilyName,
			"middle_name": ind.MiddleName,
			"birth_date":  ind.BirthDate,
			"gender":      ind.Gender,
		}

		// Check if exists in individuals
		var count int64
		tx.Table("individuals").Where("id = ?", ind.ID).Count(&count)
		if count > 0 {
			if err := tx.Table("individuals").Where("id = ?", ind.ID).Updates(individualSpecifics).Error; err != nil {
				return err
			}
		} else {
			// Migration: Create new individual record
			if err := tx.Table("individuals").Create(individualSpecifics).Error; err != nil {
				return err
			}
			// Clean up organization if exists
			if err := tx.Table("organizations").Where("id = ?", ind.ID).Delete(nil).Error; err != nil {
				return err
			}
		}

		// 3. Update Sub-resources (Replace Strategy)
		if err := r.updateSubResources(tx, ind.ID, &ind.Party); err != nil {
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

		// 2. Update/Create Organization specifics
		orgSpecifics := map[string]any{
			"id":                org.ID,
			"trading_name":      org.TradingName,
			"is_legal_entity":   org.IsLegalEntity,
			"organization_type": org.OrganizationType,
		}

		// Check if exists in organizations
		var count int64
		tx.Table("organizations").Where("id = ?", org.ID).Count(&count)
		if count > 0 {
			if err := tx.Table("organizations").Where("id = ?", org.ID).Updates(orgSpecifics).Error; err != nil {
				return err
			}
		} else {
			// Migration: Create new organization record
			if err := tx.Table("organizations").Create(orgSpecifics).Error; err != nil {
				return err
			}
			// Clean up individual if exists
			if err := tx.Table("individuals").Where("id = ?", org.ID).Delete(nil).Error; err != nil {
				return err
			}
		}

		// 3. Update Sub-resources (Replace Strategy)
		if err := r.updateSubResources(tx, org.ID, &org.Party); err != nil {
			return err
		}

		return nil
	})
}

func (r *PartyRepository) loadAttachmentContents(ctx context.Context, p *domain.Party) error {
	if len(p.Attachments) == 0 {
		return nil
	}
	for i := range p.Attachments {
		att := &p.Attachments[i]
		if att.RefType == "Internal" && att.RefID != "" {
			var content domain.AttachmentContent
			if err := GetTx(ctx, r.db).Table("attachment_contents").Where("id = ?", att.RefID).First(&content).Error; err != nil {
				continue
			}
			att.ContentData = content.Data
		}
	}
	return nil
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

func (r *PartyRepository) SearchParties(ctx context.Context, criteria map[string]any) ([]domain.Party, error) {
	var parties []domain.Party
	query := GetTx(ctx, r.db).Table("parties")

	// Apply filters
	if val, ok := criteria["id"]; ok {
		query = query.Where("parties.id = ?", val)
	}
	if val, ok := criteria["type"]; ok {
		query = query.Where("parties.type = ?", val)
	}
	if val, ok := criteria["status"]; ok {
		query = query.Where("parties.status = ?", val)
	}

	joinedIndividual := false
	joinedOrganization := false

	// Handle generic search if provided
	if searchVal, ok := criteria["search"]; ok {
		if !joinedIndividual {
			query = query.Joins("LEFT JOIN individuals ON individuals.id = parties.id")
			joinedIndividual = true
		}
		if !joinedOrganization {
			query = query.Joins("LEFT JOIN organizations ON organizations.id = parties.id")
			joinedOrganization = true
		}

		searchTerm := "%" + searchVal.(string) + "%"
		// Search across ID, Type, Names
		query = query.Where("(parties.id = ? OR parties.type ILIKE ? OR individuals.given_name ILIKE ? OR individuals.family_name ILIKE ? OR organizations.trading_name ILIKE ?)",
			searchVal, searchTerm, searchTerm, searchTerm, searchTerm)
	} else if nameVal, ok := criteria["name"]; ok {
		// Existing Name logic (only used if generic search is NOT present)
		if !joinedIndividual {
			query = query.Joins("LEFT JOIN individuals ON individuals.id = parties.id")
			joinedIndividual = true
		}
		if !joinedOrganization {
			query = query.Joins("LEFT JOIN organizations ON organizations.id = parties.id")
			joinedOrganization = true
		}

		searchTerm := "%" + nameVal.(string) + "%"
		query = query.Where("(individuals.given_name ILIKE ? OR individuals.family_name ILIKE ? OR organizations.trading_name ILIKE ?)", searchTerm, searchTerm, searchTerm)
	}

	// External Reference Search
	if extRefVal, ok := criteria["externalReference"]; ok {
		query = query.Joins("JOIN external_references ON external_references.party_id = parties.id").
			Where("external_references.external_reference_id = ?", extRefVal)
	}

	if val, ok := criteria["given_name"]; ok {
		if !joinedIndividual {
			query = query.Joins("JOIN individuals ON individuals.id = parties.id")
			joinedIndividual = true
		}
		query = query.Where("individuals.given_name = ?", val)
	}
	if val, ok := criteria["family_name"]; ok {
		if !joinedIndividual {
			query = query.Joins("JOIN individuals ON individuals.id = parties.id")
			// unique join
		}
		query = query.Where("individuals.family_name = ?", val)
	}

	if val, ok := criteria["trading_name"]; ok {
		if !joinedOrganization {
			query = query.Joins("JOIN organizations ON organizations.id = parties.id")
			joinedOrganization = true
		}
		query = query.Where("organizations.trading_name = ?", val)
	}

	if val, ok := criteria["is_legal_entity"]; ok {
		if !joinedOrganization {
			query = query.Joins("JOIN organizations ON organizations.id = parties.id")
			// unique join
		}
		query = query.Where("organizations.is_legal_entity = ?", val)
	}

	// Email lookup: match a party owning a contact medium of type "email" with
	// the given value. A subquery is used (rather than a JOIN) to avoid
	// duplicate party rows when a party has multiple contact mediums.
	if val, ok := criteria["email"]; ok {
		emailSubquery := GetTx(ctx, r.db).Table("party_contact_mediums").
			Select("party_id").
			Where("medium_type = ? AND value = ?", "email", val)
		query = query.Where("parties.id IN (?)", emailSubquery)
	}

	if err := query.Find(&parties).Error; err != nil {
		return nil, err
	}

	return parties, nil
}

// Helpers

func (r *PartyRepository) updateSubResources(tx *gorm.DB, partyID string, p *domain.Party) error {
	// ContactMediums
	if err := tx.Delete(&domain.ContactMedium{}, "party_id = ?", partyID).Error; err != nil {
		return err
	}
	if len(p.ContactMediums) > 0 {
		for i := range p.ContactMediums {
			p.ContactMediums[i].PartyID = partyID // Ensure link
		}
		if err := tx.Create(&p.ContactMediums).Error; err != nil {
			return err
		}
	}

	// Identifications
	if err := tx.Delete(&domain.Identification{}, "party_id = ?", partyID).Error; err != nil {
		return err
	}
	if len(p.Identifications) > 0 {
		for i := range p.Identifications {
			p.Identifications[i].PartyID = partyID
		}
		if err := tx.Create(&p.Identifications).Error; err != nil {
			return err
		}
	}

	// RelatedParties
	if err := tx.Delete(&domain.RelatedParty{}, "party_id = ?", partyID).Error; err != nil {
		return err
	}
	if len(p.RelatedParties) > 0 {
		for i := range p.RelatedParties {
			p.RelatedParties[i].PartyID = partyID
		}
		if err := tx.Create(&p.RelatedParties).Error; err != nil {
			return err
		}
	}

	// Characteristics
	if err := tx.Delete(&domain.PartyCharacteristic{}, "party_id = ?", partyID).Error; err != nil {
		return err
	}
	if len(p.Characteristics) > 0 {
		for i := range p.Characteristics {
			p.Characteristics[i].PartyID = partyID
		}
		if err := tx.Create(&p.Characteristics).Error; err != nil {
			return err
		}
	}

	// ExternalReferences
	if err := tx.Delete(&domain.ExternalReference{}, "party_id = ?", partyID).Error; err != nil {
		return err
	}
	if len(p.ExternalReferences) > 0 {
		for i := range p.ExternalReferences {
			p.ExternalReferences[i].PartyID = partyID
		}
		if err := tx.Create(&p.ExternalReferences).Error; err != nil {
			return err
		}
	}

	// TaxExemptions
	if err := tx.Delete(&domain.TaxExemption{}, "party_id = ?", partyID).Error; err != nil {
		return err
	}
	if len(p.TaxExemptions) > 0 {
		for i := range p.TaxExemptions {
			p.TaxExemptions[i].PartyID = partyID
		}
		if err := tx.Create(&p.TaxExemptions).Error; err != nil {
			return err
		}
	}

	// Attachments (Split Storage Strategy)
	if err := tx.Delete(&domain.Attachment{}, "owner_id = ?", partyID).Error; err != nil {
		return err
	}

	if len(p.Attachments) > 0 {
		for i := range p.Attachments {
			att := &p.Attachments[i]
			att.OwnerID = partyID

			if len(att.ContentData) > 0 {
				att.RefType = "Internal"

				att.RefID = att.ID

				content := domain.AttachmentContent{
					ID:   att.RefID,
					Data: att.ContentData,
				}
				if err := tx.Create(&content).Error; err != nil {
					return err
				}
			} else {
				if att.RefType == "" {
					att.RefType = "S3" // Default fallthrough
				}
			}
		}

		if err := tx.Create(&p.Attachments).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *PartyRepository) withUser(ctx context.Context, fn func(tx *gorm.DB) error) error {
	userID, _ := ctx.Value(domain.UserContextKey).(string)
	if userID == "" {
		userID = "unknown"
	}

	runWithUser := func(tx *gorm.DB) error {
		// Set local session variable for audit trigger
		if err := tx.Exec("SELECT set_config('app.current_user', ?, true)", userID).Error; err != nil {
			return err
		}
		return fn(tx)
	}

	// Check if we are already in a transaction
	if existingTx, ok := ctx.Value(TxKey).(*gorm.DB); ok {
		// We are in a transaction, reuse it
		return runWithUser(existingTx)
	}

	// No existing transaction, start new one
	db := GetTx(ctx, r.db)

	return db.Transaction(func(tx *gorm.DB) error {
		return runWithUser(tx)
	})
}
