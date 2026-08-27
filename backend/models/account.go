package models

import "time"

// Account represents a company/account in the CRM
type Account struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Website       string     `json:"website"`
	Industry      string     `json:"industry"`
	Phone         string     `json:"phone"`
	Email         string     `json:"email"`
	NumEmployees  *int       `json:"num_employees"`
	AnnualRevenue *float64   `json:"annual_revenue"`
	AccountType   string     `json:"account_type"`
	OwnerID       *string    `json:"owner_id"`
	OwnerName     string     `json:"owner_name"`
	Address       string     `json:"address"`
	City          string     `json:"city"`
	State         string     `json:"state"`
	Country       string     `json:"country"`
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreateAccountRequest is the request body for creating/updating an account
type CreateAccountRequest struct {
	Name          string   `json:"name"`
	Website       string   `json:"website"`
	Industry      string   `json:"industry"`
	Phone         string   `json:"phone"`
	Email         string   `json:"email"`
	NumEmployees  *int     `json:"num_employees"`
	AnnualRevenue *float64 `json:"annual_revenue"`
	AccountType   string   `json:"account_type"`
	OwnerID       *string  `json:"owner_id"`
	Address       string   `json:"address"`
	City          string   `json:"city"`
	State         string   `json:"state"`
	Country       string   `json:"country"`
	Description   string   `json:"description"`
}

type UpdateAccountRequest = CreateAccountRequest

// AccountListParams holds query params for listing accounts
type AccountListParams struct {
	Search      string
	AccountType string
	OwnerID     string
	Page        int
	PageSize    int
	SortColumn  string
	SortDir     string
}
