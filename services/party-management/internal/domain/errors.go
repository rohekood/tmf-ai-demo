package domain

import "errors"

// Domain-specific errors for Party management
var (
	// ErrNotFound indicates that a requested party was not found
	ErrNotFound = errors.New("party not found")

	// ErrInvalidType indicates an invalid party type was provided
	ErrInvalidType = errors.New("invalid party type")

	// ErrInvalidPayload indicates the request payload is malformed or missing required fields
	ErrInvalidPayload = errors.New("invalid payload")

	// ErrIDRequired indicates the party ID is required but was not provided
	ErrIDRequired = errors.New("party id is required")
)
