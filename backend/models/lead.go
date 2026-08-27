package models

import (
	"time"
)

// Lead represents a CRM lead
type Lead struct {
	ID          string     `json:"id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Company     string     `json:"company"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	Mobile      string     `json:"mobile"`
	Website     string     `json:"website"`
	LeadSource  string     `json:"lead_source"`
	LeadStatus  string     `json:"lead_status"`
	Industry    string     `json:"industry"`
	JobTitle    string     `json:"job_title"`
	AnnualRevenue *float64 `json:"annual_revenue"`
	NumEmployees  *int     `json:"num_employees"`
	Rating      string     `json:"rating"`
	Address     string     `json:"address"`
	City        string     `json:"city"`
	State       string     `json:"state"`
	Country     string     `json:"country"`
	Description string     `json:"description"`
	OwnerID     *string    `json:"owner_id"`
	OwnerName   string     `json:"owner_name"`
	IsConverted bool       `json:"is_converted"`
	ConvertedAt *time.Time `json:"converted_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateLeadRequest is the request body for creating a lead
type CreateLeadRequest struct {
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	Company       string   `json:"company"`
	Email         string   `json:"email"`
	Phone         string   `json:"phone"`
	Mobile        string   `json:"mobile"`
	Website       string   `json:"website"`
	LeadSource    string   `json:"lead_source"`
	LeadStatus    string   `json:"lead_status"`
	Industry      string   `json:"industry"`
	JobTitle      string   `json:"job_title"`
	AnnualRevenue *float64 `json:"annual_revenue"`
	NumEmployees  *int     `json:"num_employees"`
	Rating        string   `json:"rating"`
	Address       string   `json:"address"`
	City          string   `json:"city"`
	State         string   `json:"state"`
	Country       string   `json:"country"`
	Description   string   `json:"description"`
	OwnerID       *string  `json:"owner_id"`
}

// UpdateLeadRequest is the request body for updating a lead
type UpdateLeadRequest = CreateLeadRequest

// ConvertLeadRequest is the request body for converting a lead
type ConvertLeadRequest struct {
	CreateContact bool    `json:"create_contact"`
	CreateAccount bool    `json:"create_account"`
	CreateDeal    bool    `json:"create_deal"`
	AccountID     *string `json:"account_id"`
	DealName      string  `json:"deal_name"`
	DealAmount    *float64 `json:"deal_amount"`
}

// LeadConversionResult is the response for a lead conversion
type LeadConversionResult struct {
	LeadID    string  `json:"lead_id"`
	ContactID *string `json:"contact_id"`
	AccountID *string `json:"account_id"`
	DealID    *string `json:"deal_id"`
}

// LeadListParams holds query params for listing leads
type LeadListParams struct {
	Search     string
	Status     string
	Source     string
	OwnerID    string
	Page       int
	PageSize   int
	SortColumn string
	SortDir    string
}
