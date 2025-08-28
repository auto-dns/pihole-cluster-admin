package domain

import "time"

type AuthStatus struct {
	Valid           bool          `json:"valid"`
	ValiditySeconds int           `json:"validitySeconds"`
	ValidUntil      time.Time     `json:"validUntil"`
	Took            time.Duration `json:"took"`
}
