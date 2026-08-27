package models

import "time"

// RegistrationRequest represents a pending user registration
type RegistrationRequest struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	CompanyName    string    `json:"company_name"`
	TransactionID  string    `json:"transaction_id"`
	PaymentStatus  string    `json:"payment_status"`
	ApprovalStatus string    `json:"approval_status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// RegisterRequest is the request body for self-registration
type RegisterRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	CompanyName string `json:"company_name"`
}

// SubmitPaymentRequest is the request body for payment submission
type SubmitPaymentRequest struct {
	RegistrationID string `json:"registration_id"`
	TransactionID  string `json:"transaction_id"`
}

// AdminVerifyRequest is the request body for admin code verification
type AdminVerifyRequest struct {
	Code string `json:"code"`
}
