package domain

import "errors"

// Domain-specific errors for Customer management
var (
	// ErrNotFound indicates that a requested customer was not found
	ErrNotFound = errors.New("customer not found")

	// ErrInvalidType indicates an invalid type was provided
	ErrInvalidType = errors.New("invalid type")

	// ErrInvalidPayload indicates the request payload is malformed or missing required fields
	ErrInvalidPayload = errors.New("invalid payload")

	// ErrIDRequired indicates the ID is required but was not provided
	ErrIDRequired = errors.New("id is required")
)
