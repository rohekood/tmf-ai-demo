package repository

import (
	"encoding/json"
	"time"
	"tmf/services/product-catalog-management/internal/core/domain"
)

type CatalogModel struct {
	ID              string `gorm:"primaryKey"`
	Name            string `gorm:"not null"`
	Description     string
	ValidForStart   *time.Time
	ValidForEnd     *time.Time
	LastUpdate      time.Time
	LifecycleStatus string
}

func (m *CatalogModel) TableName() string {
	return "catalogs"
}

func (m *CatalogModel) ToDomain() *domain.Catalog {
	return &domain.Catalog{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		ValidFor: domain.TimePeriod{
			StartDateTime: m.ValidForStart,
			EndDateTime:   m.ValidForEnd,
		},
		LastUpdate:      m.LastUpdate,
		LifecycleStatus: m.LifecycleStatus,
	}
}

func FromDomainCatalog(d *domain.Catalog) *CatalogModel {
	return &CatalogModel{
		ID:              d.ID,
		Name:            d.Name,
		Description:     d.Description,
		ValidForStart:   d.ValidFor.StartDateTime,
		ValidForEnd:     d.ValidFor.EndDateTime,
		LastUpdate:      d.LastUpdate,
		LifecycleStatus: d.LifecycleStatus,
	}
}

type CategoryModel struct {
	ID              string `gorm:"primaryKey"`
	Name            string `gorm:"not null"`
	Description     string
	ParentID        *string
	IsRoot          bool
	CatalogID       *string
	ValidForStart   *time.Time
	ValidForEnd     *time.Time
	LastUpdate      time.Time
	LifecycleStatus string
}

func (m *CategoryModel) TableName() string {
	return "categories"
}

func (m *CategoryModel) ToDomain() *domain.Category {
	return &domain.Category{
		ID:              m.ID,
		Name:            m.Name,
		Description:     m.Description,
		ParentID:        m.ParentID,
		IsRoot:          m.IsRoot,
		CatalogID:       m.CatalogID,
		ValidFor:        domain.TimePeriod{StartDateTime: m.ValidForStart, EndDateTime: m.ValidForEnd},
		LastUpdate:      m.LastUpdate,
		LifecycleStatus: m.LifecycleStatus,
	}
}

func FromDomainCategory(d *domain.Category) *CategoryModel {
	return &CategoryModel{
		ID:              d.ID,
		Name:            d.Name,
		Description:     d.Description,
		ParentID:        d.ParentID,
		IsRoot:          d.IsRoot,
		CatalogID:       d.CatalogID,
		ValidForStart:   d.ValidFor.StartDateTime,
		ValidForEnd:     d.ValidFor.EndDateTime,
		LastUpdate:      d.LastUpdate,
		LifecycleStatus: d.LifecycleStatus,
	}
}

type ProductSpecificationModel struct {
	ID              string `gorm:"primaryKey"`
	Name            string `gorm:"not null"`
	Description     string
	ProductNumber   string
	LifecycleStatus string
	ValidForStart   *time.Time
	ValidForEnd     *time.Time
	LastUpdate      time.Time
	Characteristics []byte `gorm:"type:jsonb"`
}

func (m *ProductSpecificationModel) TableName() string {
	return "product_specifications"
}

func (m *ProductSpecificationModel) ToDomain() *domain.ProductSpecification {
	spec := &domain.ProductSpecification{
		ID:              m.ID,
		Name:            m.Name,
		Description:     m.Description,
		ProductNumber:   m.ProductNumber,
		LifecycleStatus: m.LifecycleStatus,
		ValidFor:        domain.TimePeriod{StartDateTime: m.ValidForStart, EndDateTime: m.ValidForEnd},
		LastUpdate:      m.LastUpdate,
	}
	if len(m.Characteristics) > 0 {
		_ = json.Unmarshal(m.Characteristics, &spec.Characteristics)
	}
	return spec
}

func FromDomainSpecification(d *domain.ProductSpecification) *ProductSpecificationModel {
	chars, _ := json.Marshal(d.Characteristics)
	return &ProductSpecificationModel{
		ID:              d.ID,
		Name:            d.Name,
		Description:     d.Description,
		ProductNumber:   d.ProductNumber,
		LifecycleStatus: d.LifecycleStatus,
		ValidForStart:   d.ValidFor.StartDateTime,
		ValidForEnd:     d.ValidFor.EndDateTime,
		LastUpdate:      d.LastUpdate,
		Characteristics: chars,
	}
}

type ProductOfferingModel struct {
	ID                     string `gorm:"primaryKey"`
	Name                   string `gorm:"not null"`
	Description            string
	LifecycleStatus        string
	ValidForStart          *time.Time
	ValidForEnd            *time.Time
	LastUpdate             time.Time
	IsBundle               bool
	IsSellable             bool
	ProductSpecificationID *string
	ProductOfferingPrice   []byte `gorm:"type:jsonb"`
	CategoryIDs            []byte `gorm:"type:jsonb"`
	Attachments            []byte `gorm:"type:jsonb"`
}

func (m *ProductOfferingModel) TableName() string {
	return "product_offerings"
}

func (m *ProductOfferingModel) ToDomain() *domain.ProductOffering {
	offering := &domain.ProductOffering{
		ID:                     m.ID,
		Name:                   m.Name,
		Description:            m.Description,
		LifecycleStatus:        m.LifecycleStatus,
		ValidFor:               domain.TimePeriod{StartDateTime: m.ValidForStart, EndDateTime: m.ValidForEnd},
		LastUpdate:             m.LastUpdate,
		IsBundle:               m.IsBundle,
		IsSellable:             m.IsSellable,
		ProductSpecificationID: m.ProductSpecificationID,
	}
	if len(m.ProductOfferingPrice) > 0 {
		_ = json.Unmarshal(m.ProductOfferingPrice, &offering.ProductOfferingPrice)
	}
	if len(m.CategoryIDs) > 0 {
		_ = json.Unmarshal(m.CategoryIDs, &offering.CategoryIDs)
	}
	if len(m.Attachments) > 0 {
		_ = json.Unmarshal(m.Attachments, &offering.Attachments)
	}
	return offering
}

func FromDomainOffering(d *domain.ProductOffering) *ProductOfferingModel {
	prices, _ := json.Marshal(d.ProductOfferingPrice)
	catIDs, _ := json.Marshal(d.CategoryIDs)
	attachments, _ := json.Marshal(d.Attachments)
	return &ProductOfferingModel{
		ID:                     d.ID,
		Name:                   d.Name,
		Description:            d.Description,
		LifecycleStatus:        d.LifecycleStatus,
		ValidForStart:          d.ValidFor.StartDateTime,
		ValidForEnd:            d.ValidFor.EndDateTime,
		LastUpdate:             d.LastUpdate,
		IsBundle:               d.IsBundle,
		IsSellable:             d.IsSellable,
		ProductSpecificationID: d.ProductSpecificationID,
		ProductOfferingPrice:   prices,
		CategoryIDs:            catIDs,
		Attachments:            attachments,
	}
}
