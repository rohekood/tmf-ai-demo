package postgres

import (
	"context"
	"testing"
	"tmf/services/party-management/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateIndividual_TypeMigration_OrgToInd(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	ctx := context.Background()

	org := &domain.Organization{
		Party: domain.Party{
			ID:     "type-migrate-1",
			Type:   domain.PartyTypeOrganization,
			Status: "Active",
		},
		TradingName: "OriginCorp",
	}
	require.NoError(t, repo.CreateOrganization(ctx, org))

	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "type-migrate-1",
			Type:   domain.PartyTypeIndividual,
			Status: "Migrated",
		},
		GivenName:  "Migrated",
		FamilyName: "Person",
	}
	err := repo.UpdateIndividual(ctx, ind)
	assert.NoError(t, err)

	fetched, err := repo.GetIndividual(ctx, "type-migrate-1")
	assert.NoError(t, err)
	assert.Equal(t, "Migrated", fetched.GivenName)
}

func TestUpdateOrganization_TypeMigration_IndToOrg(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	ctx := context.Background()

	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "type-migrate-2",
			Type:   domain.PartyTypeIndividual,
			Status: "Active",
		},
		GivenName: "OrigPerson",
	}
	require.NoError(t, repo.CreateIndividual(ctx, ind))

	org := &domain.Organization{
		Party: domain.Party{
			ID:     "type-migrate-2",
			Type:   domain.PartyTypeOrganization,
			Status: "Migrated",
		},
		TradingName:   "MigratedCorp",
		IsLegalEntity: true,
	}
	err := repo.UpdateOrganization(ctx, org)
	assert.NoError(t, err)

	fetched, err := repo.GetOrganization(ctx, "type-migrate-2")
	assert.NoError(t, err)
	assert.Equal(t, "MigratedCorp", fetched.TradingName)
}

func TestUpdateSubResources_AllTypes(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	ctx := context.Background()
	erID := uuid.New().String()
	teID := uuid.New().String()
	attID := uuid.New().String()

	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "sub-all-1",
			Type:   domain.PartyTypeIndividual,
			Status: "Active",
			ContactMediums: []domain.ContactMedium{
				{ID: "cm-sub-1", PartyID: "sub-all-1", MediumType: "email", Value: "a@b.com"},
				{ID: "cm-sub-2", PartyID: "sub-all-1", MediumType: "phone", Value: "555-0001"},
			},
			Identifications: []domain.Identification{
				{ID: "id-sub-1", PartyID: "sub-all-1", IdentificationType: "passport", IdentificationID: "P001"},
			},
			RelatedParties: []domain.RelatedParty{
				{ID: "rp-sub-1", PartyID: "sub-all-1", RelatedPartyID: "other-1", RelatedPartyName: "Related"},
			},
			Characteristics: []domain.PartyCharacteristic{
				{ID: "ch-sub-1", PartyID: "sub-all-1", Name: "lang", Value: "en"},
				{ID: "ch-sub-2", PartyID: "sub-all-1", Name: "pref", Value: "dark"},
			},
			ExternalReferences: []domain.ExternalReference{
				{ID: erID, PartyID: "sub-all-1", ExternalSystemID: "CRM", ExternalReferenceID: "CRM-1"},
			},
			TaxExemptions: []domain.TaxExemption{
				{ID: teID, PartyID: "sub-all-1", CertificateNumber: "TE001"},
			},
			Attachments: []domain.Attachment{
				{
					ID:          attID,
					OwnerID:     "sub-all-1",
					Name:        "doc.pdf",
					MimeType:    "application/pdf",
					ContentData: []byte("pdf-data-here"),
				},
			},
		},
		GivenName:  "SubAll",
		FamilyName: "Test",
	}
	err := repo.CreateIndividual(ctx, ind)
	require.NoError(t, err)

	party, err := repo.GetParty(ctx, "sub-all-1")
	require.NoError(t, err)
	assert.Len(t, party.ContactMediums, 2)
	assert.Len(t, party.Identifications, 1)
	assert.Len(t, party.RelatedParties, 1)
	assert.Len(t, party.Characteristics, 2)
	assert.Len(t, party.ExternalReferences, 1)
	assert.Len(t, party.TaxExemptions, 1)
	assert.Len(t, party.Attachments, 1)

	erID2 := uuid.New().String()
	teID2 := uuid.New().String()
	attID2 := uuid.New().String()

	ind.ContactMediums = []domain.ContactMedium{
		{ID: "cm-sub-3", PartyID: "sub-all-1", MediumType: "postal", Value: "123 Street"},
	}
	ind.Identifications = nil
	ind.RelatedParties = nil
	ind.Characteristics = []domain.PartyCharacteristic{
		{ID: "ch-sub-3", PartyID: "sub-all-1", Name: "updated", Value: "yes"},
	}
	ind.ExternalReferences = []domain.ExternalReference{
		{ID: erID2, PartyID: "sub-all-1", ExternalSystemID: "SAP", ExternalReferenceID: "SAP-1"},
	}
	ind.TaxExemptions = []domain.TaxExemption{
		{ID: teID2, PartyID: "sub-all-1", CertificateNumber: "TE002"},
	}
	ind.Attachments = []domain.Attachment{
		{
			ID:       attID2,
			OwnerID:  "sub-all-1",
			Name:     "new.txt",
			MimeType: "text/plain",
			RefType:  "S3",
			RefID:    "s3://bucket/key",
		},
	}
	err = repo.UpdateIndividual(ctx, ind)
	require.NoError(t, err)

	party2, err := repo.GetParty(ctx, "sub-all-1")
	require.NoError(t, err)
	assert.Len(t, party2.ContactMediums, 1)
	assert.Len(t, party2.Identifications, 0)
	assert.Len(t, party2.RelatedParties, 0)
	assert.Len(t, party2.Characteristics, 1)
	assert.Len(t, party2.ExternalReferences, 1)
	assert.Len(t, party2.TaxExemptions, 1)
	assert.Len(t, party2.Attachments, 1)
}

func TestCreateWithAttachment_S3RefType(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	attID := uuid.New().String()

	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "att-s3-1",
			Type:   domain.PartyTypeIndividual,
			Status: "Active",
			Attachments: []domain.Attachment{
				{
					ID:       attID,
					OwnerID:  "att-s3-1",
					Name:     "remote.jpg",
					MimeType: "image/jpeg",
					RefType:  "S3",
					RefID:    "s3://bucket/images/photo.jpg",
				},
			},
		},
		GivenName: "S3User",
	}
	err := repo.CreateIndividual(context.Background(), ind)
	assert.NoError(t, err)
}

func TestCreateWithAttachment_DefaultRefType(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	attID := uuid.New().String()

	ind := &domain.Individual{
		Party: domain.Party{
			ID:     "att-default-1",
			Type:   domain.PartyTypeIndividual,
			Status: "Active",
			Attachments: []domain.Attachment{
				{
					ID:       attID,
					OwnerID:  "att-default-1",
					Name:     "noref.bin",
					MimeType: "application/octet-stream",
				},
			},
		},
		GivenName: "DefaultRef",
	}
	err := repo.CreateIndividual(context.Background(), ind)
	assert.NoError(t, err)
}

func TestGetOrganization_NotFound(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	_, err := repo.GetOrganization(context.Background(), "nonexistent-org")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetIndividual_NotFound(t *testing.T) {
	db, _ := setupTestDB(t)
	if db == nil {
		return
	}

	repo := NewPartyRepository(db)
	_, err := repo.GetIndividual(context.Background(), "nonexistent-ind")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
