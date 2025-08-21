package httpx

import (
	"encoding/json"
	"net/http"
)

func WriteJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func WriteJSONErrorFromErr(w http.ResponseWriter, err error) {
	message := "internal service error"
	status := http.StatusInternalServerError

	if e, ok := err.(*HttpError); ok {
		message = e.Kind.Error()
		switch e.Kind {
		case ErrUnauthorized:
			status = http.StatusUnauthorized
		case ErrForbidden:
			status = http.StatusForbidden
		case ErrNotFound:
			status = http.StatusNotFound
		case ErrValidation:
			status = http.StatusBadRequest
		case ErrConflict:
			status = http.StatusConflict
		case ErrInternalService:
		default:
			status = http.StatusInternalServerError
		}
	}

	WriteJSONError(w, message, status)
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}
