package transport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
)

// Encoding response bodies

// WriteJSON sets Content-Type, status code, and encodes v as JSON. No-op if encoding fails.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Decoding request bodies

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return errs.Invalid("invalid JSON", err)
	}
	// Ensure no trailing non-whitespace / extra JSON values
	if err := dec.Decode(new(struct{})); !errors.Is(err, io.EOF) {
		// err==nil => a second JSON value; other errs => trailing garbage
		return errs.Invalid("request body must contain a single JSON value", err)
	}
	return nil
}

// -- Errors

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func statusForKind(k errs.Kind) int {
	switch k {
	case errs.KindInvalid:
		return http.StatusBadRequest
	case errs.KindUnauthorized:
		return http.StatusUnauthorized
	case errs.KindForbidden:
		return http.StatusForbidden
	case errs.KindNotFound:
		return http.StatusNotFound
	case errs.KindConflict:
		return http.StatusConflict
	case errs.KindUnavailable:
		return http.StatusServiceUnavailable
	case errs.KindInternal, errs.KindUnknown:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

func writeJSONError(w http.ResponseWriter, message string, status int, code string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody{
		Code:    code,
		Message: message,
	})
}

func writeJSONErrorFromErr(w http.ResponseWriter, err error) {
	k := errs.KindOf(err)
	msg := errs.SafeMessage(err)
	status := statusForKind(k)
	writeJSONError(w, msg, status, string(k))
}

func WriteErr(w http.ResponseWriter, err error) {
	writeJSONErrorFromErr(w, err)
}

func WriteBadRequestErr(w http.ResponseWriter, msg string, cause error) {
	WriteErr(w, errs.Invalid(msg, cause))
}

func WriteUnauthorizedErr(w http.ResponseWriter, msg string) {
	WriteErr(w, errs.Unauthorized(msg, nil))
}

func WriteNotFoundErr(w http.ResponseWriter, msg string) {
	WriteErr(w, errs.NotFound(msg, nil))
}
