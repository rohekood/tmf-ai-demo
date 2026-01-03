package domain

import (
	"time"
)

type ProductSpecification struct {
	ID              string                               `json:"id"`
	Name            string                               `json:"name"`
	Description     string                               `json:"description,omitempty"`
	ProductNumber   string                               `json:"productNumber"`
	IsBundle        bool                                 `json:"isBundle"`
	LifecycleStatus string                               `json:"lifecycleStatus"` // e.g., "Active", "Retired"
	ValidFor        TimePeriod                           `json:"validFor"`
	LastUpdate      time.Time                            `json:"lastUpdate"`
	Characteristics map[string]ProductSpecCharacteristic `json:"characteristics,omitempty"`
}

type ProductSpecCharacteristic struct {
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	ValueType    string   `json:"valueType"` // string, number, boolean
	Configurable bool     `json:"configurable"`
	ValidValues  []string `json:"validValues,omitempty"`
}

func (s *ProductSpecification) Validate() error {
	if s.Name == "" || s.ProductNumber == "" {
		return ErrInvalidInput
	}
	return nil
}
