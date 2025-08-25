package domain

type NodeResult[T any] struct {
	PiholeNode PiholeNodeRef `json:"piholeNode"`
	Success    bool          `json:"success"`
	Error      error         `json:"-"`
	NodeErr    *NodeError    `json:"error,omitempty"`
	Response   T             `json:"response,omitempty"`
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
