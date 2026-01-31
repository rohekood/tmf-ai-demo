package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateEligibility(t *testing.T) {
	tests := []struct {
		name      string
		inPolygon bool
		ports     int
		expected  QualificationStatus
	}{
		{
			name:      "Happy path - Qualified",
			inPolygon: true,
			ports:     5,
			expected:  StatusQualified,
		},
		{
			name:      "Geo Failure - Unqualified",
			inPolygon: false,
			ports:     5,
			expected:  StatusUnqualified,
		},
		{
			name:      "Port Failure - Unqualified",
			inPolygon: true,
			ports:     0,
			expected:  StatusUnqualified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateEligibility(tt.inPolygon, tt.ports)
			assert.Equal(t, tt.expected, result)
		})
	}
}
