package domain

import "time"

type Group struct {
	Id           int
	Name         string
	Description  *string
	Enabled      bool
	DateAdded    time.Time
	DateModified time.Time
}

type GroupSet struct {
	Groups []Group
}

type AddGroupCommand struct {
	Name        string
	Description *string
	Enabled     bool
}

type UpdateGroupCommand struct {
	Name        string
	Description *string
	Enabled     *bool
}

type RemoveGroupCommand struct {
	Name string
}
