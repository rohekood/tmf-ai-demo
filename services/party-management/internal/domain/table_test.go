package domain

import (
	"testing"
)

func TestTableNames(t *testing.T) {
	if new(Party).TableName() != "parties" {
		t.Errorf("Party TableName wrong")
	}
	if new(Individual).TableName() != "parties" {
		t.Errorf("Individual TableName wrong")
	}
	if new(Organization).TableName() != "parties" {
		t.Errorf("Organization TableName wrong")
	}
	if new(ContactMedium).TableName() != "party_contact_mediums" {
		t.Errorf("ContactMedium TableName wrong")
	}
	if new(Identification).TableName() != "identifications" {
		t.Errorf("Identification TableName wrong")
	}
	if new(RelatedParty).TableName() != "related_parties" {
		t.Errorf("RelatedParty TableName wrong")
	}
	if new(PartyCharacteristic).TableName() != "party_characteristics" {
		t.Errorf("PartyCharacteristic TableName wrong")
	}
	if new(ExternalReference).TableName() != "external_references" {
		t.Errorf("ExternalReference TableName wrong")
	}
	if new(TaxExemption).TableName() != "party_tax_exemptions" {
		t.Errorf("TaxExemption TableName wrong")
	}
	if new(Attachment).TableName() != "party_attachments" {
		t.Errorf("Attachment TableName wrong")
	}
	if new(AttachmentContent).TableName() != "attachment_contents" {
		t.Errorf("AttachmentContent TableName wrong")
	}
}

func TestOutboxTableName(t *testing.T) {
	if new(OutboxEvent).TableName() != "outbox_events" {
		t.Errorf("Outbox TableName wrong")
	}
}
