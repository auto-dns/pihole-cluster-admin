package domain

import "time"

type PiholeNode struct {
	Id          int64
	Scheme      string
	Host        string
	Port        int
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Used for log fan-out / light identity
type PiholeNodeRef struct {
	Id   int64
	Name string
	Host string
}

// Keep secrets separate so they don’t “ride along” accidentally
type PiholeNodeSecret struct {
	NodeId   int64
	Password string
}
