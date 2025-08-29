package domain

import "time"

type AuthStatus struct {
	Valid      bool
	Validity   time.Duration
	ValidUntil time.Time
	Took       time.Duration
}
