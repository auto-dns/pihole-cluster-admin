package store

type PiholeStoreInterface interface {
	AddPiholeNode(params AddPiholeParams) (*PiholeNode, error)
	UpdatePiholeNode(id int64, params UpdatePiholeParams) (*PiholeNode, error)
	GetAllPiholeNodes() ([]*PiholeNode, error)
	GetAllPiholeNodesWithPasswords() ([]*PiholeNode, error)
	RemovePiholeNode(id int64) (found bool, err error)
}

type UserStoreInterface interface {
	CreateUser(params CreateUserParams) (*User, error)
	ValidateUser(username, password string) (bool, error)
	IsInitialized() (bool, error)
}
