package pihole

type ClientInterface interface {
	FetchLogs(opts FetchLogsQueryOptions) (*QueryLogResponse, error)
	Logout() error
}

type ClusterInterface interface {
	FetchLogs(opts FetchLogsQueryOptions) ([]*QueryLogResponse, []error)
	Logout() []error
}
