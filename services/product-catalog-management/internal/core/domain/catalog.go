package domain

import (
	"time"
)

type Catalog struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description,omitempty"`
	ValidFor        TimePeriod `json:"validFor"`
	LastUpdate      time.Time  `json:"lastUpdate"`
	LifecycleStatus string     `json:"lifecycleStatus"`
}

func (c *Catalog) Validate() error {
	if c.Name == "" {
		return ErrInvalidInput
	}
	return nil
}
