package repositories

import (
	"database/sql"

	"crm/models"
)

// RegistrationRepository handles registration_requests database operations
type RegistrationRepository struct {
	db *sql.DB
}

// NewRegistrationRepository creates a new RegistrationRepository
func NewRegistrationRepository(db *sql.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

// Create inserts a new registration request
func (r *RegistrationRepository) Create(req *models.RegisterRequest, passwordHash string) (*models.RegistrationRequest, error) {
	reg := &models.RegistrationRequest{}
	err := r.db.QueryRow(`
		INSERT INTO registration_requests (name, email, password_hash, company_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, company_name, transaction_id, payment_status, approval_status, created_at, updated_at
	`, req.Name, req.Email, passwordHash, req.CompanyName).Scan(
		&reg.ID, &reg.Name, &reg.Email, &reg.CompanyName,
		&reg.TransactionID, &reg.PaymentStatus, &reg.ApprovalStatus,
		&reg.CreatedAt, &reg.UpdatedAt,
	)
	return reg, err
}

// FindByID finds a registration request by ID
func (r *RegistrationRepository) FindByID(id string) (*models.RegistrationRequest, error) {
	reg := &models.RegistrationRequest{}
	err := r.db.QueryRow(`
		SELECT id, name, email, password_hash, company_name, transaction_id, payment_status, approval_status, created_at, updated_at
		FROM registration_requests WHERE id = $1
	`, id).Scan(
		&reg.ID, &reg.Name, &reg.Email, &reg.PasswordHash, &reg.CompanyName,
		&reg.TransactionID, &reg.PaymentStatus, &reg.ApprovalStatus,
		&reg.CreatedAt, &reg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return reg, err
}

// FindByEmail checks if a registration request exists for this email
func (r *RegistrationRepository) FindByEmail(email string) (*models.RegistrationRequest, error) {
	reg := &models.RegistrationRequest{}
	err := r.db.QueryRow(`
		SELECT id, name, email, company_name, transaction_id, payment_status, approval_status, created_at, updated_at
		FROM registration_requests WHERE email = $1
	`, email).Scan(
		&reg.ID, &reg.Name, &reg.Email, &reg.CompanyName,
		&reg.TransactionID, &reg.PaymentStatus, &reg.ApprovalStatus,
		&reg.CreatedAt, &reg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return reg, err
}

// SubmitPayment updates a registration request with transaction ID
func (r *RegistrationRepository) SubmitPayment(id, transactionID string) error {
	_, err := r.db.Exec(`
		UPDATE registration_requests
		SET transaction_id = $1, payment_status = 'submitted', updated_at = NOW()
		WHERE id = $2 AND payment_status = 'pending'
	`, transactionID, id)
	return err
}

// ListPending returns all registration requests with payment submitted and pending approval
func (r *RegistrationRepository) ListPending() ([]*models.RegistrationRequest, error) {
	rows, err := r.db.Query(`
		SELECT id, name, email, company_name, transaction_id, payment_status, approval_status, created_at, updated_at
		FROM registration_requests
		WHERE payment_status = 'submitted' AND approval_status = 'pending'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regs []*models.RegistrationRequest
	for rows.Next() {
		reg := &models.RegistrationRequest{}
		if err := rows.Scan(
			&reg.ID, &reg.Name, &reg.Email, &reg.CompanyName,
			&reg.TransactionID, &reg.PaymentStatus, &reg.ApprovalStatus,
			&reg.CreatedAt, &reg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		regs = append(regs, reg)
	}
	return regs, nil
}

// ListAll returns all registration requests (pending, approved, rejected)
func (r *RegistrationRepository) ListAll() ([]*models.RegistrationRequest, error) {
	rows, err := r.db.Query(`
		SELECT id, name, email, company_name, transaction_id, payment_status, approval_status, created_at, updated_at
		FROM registration_requests
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regs []*models.RegistrationRequest
	for rows.Next() {
		reg := &models.RegistrationRequest{}
		if err := rows.Scan(
			&reg.ID, &reg.Name, &reg.Email, &reg.CompanyName,
			&reg.TransactionID, &reg.PaymentStatus, &reg.ApprovalStatus,
			&reg.CreatedAt, &reg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		regs = append(regs, reg)
	}
	return regs, nil
}

// Approve marks a registration request as approved
func (r *RegistrationRepository) Approve(id string) error {
	_, err := r.db.Exec(`
		UPDATE registration_requests SET approval_status = 'approved', updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

// Reject marks a registration request as rejected
func (r *RegistrationRepository) Reject(id string) error {
	_, err := r.db.Exec(`
		UPDATE registration_requests SET approval_status = 'rejected', updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

// Delete removes a registration request
func (r *RegistrationRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM registration_requests WHERE id = $1`, id)
	return err
}
