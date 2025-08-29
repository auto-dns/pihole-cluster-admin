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

// New errors - migrate to these

type NodeErrorCode string

const (
	NodeErrAuth      NodeErrorCode = "auth"
	NodeErrTimeout   NodeErrorCode = "timeout"
	NodeErrTransport NodeErrorCode = "transport"
	NodeErrProtocol  NodeErrorCode = "protocol" // non-2xx from Pi-hole
	NodeErrDecode    NodeErrorCode = "decode"
	NodeErrUnknown   NodeErrorCode = "unknown"
)

type NodeError struct {
	Code       NodeErrorCode
	Message    string
	HTTPStatus int
	Temporary  bool
}

func (e *NodeError) Error() string { return e.Message }
