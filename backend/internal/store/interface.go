package store

type PiholeStoreInterface interface {
	AddPiholeNode(PiholeNode) error
	GetPiholeNode(int) (*PiholeNode, error)
	UpdatePiholePassword(int, string) error
}

type UserStoreInterface interface {
	CreateUser(username, password string) error
	ValidateUser(username, password string) (bool, error)
	IsInitialized() (bool, error)
}
