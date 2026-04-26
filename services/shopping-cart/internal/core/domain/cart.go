package domain

import (
	"encoding/json"
	"time"
)

type Cart struct {
	ID                 string     `json:"id" gorm:"primaryKey"`
	CustomerID         string     `json:"customerId"`
	Items              []CartItem `json:"items" gorm:"foreignKey:CartID"`
	Status             string     `json:"status"` // Active, Closed
	TotalPrice         float64    `json:"totalPrice" gorm:"-"`
	TotalPriceAmount   float64    `json:"totalPriceAmount"`
	TotalPriceCurrency string     `json:"totalPriceCurrency"`
	Version            int        `json:"version"`
	ValidForEnd        time.Time  `json:"validForEnd"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

const (
	CartStatusActive  = "Active"
	CartStatusPricing = "Pricing"
	CartStatusClosed  = "Closed"
)

type CartItem struct {
	ID            string            `json:"id" gorm:"primaryKey"`
	CartID        string            `json:"cartId"`
	OfferingID    string            `json:"offeringId"`
	Quantity      int               `json:"quantity"`
	UnitPrice     float64           `json:"unitPrice" gorm:"-"`
	UnitAmount    float64           `json:"unitAmount"`
	Currency      string            `json:"currency"`
	ProductConfig map[string]string `json:"productConfig" gorm:"serializer:json"`
}

func (c CartItem) MarshalJSON() ([]byte, error) {
	type Alias CartItem
	return json.Marshal(&struct {
		Alias
		Price map[string]any `json:"price"`
	}{
		Alias: (Alias)(c),
		Price: map[string]any{
			"amount":   c.UnitAmount,
			"currency": c.Currency,
		},
	})
}

// Events
type CartUpdatedEvent struct {
	CartID     string     `json:"cartId"`
	CustomerID string     `json:"customerId"`
	Items      []CartItem `json:"items"`
}

type ProductPrice struct {
	ID         string    `json:"id" gorm:"primaryKey"` // This is the OfferingID
	UnitAmount float64   `json:"unitAmount"`
	Currency   string    `json:"currency"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
