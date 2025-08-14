package store

type InitializationStatusStoreInterface interface {
	GetInitializationStatus() (*InitializationStatus, error)
	SetUserCreated(userCreated bool) error
	SetPiholeStatus(piholeStatus PiholeStatus) error
}

type PiholeStoreInterface interface {
	AddPiholeNode(params AddPiholeParams) (*PiholeNode, error)
	UpdatePiholeNode(id int64, params UpdatePiholeParams) (*PiholeNode, error)
	GetPiholeNodeWithPassword(id int64) (*PiholeNode, error)
	GetAllPiholeNodes() ([]*PiholeNode, error)
	GetAllPiholeNodesWithPasswords() ([]*PiholeNode, error)
	RemovePiholeNode(id int64) (found bool, err error)
}

type SessionStoreInterface interface {
	CreateSession(params CreateSessionParams) (*Session, error)
	GetAllSessions() ([]*Session, error)
	GetSession(id string) (*Session, error)
	DeleteSession(id string) (found bool, err error)
}

type UserStoreInterface interface {
	CreateUser(params CreateUserParams) (*User, error)
	GetUser(id int64) (*User, error)
	ValidateUser(username, password string) (*User, error)
	IsInitialized() (bool, error)
}
