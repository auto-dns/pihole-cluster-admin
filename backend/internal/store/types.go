package store

import "time"

// Initialization status store

type PiholeStatus string

const (
	PiholeUninitialized PiholeStatus = "UNINITIALIZED"
	PiholeAdded         PiholeStatus = "ADDED"
	PiholeSkipped       PiholeStatus = "SKIPPED"
)

type InitializationStatus struct {
	UserCreated  bool         `json:"userCreated"`
	PiholeStatus PiholeStatus `json:"piholeStatus"`
}

// Pihole store

type PiholeNode struct {
	Id          int64     `json:"id"`
	Scheme      string    `json:"scheme"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Password    *string   `json:"password,omitempty"` // plaintext on input/output
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type AddPiholeParams struct {
	Scheme      string `json:"scheme"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Password    string `json:"password"` // plaintext on input/output
}

type UpdatePiholeParams struct {
	Scheme      *string `json:"scheme"`
	Host        *string `json:"host"`
	Port        *int    `json:"port"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Password    *string `json:"password"`
}

// User store

type User struct {
	Id           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash *string   `json:"password_hash,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
