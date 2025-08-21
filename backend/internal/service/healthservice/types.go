package healthservice

import "time"

type Status string

const (
	StatusOnline   Status = "online"
	StatusDegraded Status = "degraded"
	StatusOffline  Status = "offline"
)

type NodeHealth struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	LatencyMS int       `json:"latencyMs"`
	LastErr   string    `json:"lastErr,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Summary struct {
	Online    int       `json:"online"`
	Total     int       `json:"total"`
	UpdatedAt time.Time `json:"updatedAt"`
}
