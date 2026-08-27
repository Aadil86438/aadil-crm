package models

import "time"

// Contact represents a CRM contact
type Contact struct {
	ID          string     `json:"id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	Mobile      string     `json:"mobile"`
	JobTitle    string     `json:"job_title"`
	Department  string     `json:"department"`
	AccountID   *string    `json:"account_id"`
	AccountName string     `json:"account_name"`
	OwnerID     *string    `json:"owner_id"`
	OwnerName   string     `json:"owner_name"`
	Address     string     `json:"address"`
	City        string     `json:"city"`
	State       string     `json:"state"`
	Country     string     `json:"country"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateContactRequest is the request body for creating/updating a contact
type CreateContactRequest struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Email       string  `json:"email"`
	Phone       string  `json:"phone"`
	Mobile      string  `json:"mobile"`
	JobTitle    string  `json:"job_title"`
	Department  string  `json:"department"`
	AccountID   *string `json:"account_id"`
	OwnerID     *string `json:"owner_id"`
	Address     string  `json:"address"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	Country     string  `json:"country"`
	Description string  `json:"description"`
}

type UpdateContactRequest = CreateContactRequest

// ContactListParams holds query params for listing contacts
type ContactListParams struct {
	Search     string
	AccountID  string
	OwnerID    string
	Page       int
	PageSize   int
	SortColumn string
	SortDir    string
}
