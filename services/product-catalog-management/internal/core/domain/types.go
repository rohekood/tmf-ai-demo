package domain

import "time"

type contextKey string

const (
	UserContextKey contextKey = "user"
	AuthContextKey contextKey = "Authorization"
)

type TimePeriod struct {
	StartDateTime *time.Time `json:"startDateTime,omitempty"`
	EndDateTime   *time.Time `json:"endDateTime,omitempty"`
}

type Money struct {
	Unit  string  `json:"unit,omitempty"`
	Value float64 `json:"value,omitempty"`
}
