package models

import "time"

// Deal represents a sales opportunity/deal in the CRM
type Deal struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Amount            *float64   `json:"amount"`
	AccountID         *string    `json:"account_id"`
	AccountName       string     `json:"account_name"`
	ContactID         *string    `json:"contact_id"`
	ContactName       string     `json:"contact_name"`
	Stage             string     `json:"stage"`
	Probability       int        `json:"probability"`
	ExpectedCloseDate *time.Time `json:"expected_close_date"`
	LeadSource        string     `json:"lead_source"`
	OwnerID           *string    `json:"owner_id"`
	OwnerName         string     `json:"owner_name"`
	Description       string     `json:"description"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateDealRequest is the request body for creating/updating a deal
type CreateDealRequest struct {
	Name              string     `json:"name"`
	Amount            *float64   `json:"amount"`
	AccountID         *string    `json:"account_id"`
	ContactID         *string    `json:"contact_id"`
	Stage             string     `json:"stage"`
	Probability       int        `json:"probability"`
	ExpectedCloseDate *time.Time `json:"expected_close_date"`
	LeadSource        string     `json:"lead_source"`
	OwnerID           *string    `json:"owner_id"`
	Description       string     `json:"description"`
}

type UpdateDealRequest = CreateDealRequest

// UpdateDealStageRequest is used when moving a deal between Kanban stages
type UpdateDealStageRequest struct {
	Stage string `json:"stage"`
}

// DealListParams holds query params for listing deals
type DealListParams struct {
	Search     string
	Stage      string
	AccountID  string
	OwnerID    string
	Page       int
	PageSize   int
	SortColumn string
	SortDir    string
}

// PipelineSummary holds aggregated pipeline statistics
type PipelineSummary struct {
	TotalDeals     int     `json:"total_deals"`
	TotalValue     float64 `json:"total_value"`
	WeightedValue  float64 `json:"weighted_value"`
	WonDeals       int     `json:"won_deals"`
	WonRevenue     float64 `json:"won_revenue"`
	LostDeals      int     `json:"lost_deals"`
	LostRevenue    float64 `json:"lost_revenue"`
}
