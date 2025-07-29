package pihole

type ClientInterface interface {
	FetchLogs(from, length int64) (*QueryLogResponse, error)
	Logout() error
}
