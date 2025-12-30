package postgres

import (
	"context"
	"testing"
	"time"
	"tmf/services/party-management/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIndividual_WithAttachments(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "ind-att-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Attachment",
		FamilyName: "Tester",
	}

	// Attachments (Mixed)
	attInternal := domain.Attachment{
		ID:             "11111111-1111-1111-1111-111111111111", // Valid UUID
		OwnerID:        ind.ID,
		Name:           "scan.pdf",
		MimeType:       "application/pdf",
		AttachmentType: "Identification",
		RefType:        "Internal",
		RefID:          "11111111-1111-1111-1111-111111111111", // Same ID for content
		ContentData:    []byte("fake-pdf-content"),
	}
	attS3 := domain.Attachment{
		ID:             "22222222-2222-2222-2222-222222222222", // Valid UUID
		OwnerID:        ind.ID,
		Name:           "photo.jpg",
		MimeType:       "image/jpeg",
		AttachmentType: "Avartar",
		RefType:        "S3",
		RefID:          "s3://bucket/key/photo.jpg",
	}
	ind.Attachments = []domain.Attachment{attInternal, attS3}

	// ACTION
	err := repo.CreateIndividual(ctx, ind)
	require.NoError(t, err)

	// VERIFICATION
	saved, err := repo.GetIndividual(ctx, "ind-att-1")
	require.NoError(t, err)

	assert.Len(t, saved.Attachments, 2)

	var internalAtt, s3Att domain.Attachment
	for _, att := range saved.Attachments {
		if att.RefType == "Internal" {
			internalAtt = att
			assert.Equal(t, "11111111-1111-1111-1111-111111111111", att.ID)
			assert.Equal(t, "11111111-1111-1111-1111-111111111111", att.RefID)
			assert.NotEmpty(t, att.ContentData) // Should load content
		} else { // S3
			s3Att = att
			assert.Equal(t, "22222222-2222-2222-2222-222222222222", att.ID)
			assert.Equal(t, "s3://bucket/key/photo.jpg", att.RefID)
			assert.Empty(t, att.ContentData)
		}
	}

	// Verify Content Storage directly in DB
	var contentCount int64
	db.Table("attachment_contents").Where("id = ?", internalAtt.RefID).Count(&contentCount)
	assert.Equal(t, int64(1), contentCount, "Internal attachment content should exist in attachment_contents table")

	db.Table("attachment_contents").Where("id = ?", s3Att.ID).Count(&contentCount)
	assert.Equal(t, int64(0), contentCount, "S3 attachment should NOT exist in attachment_contents table by Attachment ID")
}
