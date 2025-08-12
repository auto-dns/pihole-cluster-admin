package store

type InitializationStatusStoreInterface interface {
	GetInitializationStatus() (*InitializationStatus, error)
	SetUserCreated(userCreated bool) error
	SetPiholeStatus(piholeStatus PiholeStatus) error
}

type PiholeStoreInterface interface {
	AddPiholeNode(params AddPiholeParams) (*PiholeNode, error)
	UpdatePiholeNode(id int64, params UpdatePiholeParams) (*PiholeNode, error)
	GetPiholeNode(id int64) (*PiholeNode, error)
	GetAllPiholeNodes() ([]*PiholeNode, error)
	GetAllPiholeNodesWithPasswords() ([]*PiholeNode, error)
	RemovePiholeNode(id int64) (found bool, err error)
}

type UserStoreInterface interface {
	CreateUser(params CreateUserParams) (*User, error)
	GetUser(id int64) (*User, error)
	ValidateUser(username, password string) (*User, error)
	IsInitialized() (bool, error)
}
