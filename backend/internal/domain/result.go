package domain

type NodeResult[T any] struct {
	PiholeNode  PiholeNodeRef
	Success     bool
	Error       error
	ErrorString string
	Response    *T
}
