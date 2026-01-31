package domain

import (
	"time"
)

// Pure Domain (No GORM tags)
type OutboxEvent struct {
	ID        string
	Topic     string
	Payload   []byte
	Status    string // PENDING, PUBLISHED
	CreatedAt time.Time
}
