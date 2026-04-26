package rabbitmq

import (
	"encoding/json"
	"testing"
	"tmf/services/party-management/internal/domain"

	"github.com/stretchr/testify/assert"
)

// TestPartySerializationContract verifies that the JSON serialization of domain objects
// matches the contract expected by the UI and other consumers.
// Specifically, it checks for proper camelCase keys and the presence of @type.
func TestPartySerializationContract(t *testing.T) {
	t.Run("Individual serialization", func(t *testing.T) {
		ind := &domain.Individual{
			Party: domain.Party{
				ID:   "123",
				Type: domain.PartyTypeIndividual,
				Href: "http://example.com/123",
			},
			GivenName:  "Test",
			FamilyName: "User",
		}

		data, err := json.Marshal(ind)
		assert.NoError(t, err)

		var asMap map[string]any
		err = json.Unmarshal(data, &asMap)
		assert.NoError(t, err)

		// Assertions based on what the UI expects
		assert.Equal(t, "123", asMap["id"])
		assert.Equal(t, "Individual", asMap["@type"], "Missing or incorrect @type field")
		assert.Equal(t, "Test", asMap["givenName"])
		assert.Equal(t, "User", asMap["familyName"])
		// Ensure internal GORM fields are not leaked if not tagged (though they might be embedded)
		// With strict JSON tags we control this.
	})

	t.Run("Organization serialization", func(t *testing.T) {
		org := &domain.Organization{
			Party: domain.Party{
				ID:   "456",
				Type: domain.PartyTypeOrganization,
			},
			TradingName:   "Acme Corp",
			IsLegalEntity: true,
		}

		data, err := json.Marshal(org)
		assert.NoError(t, err)

		var asMap map[string]any
		err = json.Unmarshal(data, &asMap)
		assert.NoError(t, err)

		assert.Equal(t, "456", asMap["id"])
		assert.Equal(t, "Organization", asMap["@type"], "Missing or incorrect @type field")
		assert.Equal(t, "Acme Corp", asMap["tradingName"])
		assert.Equal(t, true, asMap["isLegalEntity"])
	})
}
