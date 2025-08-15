package domain

import (
	"database/sql/driver"
	"fmt"
)

type PiholeStatus string

const (
	PiholeUninitialized PiholeStatus = "UNINITIALIZED"
	PiholeAdded         PiholeStatus = "ADDED"
	PiholeSkipped       PiholeStatus = "SKIPPED"
)

func (s PiholeStatus) IsValid() bool {
	switch s {
	case PiholeUninitialized, PiholeAdded, PiholeSkipped:
		return true
	default:
		return false
	}
}

// sql.Scanner implementation for DB to read value into PiholeStatus.
func (s *PiholeStatus) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*s = PiholeStatus(v)
	case []byte:
		*s = PiholeStatus(string(v))
	default:
		return fmt.Errorf("pihole status: unsupported scan type %T", src)
	}
	if !s.IsValid() {
		return fmt.Errorf("pihole status: invalid value %q", string(*s))
	}
	return nil
}

// driver.Valuer implementation for writing directly to DB
func (s PiholeStatus) Value() (driver.Value, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("pihole status: invalid value %q", string(s))
	}
	return string(s), nil
}

type InitStatus struct {
	UserCreated  bool
	PiholeStatus PiholeStatus
}
