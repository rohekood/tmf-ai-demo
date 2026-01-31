package repository

import (
	"time"
	"tmf/services/shopping-cart/internal/core/domain"
)

// DAO Definitions
type CartTable struct {
	ID                 string  `gorm:"primaryKey;type:uuid"`
	CustomerID         *string `gorm:"type:uuid"`
	Status             string  `gorm:"not null"`
	Version            int     `gorm:"not null;default:1"`
	TotalPriceAmount   float64 `gorm:"type:decimal(10,2)"`
	TotalPriceCurrency string  `gorm:"type:char(3)"`
	ValidForEnd        time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Items              []CartItemTable `gorm:"foreignKey:CartID"`
}

func (CartTable) TableName() string {
	return "carts"
}

type CartItemTable struct {
	ID            string            `gorm:"primaryKey;type:uuid"`
	CartID        string            `gorm:"not null;type:uuid"`
	OfferingID    string            `gorm:"not null;type:uuid"`
	Quantity      int               `gorm:"not null"`
	ProductConfig map[string]string `gorm:"serializer:json"`
	UnitAmount    float64           `gorm:"type:decimal(10,2)"`
	Currency      string            `gorm:"type:char(3)"`
	CreatedAt     time.Time
}

func (CartItemTable) TableName() string {
	return "cart_items"
}

type OutboxTable struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	Topic     string `gorm:"not null"`
	Payload   []byte `gorm:"type:jsonb;not null"`
	Status    string `gorm:"default:'PENDING'"`
	CreatedAt time.Time
}

func (OutboxTable) TableName() string {
	return "outbox_events"
}

type ProductPriceTable struct {
	ID         string  `gorm:"primaryKey;type:uuid"`
	UnitAmount float64 `gorm:"type:decimal(10,2)"`
	Currency   string  `gorm:"type:char(3)"`
	UpdatedAt  time.Time
}

func (ProductPriceTable) TableName() string {
	return "product_prices"
}

// Mappers

func toDomainCart(t *CartTable) *domain.Cart {
	items := make([]domain.CartItem, len(t.Items))
	for i, item := range t.Items {
		items[i] = domain.CartItem{
			ID:            item.ID,
			CartID:        item.CartID,
			OfferingID:    item.OfferingID,
			Quantity:      item.Quantity,
			ProductConfig: item.ProductConfig,
			UnitAmount:    item.UnitAmount,
			Currency:      item.Currency,
		}
	}
	return &domain.Cart{
		ID: t.ID,
		CustomerID: func() string {
			if t.CustomerID == nil {
				return ""
			}
			return *t.CustomerID
		}(),
		Status:             t.Status,
		Version:            t.Version,
		TotalPriceAmount:   t.TotalPriceAmount,
		TotalPriceCurrency: t.TotalPriceCurrency,
		ValidForEnd:        t.ValidForEnd,
		Items:              items,
	}
}

func toDAOCart(d *domain.Cart) *CartTable {
	items := make([]CartItemTable, len(d.Items))
	for i, item := range d.Items {
		items[i] = CartItemTable{
			ID:            item.ID,
			CartID:        item.CartID,
			OfferingID:    item.OfferingID,
			Quantity:      item.Quantity,
			ProductConfig: item.ProductConfig,
			UnitAmount:    item.UnitAmount,
			Currency:      item.Currency,
		}
	}

	var custID *string
	if d.CustomerID != "" {
		val := d.CustomerID
		custID = &val
	}

	return &CartTable{
		ID:                 d.ID,
		CustomerID:         custID,
		Status:             d.Status,
		Version:            d.Version,
		TotalPriceAmount:   d.TotalPriceAmount,
		TotalPriceCurrency: d.TotalPriceCurrency,
		ValidForEnd:        d.ValidForEnd,
		Items:              items,
	}
}

func toDAOOutbox(events []domain.OutboxEvent) []OutboxTable {
	tables := make([]OutboxTable, len(events))
	for i, e := range events {
		tables[i] = OutboxTable{
			ID:        e.ID,
			Topic:     e.Topic,
			Payload:   e.Payload,
			Status:    e.Status,
			CreatedAt: e.CreatedAt,
		}
	}
	return tables
}
