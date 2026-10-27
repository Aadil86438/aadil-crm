package models

import (
	"time"
)

// User represents a CRM system user
type User struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	CRMCode        string     `json:"crm_code"`
	ApprovalStatus string     `json:"approval_status"`
	CompanyName    string     `json:"company_name"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"-"`
}

// CreateUserRequest is the request body for creating a user
type CreateUserRequest struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	Role           string `json:"role"`
	CompanyName    string `json:"company_name"`
	ApprovalStatus string `json:"approval_status"`
	CRMCode        string `json:"crm_code"`
}

// UpdateUserRequest is the request body for updating a user
type UpdateUserRequest struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	Status         string `json:"status"`
	ApprovalStatus string `json:"approval_status"`
	CRMCode        string `json:"crm_code"`
}

// ChangePasswordRequest is the request body for changing a password
type ChangePasswordRequest struct {
	Password string `json:"password"`
}

// LoginRequest is the request body for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	CRMCode  string `json:"crm_code"`
}

// LoginResponse is the response for a successful login
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

