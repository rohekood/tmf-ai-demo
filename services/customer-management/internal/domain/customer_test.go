package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCustomer_IsActive(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		customer Customer
		expected bool
	}{
		{
			name: "Active customer no end date",
			customer: Customer{
				Status:        CustomerStatusActive,
				ValidForStart: now.Add(-time.Hour),
				ValidForEnd:   nil,
			},
			expected: true,
		},
		{
			name: "Active customer within dates",
			customer: Customer{
				Status:        CustomerStatusActive,
				ValidForStart: now.Add(-time.Hour),
				ValidForEnd:   func() *time.Time { t := now.Add(time.Hour); return &t }(),
			},
			expected: true,
		},
		{
			name: "Suspended customer",
			customer: Customer{
				Status: CustomerStatusSuspended,
			},
			expected: false,
		},
		{
			name: "Expired customer",
			customer: Customer{
				Status:        CustomerStatusActive,
				ValidForStart: now.Add(-2 * time.Hour),
				ValidForEnd:   func() *time.Time { t := now.Add(-time.Hour); return &t }(),
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.customer.IsActive(now))
		})
	}
}

// Sub-task: Add IsActive method to Customer struct in customer.go
