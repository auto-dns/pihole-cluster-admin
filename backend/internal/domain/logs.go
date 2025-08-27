package domain

import "time"

type QueryLogEntry struct {
	ID         int64
	Time       time.Time
	QType      string
	Status     string
	DNSSEC     string
	Domain     string
	Upstream   *string
	ReplyType  string
	ReplyTime  time.Duration
	ClientIP   string
	ClientName *string
	ListID     *int64
	EDECode    int64
	EDEText    *string
	CNAME      *string
}

type QueryLogPage struct {
	Entries         []QueryLogEntry
	Cursor          int
	RecordsTotal    int64
	RecordsFiltered int64
	Draw            int64
	Took            time.Duration
}

type QueryLogFilters struct {
	From, Until                            *time.Time
	Domain, ClientIP, ClientName, Upstream *string
	Type, Status, Reply, DNSSEC            *string
	Disk                                   *bool
}

type QueryLogRequest struct {
	Cursor  *string
	Length  *int
	Start   *int
	Filters QueryLogFilters
}

type ClusterQueryLogResponse struct {
	Cursor       string
	Results      map[int64]*NodeResult[*QueryLogPage]
	EndOfResults bool
}
