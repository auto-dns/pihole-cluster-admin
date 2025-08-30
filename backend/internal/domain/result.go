package domain

type NodeResult[T any] struct {
	PiholeNode PiholeNodeRef
	Success    bool
	Error      error
	Response   T
}
