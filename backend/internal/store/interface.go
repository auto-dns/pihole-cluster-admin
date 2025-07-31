package store

type PiholeStoreInterface interface {
	AddPiholeNode(params AddPiholeParams) (*PiholeNode, error)
	UpdatePiholeNode(id int64, params UpdatePiholeParams) (*PiholeNode, error)
	GetAllPiholeNodes() ([]PiholeNode, error)
	GetPiholeNode(id int64) (*PiholeNode, error)
	UpdatePiholePassword(id int64, newPassword string) error
	RemovePiholeNode(id int64) error
}

type UserStoreInterface interface {
	CreateUser(username, password string) error
	ValidateUser(username, password string) (bool, error)
	IsInitialized() (bool, error)
}
