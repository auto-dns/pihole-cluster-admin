package errs

import "errors"

type Kind string

const (
	KindUnknown      Kind = "unknown"
	KindInvalid      Kind = "invalid" // validation failed / bad input
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindUnavailable  Kind = "unavailable" // upstream/downstream not reachable
	KindInternal     Kind = "internal"
)

type E struct {
	Kind Kind   // machine class of error
	Msg  string // safe user-facing message
	Err  error  // underlying cause (not serialized)
}

func (e *E) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Kind)
}
func (e *E) Unwrap() error { return e.Err }

// Constructors
func New(kind Kind, msg string, cause error) error { return &E{Kind: kind, Msg: msg, Err: cause} }
func Unknown(msg string, cause error) error        { return &E{Kind: KindUnknown, Msg: msg, Err: cause} }
func Invalid(msg string, cause error) error        { return &E{Kind: KindInvalid, Msg: msg, Err: cause} }
func Unauthorized(msg string, cause error) error {
	return &E{Kind: KindUnauthorized, Msg: msg, Err: cause}
}
func Forbidden(msg string, cause error) error { return &E{Kind: KindForbidden, Msg: msg, Err: cause} }
func NotFound(msg string, cause error) error  { return &E{Kind: KindNotFound, Msg: msg, Err: cause} }
func Conflict(msg string, cause error) error  { return &E{Kind: KindConflict, Msg: msg, Err: cause} }
func Unavailable(msg string, cause error) error {
	return &E{Kind: KindUnavailable, Msg: msg, Err: cause}
}
func Internal(msg string, cause error) error { return &E{Kind: KindInternal, Msg: msg, Err: cause} }

// Helpers
func KindOf(err error) Kind {
	var e *E
	if errors.As(err, &e) {
		return e.Kind
	}
	return KindUnknown
}

// SafeMessage returns a safe, user-facing message if present, else a generic one by kind.
func SafeMessage(err error) string {
	var e *E
	if errors.As(err, &e) && e.Msg != "" {
		return e.Msg
	}
	switch KindOf(err) {
	case KindInvalid:
		return "validation failed"
	case KindUnauthorized:
		return "unauthorized"
	case KindForbidden:
		return "forbidden"
	case KindNotFound:
		return "not found"
	case KindConflict:
		return "conflict"
	case KindUnavailable:
		return "service unavailable"
	case KindInternal, KindUnknown:
		fallthrough
	default:
		return "internal error"
	}
}
