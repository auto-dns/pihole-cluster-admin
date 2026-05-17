package domain

import "time"

type AdlistType string

const (
	AdlistTypeBlock AdlistType = "block"
	AdlistTypeAllow AdlistType = "allow"
)

type Adlist struct {
	Id             int64
	Address        string
	Type           AdlistType
	Comment        *string
	Groups         []int
	Enabled        bool
	DateAdded      time.Time
	DateModified   time.Time
	DateUpdated    time.Time
	Number         int64
	InvalidDomains int64
	Status         int
}

type AdlistSet struct {
	Lists []Adlist
}

type AddAdlistCommand struct {
	Address string
	Type    AdlistType
	Comment *string
	Groups  []int
	Enabled bool
}

type UpdateAdlistCommand struct {
	Id      int64
	Enabled *bool
	Comment *string
}

type RemoveAdlistCommand struct {
	Id int64
}
