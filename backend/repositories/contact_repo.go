package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"crm/models"
)

// ContactRepository handles all contact database operations
type ContactRepository struct {
	db *sql.DB
}

// NewContactRepository creates a new ContactRepository
func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func scanContact(rows interface{ Scan(...interface{}) error }) (*models.Contact, error) {
	c := &models.Contact{}
	var ownerID, ownerName, accountID, accountName sql.NullString

	err := rows.Scan(
		&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Phone, &c.Mobile,
		&c.JobTitle, &c.Department, &accountID, &accountName, &ownerID, &ownerName,
		&c.Address, &c.City, &c.State, &c.Country, &c.Description,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ownerID.Valid {
		c.OwnerID = &ownerID.String
	}
	c.OwnerName = ownerName.String
	if accountID.Valid {
		c.AccountID = &accountID.String
	}
	c.AccountName = accountName.String
	return c, nil
}

const contactSelectBase = `
	SELECT c.id, c.first_name, c.last_name, c.email, c.phone, c.mobile,
		c.job_title, c.department, c.account_id, a.name as account_name,
		c.owner_id, u.name as owner_name,
		c.address, c.city, c.state, c.country, c.description,
		c.created_at, c.updated_at
	FROM contacts c
	LEFT JOIN accounts a ON c.account_id = a.id
	LEFT JOIN users u ON c.owner_id = u.id
`

// List returns paginated contacts
func (r *ContactRepository) List(params models.ContactListParams) ([]*models.Contact, int, error) {
	args := []interface{}{}
	conditions := []string{"c.deleted_at IS NULL"}
	argIdx := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(c.first_name ILIKE $%d OR c.last_name ILIKE $%d OR c.email ILIKE $%d)",
			argIdx, argIdx+1, argIdx+2,
		))
		search := "%" + params.Search + "%"
		args = append(args, search, search, search)
		argIdx += 3
	}
	if params.AccountID != "" {
		conditions = append(conditions, fmt.Sprintf("c.account_id = $%d", argIdx))
		args = append(args, params.AccountID)
		argIdx++
	}
	if params.OwnerID != "" {
		conditions = append(conditions, fmt.Sprintf("c.owner_id = $%d", argIdx))
		args = append(args, params.OwnerID)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")
	var total int
	r.db.QueryRow("SELECT COUNT(*) FROM contacts c "+where, args...).Scan(&total)

	sortCol := "c.created_at"
	allowedSort := map[string]string{
		"first_name": "c.first_name", "last_name": "c.last_name",
		"email": "c.email", "created_at": "c.created_at",
	}
	if col, ok := allowedSort[params.SortColumn]; ok {
		sortCol = col
	}
	sortDir := "DESC"
	if strings.ToUpper(params.SortDir) == "ASC" {
		sortDir = "ASC"
	}

	offset := (params.Page - 1) * params.PageSize
	query := contactSelectBase + where + fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contacts []*models.Contact
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, 0, err
		}
		contacts = append(contacts, c)
	}
	return contacts, total, nil
}

// FindByID finds a contact by ID
func (r *ContactRepository) FindByID(id string) (*models.Contact, error) {
	row := r.db.QueryRow(contactSelectBase+"WHERE c.id = $1 AND c.deleted_at IS NULL", id)
	c, err := scanContact(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

// Create creates a new contact
func (r *ContactRepository) Create(req *models.CreateContactRequest) (*models.Contact, error) {
	var id string
	err := r.db.QueryRow(`
		INSERT INTO contacts (first_name, last_name, email, phone, mobile, job_title,
			department, account_id, owner_id, address, city, state, country, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id
	`, req.FirstName, req.LastName, req.Email, req.Phone, req.Mobile, req.JobTitle,
		req.Department, req.AccountID, req.OwnerID,
		req.Address, req.City, req.State, req.Country, req.Description,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// Update updates a contact
func (r *ContactRepository) Update(id string, req *models.UpdateContactRequest) (*models.Contact, error) {
	_, err := r.db.Exec(`
		UPDATE contacts SET first_name=$1, last_name=$2, email=$3, phone=$4, mobile=$5,
			job_title=$6, department=$7, account_id=$8, owner_id=$9,
			address=$10, city=$11, state=$12, country=$13, description=$14, updated_at=NOW()
		WHERE id=$15 AND deleted_at IS NULL
	`, req.FirstName, req.LastName, req.Email, req.Phone, req.Mobile, req.JobTitle,
		req.Department, req.AccountID, req.OwnerID,
		req.Address, req.City, req.State, req.Country, req.Description, id,
	)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// SoftDelete marks a contact as deleted
func (r *ContactRepository) SoftDelete(id string) error {
	_, err := r.db.Exec("UPDATE contacts SET deleted_at=$1 WHERE id=$2", time.Now(), id)
	return err
}

// Count returns total non-deleted contacts
func (r *ContactRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM contacts WHERE deleted_at IS NULL").Scan(&count)
	return count, err
}

// Search searches contacts by term
func (r *ContactRepository) Search(term string, limit int) ([]*models.Contact, error) {
	params := models.ContactListParams{Search: term, Page: 1, PageSize: limit}
	contacts, _, err := r.List(params)
	return contacts, err
}

// ListSimple returns contacts for dropdown
func (r *ContactRepository) ListSimple() ([]*models.Contact, error) {
	rows, err := r.db.Query(`SELECT id, first_name, last_name, email FROM contacts WHERE deleted_at IS NULL ORDER BY first_name LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contacts []*models.Contact
	for rows.Next() {
		c := &models.Contact{}
		rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email)
		contacts = append(contacts, c)
	}
	return contacts, nil
}
