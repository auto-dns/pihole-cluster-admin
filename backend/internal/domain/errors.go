package domain

type DuplicateHostPortError struct{}

func (e *DuplicateHostPortError) Error() string {
	return "duplicate host:port"
}

type DuplicateNameError struct{}

func (e *DuplicateNameError) Error() string {
	return "duplicate name"
}
