package rabbitmq

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_CreateParty_WithReferencesAndAttachments(t *testing.T) {
	suite := setupTestSuite(t)

	// Payload with all new gap fields
	payload := map[string]interface{}{
		"@type":      "Individual",
		"id":         "int-gaps-1",
		"givenName":  "Integration",
		"familyName": "Gaps",
		"externalReferences": []map[string]interface{}{
			{"externalSystemId": "CRM", "externalReferenceId": "C123"},
		},
		"taxExemptions": []map[string]interface{}{
			{"certificateNumber": "TAX-123", "issuingJurisdiction": "US"},
		},
		"attachments": []map[string]interface{}{
			{"name": "doc.txt", "content": []byte("hello world"), "attachmentType": "Note"},    // Implicit Internal
			{"name": "link.png", "url": "http://s3/bucket/img.png", "attachmentType": "Image"}, // Implicit S3
		},
	}
	body, _ := json.Marshal(payload)

	// Action
	err := suite.Handlers.HandleCreateParty(context.Background(), amqp091.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB
	saved, err := suite.Repo.GetIndividual(context.Background(), "int-gaps-1")
	require.NoError(t, err)

	// Check fields
	assert.Len(t, saved.ExternalReferences, 1)
	assert.Equal(t, "CRM", saved.ExternalReferences[0].ExternalSystemID)

	assert.Len(t, saved.TaxExemptions, 1)
	assert.Equal(t, "TAX-123", saved.TaxExemptions[0].CertificateNumber)

	assert.Len(t, saved.Attachments, 2)
	// Verify split logic in Repo via Handlers
	var internalCount, s3Count int
	for _, att := range saved.Attachments {
		if att.RefType == "Internal" {
			internalCount++
		}
		if att.RefType == "S3" {
			s3Count++
		}
	}
	assert.Equal(t, 1, internalCount)
	assert.Equal(t, 1, s3Count)

	// Verify Event
	evt := suite.waitForEvent(t, 5*time.Second)
	require.NotNil(t, evt, "Expected evt.party.created")
	assert.Equal(t, EvtPartyCreated, evt.RoutingKey)

	// Unmarshal event to check if sub-resources are present (if our event structure supports it)
	// For now, just verifying the event fired is good proof of flow completion.
}

func TestIntegration_SearchParty_ByExternalReference_RPC(t *testing.T) {
	suite := setupTestSuite(t)

	// 1. Create a Party with External Ref
	payloadCreate := map[string]interface{}{
		"@type":     "Individual",
		"id":        "int-search-ext-1",
		"givenName": "FoundMe",
		"externalReferences": []map[string]interface{}{
			{"externalSystemId": "Legacy", "externalReferenceId": "SEARCH-TARGET"},
		},
	}
	bodyCreate, _ := json.Marshal(payloadCreate)
	require.NoError(t, suite.Handlers.HandleCreateParty(context.Background(), amqp091.Delivery{Body: bodyCreate}))

	// 2. Setup Reply Queue for RPC
	replyQueue, err := suite.channel.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)
	replies, err := suite.channel.Consume(replyQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	// 3. Search Request
	payloadSearch := map[string]interface{}{
		"externalReference": "SEARCH-TARGET",
	}
	bodySearch, _ := json.Marshal(payloadSearch)

	err = suite.Handlers.HandleSearchParty(context.Background(), amqp091.Delivery{
		Body:          bodySearch,
		ReplyTo:       replyQueue.Name,
		CorrelationId: "rpc-123",
	})
	require.NoError(t, err)

	// 4. Verify Reply
	select {
	case reply := <-replies:
		var results []map[string]interface{}
		err := json.Unmarshal(reply.Body, &results)
		require.NoError(t, err)
		require.Len(t, results, 1)

		assert.Equal(t, "int-search-ext-1", results[0]["id"])
		assert.Equal(t, "FoundMe", results[0]["givenName"])

	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for RPC reply")
	}
}
