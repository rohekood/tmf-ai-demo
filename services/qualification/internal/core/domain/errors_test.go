package domain_test

import (
	"errors"
	"testing"

	"tmf/services/qualification/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestDomainErrors(t *testing.T) {
	err1 := domain.NewInvalidInputError("bad input", nil)
	assert.Equal(t, "[INVALID_INPUT] bad input", err1.Error())
	assert.Nil(t, errors.Unwrap(err1))

	inner := errors.New("inner")
	err2 := domain.NewBackendError("backend failed", inner)
	assert.Equal(t, "[BACKEND_FAILURE] backend failed: inner", err2.Error())
	assert.Equal(t, inner, errors.Unwrap(err2))
}
