package domain

import (
	"time"
)

type ProductOffering struct {
	ID                     string                 `json:"id"`
	Name                   string                 `json:"name"`
	Description            string                 `json:"description,omitempty"`
	LifecycleStatus        string                 `json:"lifecycleStatus"`
	ValidFor               TimePeriod             `json:"validFor"`
	LastUpdate             time.Time              `json:"lastUpdate"`
	IsBundle               bool                   `json:"isBundle"`
	IsSellable             bool                   `json:"isSellable"`
	ProductSpecificationID *string                `json:"productSpecificationId,omitempty"`
	ProductOfferingPrice   []ProductOfferingPrice `json:"productOfferingPrice,omitempty"`
	CategoryIDs            []string               `json:"categoryIds,omitempty"` // simplified link
	Attachments            []Attachment           `json:"attachments,omitempty"`

	// Enriched Data (Populated if requested)
	ProductSpecification *ProductSpecification `json:"productSpecification,omitempty"`
	Categories           []Category            `json:"categories,omitempty"`
}

type Attachment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	Type        string `json:"type"` // e.g. "Picture", "Document"
}

type ProductOfferingPrice struct {
	ID              string           `json:"id"`        // optional, might be embedded
	PriceType       string           `json:"priceType"` // recurring, one_time, usage
	Price           Money            `json:"price"`
	UnitOfMeasure   string           `json:"unitOfMeasure,omitempty"` // e.g. "month", "GB"
	PriceAlteration *PriceAlteration `json:"priceAlteration,omitempty"`
}

type PriceAlteration struct {
	Name string `json:"name"`
	Type string `json:"type"` // discount, fee
}

func (o *ProductOffering) Validate() error {
	if o.Name == "" {
		return ErrInvalidInput
	}
	return nil
}
