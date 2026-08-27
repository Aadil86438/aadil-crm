package models

import "time"

// Activity represents a CRM activity (call, meeting, task, note, email)
type Activity struct {
	ID                 string     `json:"id"`
	Type               string     `json:"type"`
	Subject            string     `json:"subject"`
	Description        string     `json:"description"`
	Status             string     `json:"status"`
	DueDate            *time.Time `json:"due_date"`
	CompletedAt        *time.Time `json:"completed_at"`
	OwnerID            *string    `json:"owner_id"`
	OwnerName          string     `json:"owner_name"`
	RelatedLeadID      *string    `json:"related_lead_id"`
	RelatedLeadName    string     `json:"related_lead_name"`
	RelatedContactID   *string    `json:"related_contact_id"`
	RelatedContactName string     `json:"related_contact_name"`
	RelatedAccountID   *string    `json:"related_account_id"`
	RelatedAccountName string     `json:"related_account_name"`
	RelatedDealID      *string    `json:"related_deal_id"`
	RelatedDealName    string     `json:"related_deal_name"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateActivityRequest is the request body for creating an activity
type CreateActivityRequest struct {
	Type             string     `json:"type"`
	Subject          string     `json:"subject"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	DueDate          *time.Time `json:"due_date"`
	OwnerID          *string    `json:"owner_id"`
	RelatedLeadID    *string    `json:"related_lead_id"`
	RelatedContactID *string    `json:"related_contact_id"`
	RelatedAccountID *string    `json:"related_account_id"`
	RelatedDealID    *string    `json:"related_deal_id"`
}

// ActivityListParams holds query params for listing activities
type ActivityListParams struct {
	Type             string
	RelatedLeadID    string
	RelatedContactID string
	RelatedAccountID string
	RelatedDealID    string
	OwnerID          string
	Page             int
	PageSize         int
}
