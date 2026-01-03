package domain

import "time"

type TimePeriod struct {
	StartDateTime *time.Time `json:"startDateTime"`
	EndDateTime   *time.Time `json:"endDateTime,omitempty"`
}

type Money struct {
	Unit  string  `json:"unit"` // e.g., "USD", "EUR"
	Value float64 `json:"value"`
}
