package form

import "time"

// Invite represents the form for creating an invite link for a visitor user.
type Invite struct {
	Name      string     `json:"Name"`
	Perm      uint       `json:"Perm"`
	ExpiresAt *time.Time `json:"ExpiresAt"`
	Comment   string     `json:"Comment"`
}

// SharePerm represents the form for updating a UserShare permission.
type SharePerm struct {
	Perm      uint       `json:"Perm"`
	ExpiresAt *time.Time `json:"ExpiresAt"`
	Comment   string     `json:"Comment"`
}
