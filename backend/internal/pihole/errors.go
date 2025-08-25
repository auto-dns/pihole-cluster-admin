package pihole

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

type httpStatusError struct {
	Status int
	Body   string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("http %d: %s", e.Status, e.Body)
}

func mapClientErr(err error) *domain.NodeError {
	if err == nil {
		return nil
	}

	// Unwrap common wrappers
	var uerr *url.Error
	if errors.As(err, &uerr) {
		err = uerr.Err
	}

	// Context / timeouts
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return &domain.NodeError{Code: domain.NodeErrTimeout, Message: "request timed out", Temporary: true}
	}

	// net.Error
	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() {
			return &domain.NodeError{Code: domain.NodeErrTimeout, Message: "request timed out", Temporary: true}
		}
		if ne.Temporary() {
			return &domain.NodeError{Code: domain.NodeErrTransport, Message: "temporary network error", Temporary: true}
		}
	}

	// Our HTTP status wrapper
	var he *httpStatusError
	if errors.As(err, &he) {
		code := domain.NodeErrProtocol
		msg := fmt.Sprintf("Pi-hole returned HTTP %d", he.Status)
		if he.Status == 401 || he.Status == 403 {
			code = domain.NodeErrAuth
			msg = "authentication failed"
		}
		return &domain.NodeError{
			Code:       code,
			Message:    msg,
			HTTPStatus: he.Status,
			Temporary:  he.Status >= 500,
		}
	}

	// JSON decode
	var se *json.SyntaxError
	if errors.As(err, &se) {
		return &domain.NodeError{Code: domain.NodeErrDecode, Message: "invalid JSON from Pi-hole"}
	}
	var te *json.UnmarshalTypeError
	if errors.As(err, &te) {
		return &domain.NodeError{Code: domain.NodeErrDecode, Message: "unexpected JSON shape from Pi-hole"}
	}

	// Fallback
	return &domain.NodeError{Code: domain.NodeErrUnknown, Message: err.Error()}
}
