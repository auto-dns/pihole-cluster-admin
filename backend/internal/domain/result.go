package domain

type NodeResult[T any] struct {
	PiholeNode  PiholeNodeRef `json:"piholeNode"`
	Success     bool          `json:"success"`
	Error       error         `json:"-"`
	ErrorString string        `json:"error,omitempty"`
	Response    *T            `json:"response,omitempty"`
}
