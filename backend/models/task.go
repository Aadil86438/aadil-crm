package models

import "time"

// Task represents a CRM task
type Task struct {
	ID               string     `json:"id"`
	Subject          string     `json:"subject"`
	Description      string     `json:"description"`
	DueDate          *time.Time `json:"due_date"`
	Priority         string     `json:"priority"`
	Status           string     `json:"status"`
	OwnerID          *string    `json:"owner_id"`
	OwnerName        string     `json:"owner_name"`
	RelatedLeadID    *string    `json:"related_lead_id"`
	RelatedLeadName  string     `json:"related_lead_name"`
	RelatedContactID *string    `json:"related_contact_id"`
	RelatedContactName string   `json:"related_contact_name"`
	RelatedAccountID *string    `json:"related_account_id"`
	RelatedAccountName string   `json:"related_account_name"`
	RelatedDealID    *string    `json:"related_deal_id"`
	RelatedDealName  string     `json:"related_deal_name"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CreateTaskRequest is the request body for creating/updating a task
type CreateTaskRequest struct {
	Subject          string     `json:"subject"`
	Description      string     `json:"description"`
	DueDate          *time.Time `json:"due_date"`
	Priority         string     `json:"priority"`
	Status           string     `json:"status"`
	OwnerID          *string    `json:"owner_id"`
	RelatedLeadID    *string    `json:"related_lead_id"`
	RelatedContactID *string    `json:"related_contact_id"`
	RelatedAccountID *string    `json:"related_account_id"`
	RelatedDealID    *string    `json:"related_deal_id"`
}

type UpdateTaskRequest = CreateTaskRequest

// TaskListParams holds query params for listing tasks
type TaskListParams struct {
	Search     string
	Status     string
	Priority   string
	OwnerID    string
	DueBefore  *time.Time
	DueAfter   *time.Time
	Page       int
	PageSize   int
	SortColumn string
	SortDir    string
}
