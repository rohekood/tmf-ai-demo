package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductOffering_Validate_Lifecycle(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"Draft", "Draft", false},
		{"Active", "Active", false},
		{"Retired", "Retired", false},
		{"Suspended", "Suspended", false},
		{"Invalid", "InvalidStatus", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := ProductOffering{Name: "O1", LifecycleStatus: tt.status}
			err := o.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProductOffering_Validate_Price(t *testing.T) {
	tests := []struct {
		name    string
		price   []ProductOfferingPrice
		wantErr bool
	}{
		{
			name: "Valid Price",
			price: []ProductOfferingPrice{
				{Price: Money{Value: 10, Unit: "EUR"}},
			},
			wantErr: false,
		},
		{
			name: "Negative Price",
			price: []ProductOfferingPrice{
				{Price: Money{Value: -10, Unit: "EUR"}},
			},
			wantErr: true,
		},
		{
			name: "Missing Currency",
			price: []ProductOfferingPrice{
				{Price: Money{Value: 10, Unit: ""}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := ProductOffering{
				Name:                 "O1",
				LifecycleStatus:      "Active",
				ProductOfferingPrice: tt.price,
			}
			err := o.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
