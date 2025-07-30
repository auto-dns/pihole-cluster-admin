package store

type PiholeStoreInterface interface {
	GetAllNodes() ([]PiholeNode, error)
	AddPiholeNode(PiholeNode) error
	GetPiholeNode(int) (*PiholeNode, error)
	UpdatePiholePassword(int, string) error
}

type UserStoreInterface interface {
	CreateUser(username, password string) error
	ValidateUser(username, password string) (bool, error)
	IsInitialized() (bool, error)
}
