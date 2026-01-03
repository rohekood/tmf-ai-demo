package domain

import (
	"time"
)

type Category struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description,omitempty"`
	ParentID        *string    `json:"parentId,omitempty"` // For sub-categories
	IsRoot          bool       `json:"isRoot"`
	CatalogID       *string    `json:"catalogId,omitempty"` // If attached directly to a catalog
	ValidFor        TimePeriod `json:"validFor"`
	LastUpdate      time.Time  `json:"lastUpdate"`
	LifecycleStatus string     `json:"lifecycleStatus"`
}

func (c *Category) Validate() error {
	if c.Name == "" {
		return ErrInvalidInput
	}
	return nil
}
