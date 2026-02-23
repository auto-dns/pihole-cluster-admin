package middleware

type sessionManager interface {
	GetUserId(sessionId string) (int64, bool, error)
}
