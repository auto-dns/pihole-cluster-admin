package domain

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("user not found")
	ErrValidation   = errors.New("validation failed")
	ErrConflict     = errors.New("conflict")
)

// Pihole errors

type DuplicateHostPortError struct{}

func (e *DuplicateHostPortError) Error() string {
	return "duplicate host:port"
}

type DuplicateNameError struct{}

func (e *DuplicateNameError) Error() string {
	return "duplicate name"
}

//
