package domain

type NodeResult[T any] struct {
	PiholeNode PiholeNodeRef
	Success    bool
	Error      error
	NodeErr    *NodeError
	Response   T
}

func (r *NodeResult[T]) ErrorMessage() string {
	if r == nil {
		return ""
	}
	if r.NodeErr != nil && r.NodeErr.Message != "" {
		return r.NodeErr.Message
	}
	if r.Error != nil {
		return r.Error.Error()
	}
	return ""
}
