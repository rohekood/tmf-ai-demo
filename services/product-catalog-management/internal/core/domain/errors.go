package domain

import "errors"

var (
	ErrNotFound          = errors.New("record not found")
	ErrAlreadyExists     = errors.New("record already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInternal          = errors.New("internal error")
	ErrSpecificationUsed = errors.New("specification is in use")
)
