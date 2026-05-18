package domain

import "time"

type PiholeClient struct {
	Id           int64
	IP           string
	Name         string
	Comment      *string
	Groups       []int
	DateAdded    time.Time
	DateModified time.Time
}

type PiholeClientSet struct {
	Clients []PiholeClient
}

type UpdatePiholeClientCommand struct {
	Id      int64
	Groups  []int
	Comment *string
}

type RemovePiholeClientCommand struct {
	Id int64
}
