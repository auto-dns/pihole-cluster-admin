package domain

import "time"

type PiholeNode struct {
	Id          int64     `json:"id"`
	Scheme      string    `json:"scheme"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Used for log fan-out / light identity
type PiholeNodeRef struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
}

// Keep secrets separate so they don’t “ride along” accidentally
type PiholeNodeSecret struct {
	NodeId   int64
	Password string
}
