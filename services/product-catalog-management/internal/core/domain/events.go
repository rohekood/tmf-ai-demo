package domain

type CatalogCreateEvent struct {
	Name            string     `json:"name"`
	Description     string     `json:"description,omitempty"`
	ValidFor        TimePeriod `json:"validFor"`
	LifecycleStatus string     `json:"lifecycleStatus,omitempty"`
}

// Add other event structs as needed
type CategoryCreateEvent struct {
	Name            string     `json:"name"`
	Description     string     `json:"description,omitempty"`
	ParentID        *string    `json:"parentId,omitempty"`
	IsRoot          bool       `json:"isRoot"`
	CatalogID       *string    `json:"catalogId,omitempty"`
	ValidFor        TimePeriod `json:"validFor"`
	LifecycleStatus string     `json:"lifecycleStatus,omitempty"`
}

type ProductSpecificationCreateEvent struct {
	Name            string                               `json:"name"`
	ProductNumber   string                               `json:"productNumber,omitempty"`
	Description     string                               `json:"description,omitempty"`
	IsBundle        bool                                 `json:"isBundle"`
	LifecycleStatus string                               `json:"lifecycleStatus,omitempty"`
	ValidFor        TimePeriod                           `json:"validFor"`
	Characteristics map[string]ProductSpecCharacteristic `json:"characteristics,omitempty"`
}

type ProductOfferingCreateEvent struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	IsBundle        bool                   `json:"isBundle"`
	IsSellable      bool                   `json:"isSellable"`
	LifecycleStatus string                 `json:"lifecycleStatus,omitempty"`
	ValidFor        TimePeriod             `json:"validFor"`
	ProductSpecID   *string                `json:"productSpecId,omitempty"`
	CategoryIDs     []string               `json:"categoryIds,omitempty"`
	Prices          []ProductOfferingPrice `json:"productOfferingPrice,omitempty"`
	Attachments     []Attachment           `json:"attachments,omitempty"`
}

// Output Events

type CatalogCreatedEvent struct {
	Catalog *Catalog `json:"catalog"`
}

type CategoryCreatedEvent struct {
	Category *Category `json:"category"`
}

type ProductSpecificationCreatedEvent struct {
	ProductSpecification *ProductSpecification `json:"productSpecification"`
}

type ProductOfferingCreatedEvent struct {
	ProductOffering *ProductOffering `json:"productOffering"`
}

type CatalogUpdatedEvent struct {
	Catalog *Catalog `json:"catalog"`
}

type CatalogDeletedEvent struct {
	ID string `json:"id"`
}

type CategoryUpdatedEvent struct {
	Category *Category `json:"category"`
}

type CategoryDeletedEvent struct {
	ID string `json:"id"`
}

type ProductSpecificationUpdatedEvent struct {
	ProductSpecification *ProductSpecification `json:"productSpecification"`
}

type ProductSpecificationDeletedEvent struct {
	ID string `json:"id"`
}

type ProductOfferingUpdatedEvent struct {
	ProductOffering *ProductOffering `json:"productOffering"`
}

type ProductOfferingDeletedEvent struct {
	ID string `json:"id"`
}
