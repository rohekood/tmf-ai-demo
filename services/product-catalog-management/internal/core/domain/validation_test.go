package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCatalog_Validate(t *testing.T) {
	tests := []struct {
		name    string
		catalog Catalog
		wantErr bool
	}{
		{
			name: "Valid Catalog",
			catalog: Catalog{
				ID:              "cat-1",
				Name:            "Main Catalog",
				LifecycleStatus: "Active",
			},
			wantErr: false,
		},
		{
			name: "Empty Name",
			catalog: Catalog{
				ID:              "cat-2",
				Name:            "",
				LifecycleStatus: "Active",
			},
			wantErr: true,
		},
		{
			name: "Empty Status",
			catalog: Catalog{
				ID:              "cat-3",
				Name:            "Catalog 3",
				LifecycleStatus: "",
			},
			wantErr: false, // Currently logic only checks Name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.catalog.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCategory_Validate(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		wantErr  bool
	}{
		{
			name: "Valid Category",
			category: Category{
				ID:              "c-1",
				Name:            "Phones",
				LifecycleStatus: "Active",
			},
			wantErr: false,
		},
		{
			name: "Empty Name",
			category: Category{
				ID:              "c-2",
				Name:            "",
				LifecycleStatus: "Active",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.category.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProductSpecification_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spec    ProductSpecification
		wantErr bool
	}{
		{
			name: "Valid Spec",
			spec: ProductSpecification{
				ID:              "s-1",
				Name:            "Fiber Spec",
				LifecycleStatus: "Active",
				ProductNumber:   "PN-123",
			},
			wantErr: false,
		},
		{
			name: "Empty Name",
			spec: ProductSpecification{
				ID:              "s-2",
				Name:            "",
				LifecycleStatus: "Active",
				ProductNumber:   "PN-123",
			},
			wantErr: true,
		},
		{
			name: "Empty ProductNumber",
			spec: ProductSpecification{
				ID:              "s-3",
				Name:            "Spec 3",
				LifecycleStatus: "Active",
				ProductNumber:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProductOffering_Validate(t *testing.T) {
	tests := []struct {
		name     string
		offering ProductOffering
		wantErr  bool
	}{
		{
			name: "Valid Offering",
			offering: ProductOffering{
				ID:                     "o-1",
				Name:                   "Fiber 100",
				LifecycleStatus:        "Active",
				ProductSpecificationID: func() *string { s := "spec-1"; return &s }(),
			},
			wantErr: false,
		},
		{
			name: "Empty Name",
			offering: ProductOffering{
				ID:                     "o-2",
				Name:                   "",
				LifecycleStatus:        "Active",
				ProductSpecificationID: func() *string { s := "spec-1"; return &s }(),
			},
			wantErr: true,
		},
		{
			name: "Missing Spec ID",
			offering: ProductOffering{
				ID:              "o-3",
				Name:            "Offering 3",
				LifecycleStatus: "Active",
			},
			wantErr: false, // Currently not checked
		},
		{
			name: "Invalid Dates",
			offering: ProductOffering{
				ID:                     "o-4",
				Name:                   "Offering 4",
				LifecycleStatus:        "Active",
				ProductSpecificationID: func() *string { s := "spec-1"; return &s }(),
				ValidFor: TimePeriod{
					StartDateTime: func() *time.Time { t := time.Now().Add(time.Hour); return &t }(),
					EndDateTime:   func() *time.Time { t := time.Now(); return &t }(),
				},
			},
			wantErr: false, // TimePeriod logic not in Validate()
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.offering.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
