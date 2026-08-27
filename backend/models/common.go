package models

import "time"

// Note represents a CRM note
type Note struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	Body               string    `json:"body"`
	OwnerID            *string   `json:"owner_id"`
	OwnerName          string    `json:"owner_name"`
	RelatedLeadID      *string   `json:"related_lead_id"`
	RelatedContactID   *string   `json:"related_contact_id"`
	RelatedAccountID   *string   `json:"related_account_id"`
	RelatedDealID      *string   `json:"related_deal_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// CreateNoteRequest is the request body for creating/updating a note
type CreateNoteRequest struct {
	Title            string  `json:"title"`
	Body             string  `json:"body"`
	RelatedLeadID    *string `json:"related_lead_id"`
	RelatedContactID *string `json:"related_contact_id"`
	RelatedAccountID *string `json:"related_account_id"`
	RelatedDealID    *string `json:"related_deal_id"`
}

type UpdateNoteRequest = CreateNoteRequest

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        string      `json:"id"`
	UserID    *string     `json:"user_id"`
	UserEmail string      `json:"user_email"`
	Action    string      `json:"action"`
	Entity    string      `json:"entity"`
	EntityID  *string     `json:"entity_id"`
	Metadata  interface{} `json:"metadata"`
	IPAddress string      `json:"ip_address"`
	CreatedAt time.Time   `json:"created_at"`
}

// CreateAuditLogRequest is the request for creating an audit log entry
type CreateAuditLogRequest struct {
	UserID    *string     `json:"user_id"`
	UserEmail string      `json:"user_email"`
	Action    string      `json:"action"`
	Entity    string      `json:"entity"`
	EntityID  *string     `json:"entity_id"`
	Metadata  interface{} `json:"metadata"`
	IPAddress string      `json:"ip_address"`
}

// DashboardStats holds the main dashboard statistics
type DashboardStats struct {
	TotalLeads         int     `json:"total_leads"`
	NewLeads           int     `json:"new_leads"`
	QualifiedLeads     int     `json:"qualified_leads"`
	TotalContacts      int     `json:"total_contacts"`
	TotalAccounts      int     `json:"total_accounts"`
	OpenDeals          int     `json:"open_deals"`
	WonDeals           int     `json:"won_deals"`
	LostDeals          int     `json:"lost_deals"`
	TotalRevenue       float64 `json:"total_revenue"`
	PipelineValue      float64 `json:"pipeline_value"`
	ActivitiesToday    int     `json:"activities_today"`
	UpcomingTasks      int     `json:"upcoming_tasks"`
}

// SearchResult holds categorized search results
type SearchResult struct {
	Leads    []SearchItem `json:"leads"`
	Contacts []SearchItem `json:"contacts"`
	Accounts []SearchItem `json:"accounts"`
	Deals    []SearchItem `json:"deals"`
	Tasks    []SearchItem `json:"tasks"`
}

// SearchItem is a generic search result item
type SearchItem struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}
