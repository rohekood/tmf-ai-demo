package rabbitmq

import (
	"context"
	"encoding/json"
	"testing"
	"tmf/services/party-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleCreateParty_MappingLogic(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)

	h := NewHandlers(mockRepo, nil, nil, &NoOpTransactionManager{})

	// Construct flat JSON payload as expected by handler
	flatPayload := map[string]interface{}{
		"@type":      "Individual",
		"id":         "unit-gaps-1",
		"givenName":  "Unit",
		"familyName": "Test",
		"externalReferences": []map[string]interface{}{
			{"externalSystemId": "Sys1", "externalReferenceId": "Ref1"},
		},
		"taxExemptions": []map[string]interface{}{
			{"certificateNumber": "CERT-999"},
		},
		"attachments": []map[string]interface{}{
			{"name": "note.txt", "content": []byte("content"), "refType": "Internal"},
		},
	}
	body, _ := json.Marshal(flatPayload)

	// Expectation
	mockRepo.On("CreateIndividual", mock.Anything, mock.MatchedBy(func(ind *domain.Individual) bool {
		if ind.ID != "unit-gaps-1" {
			return false
		}
		if len(ind.ExternalReferences) != 1 {
			return false
		}
		if ind.ExternalReferences[0].ExternalSystemID != "Sys1" {
			return false
		}

		if len(ind.TaxExemptions) != 1 {
			return false
		}
		if ind.TaxExemptions[0].CertificateNumber != "CERT-999" {
			return false
		}

		if len(ind.Attachments) != 1 {
			return false
		}
		if ind.Attachments[0].Name != "note.txt" {
			return false
		}

		return true
	})).Return(nil)

	// Execute
	err := h.HandleCreateParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestHandlers_MapHelpers(t *testing.T) {
	h := &Handlers{} // Empty handler is fine for helpers

	t.Run("mapExternalReferences", func(t *testing.T) {
		dtos := []ExternalReferenceDTO{
			{ExternalSystemID: "SysA", ExternalReferenceID: "RefA"},
		}
		result := h.mapExternalReferences(dtos, "p-1")

		require.Len(t, result, 1)
		assert.Equal(t, "SysA", result[0].ExternalSystemID)
		assert.Equal(t, "RefA", result[0].ExternalReferenceID)
		assert.Equal(t, "p-1", result[0].PartyID)
		assert.NotEmpty(t, result[0].ID)
	})

	t.Run("mapTaxExemptions", func(t *testing.T) {
		dtos := []TaxExemptionDTO{
			{CertificateNumber: "C1", IssuingJurisdiction: "J1"},
		}
		result := h.mapTaxExemptions(dtos, "p-1")

		require.Len(t, result, 1)
		assert.Equal(t, "C1", result[0].CertificateNumber)
		assert.NotEmpty(t, result[0].ID)
	})

	t.Run("mapAttachments", func(t *testing.T) {
		dtos := []AttachmentDTO{
			{Name: "f1", Content: []byte("abc"), RefType: ""}, // Should default to Internal? No, logic says if URL empty -> Internal
			{Name: "f2", URL: "http://s3", RefType: ""},       // Should default to S3
		}
		result := h.mapAttachments(dtos, "p-1")

		require.Len(t, result, 2)

		// Check f1
		assert.Equal(t, "Internal", result[0].RefType)
		assert.Equal(t, []byte("abc"), result[0].ContentData)

		// Check f2
		assert.Equal(t, "S3", result[1].RefType)
		assert.Equal(t, "http://s3", result[1].RefID)
	})

	t.Run("mapRelatedParties_Permissions", func(t *testing.T) {
		dtos := []RelatedPartyDTO{
			{RelatedPartyName: "Rel1", Permissions: []string{"READ", "WRITE"}},
		}
		result := h.mapRelatedParties(dtos, "p-1")

		require.Len(t, result, 1)
		assert.Equal(t, []string{"READ", "WRITE"}, result[0].Permissions)
	})
}
