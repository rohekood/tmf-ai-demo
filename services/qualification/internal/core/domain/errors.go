package domain

import "fmt"

type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Predefined Domain Errors
var (
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeBackendFailure   = "BACKEND_FAILURE"
	ErrCodeResourceNotFound = "RESOURCE_NOT_FOUND"
)

func NewInvalidInputError(msg string, err error) error {
	return &DomainError{Code: ErrCodeInvalidInput, Message: msg, Err: err}
}

func NewBackendError(msg string, err error) error {
	return &DomainError{Code: ErrCodeBackendFailure, Message: msg, Err: err}
}
