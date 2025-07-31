package store

type PiholeStoreInterface interface {
	AddPiholeNode(node PiholeNode) error
	GetAllPiholeNodes() ([]PiholeNode, error)
	GetPiholeNode(id int) (*PiholeNode, error)
	UpdatePiholePassword(id int, newPassword string) error
	RemovePiholeNode(id int) error
}

type UserStoreInterface interface {
	CreateUser(username, password string) error
	ValidateUser(username, password string) (bool, error)
	IsInitialized() (bool, error)
}
