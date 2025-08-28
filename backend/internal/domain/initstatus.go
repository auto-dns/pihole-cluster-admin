package domain

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

type InitStatus struct {
	UserCreated  bool         `json:"userCreated"`
	PiholeStatus PiholeStatus `json:"piholeStatus"`
}
