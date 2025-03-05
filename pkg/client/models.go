package client

import "time"

type User struct {
	ID                  string `json:"user_id"`
	Email               string `json:"email"`
	FirstName           string `json:"first_name,omitempty"`
	Lastame             string `json:"last_name,omitempty"`
	Phone               string `json:"phone_number,omitempty"`
	IsOrganizationOwner bool   `json:"is_organization_owner"`
	License             string `json:"license"`
	Workspaces          []struct {
		Role         string `json:"role"`
		WorkspaceID  string `json:"workspace_id"`
		MembershipID string `json:"membership_id"`
	}
}

type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Owner       string    `json:"owner"`
	DateCreated time.Time `json:"date_created"`
}

type Role struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	IsSystem    bool   `json:"is_system,omitempty"`
}
