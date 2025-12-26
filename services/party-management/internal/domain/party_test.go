package domain_test

import (
	"testing"
	"time"

	"tmf/services/party-management/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestPartyTypes(t *testing.T) {
	assert.Equal(t, domain.PartyType("Individual"), domain.PartyTypeIndividual)
	assert.Equal(t, domain.PartyType("Organization"), domain.PartyTypeOrganization)
}

func TestIndividualCreation(t *testing.T) {
	now := time.Now()
	ind := domain.Individual{
		Party: domain.Party{
			ID:        "ind-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		},
		GivenName:  "John",
		FamilyName: "Doe",
	}

	assert.Equal(t, "ind-1", ind.ID)
	assert.Equal(t, domain.PartyTypeIndividual, ind.Type)
	assert.Equal(t, "John", ind.GivenName)
	assert.Equal(t, "parties", ind.Party.TableName())
}

func TestOrganizationCreation(t *testing.T) {
	now := time.Now()
	org := domain.Organization{
		Party: domain.Party{
			ID:        "org-1",
			Type:      domain.PartyTypeOrganization,
			Status:    "initialized",
			CreatedAt: now,
			UpdatedAt: now,
		},
		TradingName:   "Acme Corp",
		IsLegalEntity: true,
	}

	assert.Equal(t, "org-1", org.ID)
	assert.Equal(t, domain.PartyTypeOrganization, org.Type)
	assert.Equal(t, "Acme Corp", org.TradingName)
	assert.True(t, org.IsLegalEntity)
}
