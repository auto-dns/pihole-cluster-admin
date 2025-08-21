package httpx

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrNotFound        = errors.New("user not found")
	ErrValidation      = errors.New("validation failed")
	ErrConflict        = errors.New("conflict")
	ErrInternalService = errors.New("internal service error")
)

type HttpError struct {
	Kind    error
	Message string
}

func (e *HttpError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Kind, e.Message)
	}
	return e.Kind.Error()
}

func NewHttpError(kind error, msg string) *HttpError {
	return &HttpError{
		Kind:    kind,
		Message: msg,
	}
}
