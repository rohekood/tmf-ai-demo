package offering

import (
	"testing"
	"tmf/services/product-catalog-management/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestValidateLifecycleTransition(t *testing.T) {
	// Draft -> Active
	assert.NoError(t, validateLifecycleTransition(domain.LifecycleStatusDraft, domain.LifecycleStatusActive))
	assert.NoError(t, validateLifecycleTransition(domain.LifecycleStatusDraft, domain.LifecycleStatusRetired))
	
	// Active -> Suspended
	assert.NoError(t, validateLifecycleTransition(domain.LifecycleStatusActive, domain.LifecycleStatusSuspended))
	assert.NoError(t, validateLifecycleTransition(domain.LifecycleStatusActive, domain.LifecycleStatusRetired))

	// Suspended -> Active
	assert.NoError(t, validateLifecycleTransition(domain.LifecycleStatusSuspended, domain.LifecycleStatusActive))
	assert.NoError(t, validateLifecycleTransition(domain.LifecycleStatusSuspended, domain.LifecycleStatusRetired))

	// Retired -> Anything
	assert.Error(t, validateLifecycleTransition(domain.LifecycleStatusActive, domain.LifecycleStatusDraft))
	assert.Error(t, validateLifecycleTransition(domain.LifecycleStatusSuspended, domain.LifecycleStatusDraft))


	// Migration case
	assert.NoError(t, validateLifecycleTransition("", domain.LifecycleStatusActive))
	assert.Error(t, validateLifecycleTransition("UnknownStatus", domain.LifecycleStatusActive)) // Will fall through and return ErrInvalidInput
}
