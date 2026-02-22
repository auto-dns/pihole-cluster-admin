package pihole

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
)

type httpStatusError struct {
	Status int
	Body   string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("http %d: %s", e.Status, e.Body)
}

func mapClientErr(err error) error {
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
		return errs.Internal("request timed out", err)
	}

	// net.Error
	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() {
			return errs.Unavailable("request timed out", err)
		}
	}

	// Our HTTP status wrapper
	var he *httpStatusError
	if errors.As(err, &he) {
		if he.Status == 401 || he.Status == 403 {
			return errs.Internal("Pi-hole auth failed", err)
		}
		return errs.Unknown(fmt.Sprintf("Pi-hole returned status %d", he.Status), fmt.Errorf("Pi-hole returned status %d: %w", he.Status, err))
	}

	// JSON decode
	var se *json.SyntaxError
	if errors.As(err, &se) {
		return errs.Internal("invalid JSON from Pi-hole", err)
	}
	var te *json.UnmarshalTypeError
	if errors.As(err, &te) {
		return errs.Internal("unexpected JSON shape from Pi-hole", err)
	}

	// Fallback: include underlying error for debugging
	return errs.Unknown(fmt.Sprintf("Pi-hole request failed: %v", err), err)
}
