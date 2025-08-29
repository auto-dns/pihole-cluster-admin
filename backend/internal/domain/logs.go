package domain

import "time"

type QueryLogEntry struct {
	Id         int64
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
	From       *time.Time
	Until      *time.Time
	Domain     *string
	ClientIP   *string
	ClientName *string
	Upstream   *string
	Type       *string
	Status     *string
	Reply      *string
	DNSSEC     *string
	Disk       *bool
}

type QueryLogQuery struct {
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
