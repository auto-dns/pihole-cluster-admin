package sessions

type storage interface {
	Create(session Session) error
	GetAll() ([]Session, error)
	GetUserId(sessionId string) (int64, bool, error)
	Delete(sessionId string) error
}
