package v1

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// ParseInt64Param parses a URL path parameter as int64.
// Returns (value, true) if the parameter is a valid int64 >= minVal, otherwise (0, false).
func ParseInt64Param(r *http.Request, name string, minVal int64) (int64, bool) {
	s := chi.URLParam(r, name)
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < minVal {
		return 0, false
	}
	return n, true
}
