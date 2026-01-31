package domain

import (
	"time"
)

type Cart struct {
	ID                 string     `json:"id" gorm:"primaryKey"`
	CustomerID         string     `json:"customerId"`
	Items              []CartItem `json:"items" gorm:"foreignKey:CartID"`
	Status             string     `json:"status"` // Active, Closed
	TotalPrice         float64    `json:"totalPrice"`
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
	UnitPrice     float64           `json:"unitPrice"`
	UnitAmount    float64           `json:"unitAmount"`
	Currency      string            `json:"currency"`
	ProductConfig map[string]string `json:"productConfig" gorm:"serializer:json"`
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
