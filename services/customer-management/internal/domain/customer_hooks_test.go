package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCustomer_TableName(t *testing.T) {
	c := Customer{}
	assert.Equal(t, "customers", c.TableName())
}

func TestHooks_BeforeCreate_GeneratesID(t *testing.T) {
	// GORM hooks need a DB connection usually, but we can call them manually with nil tx if they don't use it
	// Customer sub-structs only check ID == ""

	// CustomerAccount
	ca := CustomerAccount{}
	_ = ca.BeforeCreate(nil)
	assert.NotEmpty(t, ca.ID)
	assert.Equal(t, "customer_accounts", ca.TableName())

	// CreditProfile
	cp := CreditProfile{}
	_ = cp.BeforeCreate(nil)
	assert.NotEmpty(t, cp.ID)
	assert.Equal(t, "credit_profiles", cp.TableName())

	// ContactMedium
	cm := ContactMedium{}
	_ = cm.BeforeCreate(nil)
	assert.NotEmpty(t, cm.ID)
	assert.Equal(t, "contact_mediums", cm.TableName())

	// CustomerCharacteristic
	cc := CustomerCharacteristic{}
	_ = cc.BeforeCreate(nil)
	assert.NotEmpty(t, cc.ID)
	assert.Equal(t, "customer_characteristics", cc.TableName())

	// PrivacyConsent
	pc := PrivacyConsent{}
	_ = pc.BeforeCreate(nil)
	assert.NotEmpty(t, pc.ID)
	assert.Equal(t, "privacy_consents", pc.TableName())

	// RelatedParty
	rp := RelatedParty{}
	_ = rp.BeforeCreate(nil)
	assert.NotEmpty(t, rp.ID)
	assert.Equal(t, "related_parties", rp.TableName())

	// PaymentMethod
	pm := PaymentMethod{}
	_ = pm.BeforeCreate(nil)
	assert.NotEmpty(t, pm.ID)
	assert.Equal(t, "payment_methods", pm.TableName())

	// MarketSegment
	ms := MarketSegment{}
	_ = ms.BeforeCreate(nil)
	assert.NotEmpty(t, ms.ID)
	assert.Equal(t, "market_segments", ms.TableName())

	// CustomerInteraction
	ci := CustomerInteraction{}
	_ = ci.BeforeCreate(nil)
	assert.NotEmpty(t, ci.ID)
	assert.Equal(t, "customer_interactions", ci.TableName())
}

// Mock DB for hooks that might use it (none currently do, but good practice)
func mockDB() *gorm.DB {
	return &gorm.DB{}
}
