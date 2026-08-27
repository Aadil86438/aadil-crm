package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"crm/models"
)

// AccountRepository handles all account database operations
type AccountRepository struct {
	db *sql.DB
}

// NewAccountRepository creates a new AccountRepository
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func scanAccount(rows interface{ Scan(...interface{}) error }) (*models.Account, error) {
	a := &models.Account{}
	var ownerID, ownerName sql.NullString
	var numEmployees sql.NullInt64
	var annualRevenue sql.NullFloat64

	err := rows.Scan(
		&a.ID, &a.Name, &a.Website, &a.Industry, &a.Phone, &a.Email,
		&numEmployees, &annualRevenue, &a.AccountType, &ownerID, &ownerName,
		&a.Address, &a.City, &a.State, &a.Country, &a.Description,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ownerID.Valid {
		a.OwnerID = &ownerID.String
	}
	a.OwnerName = ownerName.String
	if numEmployees.Valid {
		n := int(numEmployees.Int64)
		a.NumEmployees = &n
	}
	if annualRevenue.Valid {
		a.AnnualRevenue = &annualRevenue.Float64
	}
	return a, nil
}

const accountSelectBase = `
	SELECT a.id, a.name, a.website, a.industry, a.phone, a.email,
		a.num_employees, a.annual_revenue, a.account_type, a.owner_id, u.name as owner_name,
		a.address, a.city, a.state, a.country, a.description, a.created_at, a.updated_at
	FROM accounts a
	LEFT JOIN users u ON a.owner_id = u.id
`

// List returns paginated accounts
func (r *AccountRepository) List(params models.AccountListParams) ([]*models.Account, int, error) {
	args := []interface{}{}
	conditions := []string{"a.deleted_at IS NULL"}
	argIdx := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(a.name ILIKE $%d OR a.email ILIKE $%d OR a.industry ILIKE $%d)",
			argIdx, argIdx+1, argIdx+2,
		))
		search := "%" + params.Search + "%"
		args = append(args, search, search, search)
		argIdx += 3
	}
	if params.AccountType != "" {
		conditions = append(conditions, fmt.Sprintf("a.account_type = $%d", argIdx))
		args = append(args, params.AccountType)
		argIdx++
	}
	if params.OwnerID != "" {
		conditions = append(conditions, fmt.Sprintf("a.owner_id = $%d", argIdx))
		args = append(args, params.OwnerID)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")
	var total int
	r.db.QueryRow("SELECT COUNT(*) FROM accounts a "+where, args...).Scan(&total)

	sortCol := "a.created_at"
	allowedSort := map[string]string{
		"name": "a.name", "account_type": "a.account_type", "created_at": "a.created_at",
	}
	if col, ok := allowedSort[params.SortColumn]; ok {
		sortCol = col
	}
	sortDir := "DESC"
	if strings.ToUpper(params.SortDir) == "ASC" {
		sortDir = "ASC"
	}

	offset := (params.Page - 1) * params.PageSize
	query := accountSelectBase + where + fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var accounts []*models.Account
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, 0, err
		}
		accounts = append(accounts, a)
	}
	return accounts, total, nil
}

// FindByID finds an account by ID
func (r *AccountRepository) FindByID(id string) (*models.Account, error) {
	row := r.db.QueryRow(accountSelectBase+"WHERE a.id = $1 AND a.deleted_at IS NULL", id)
	a, err := scanAccount(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return a, err
}

// Create creates a new account
func (r *AccountRepository) Create(req *models.CreateAccountRequest) (*models.Account, error) {
	var id string
	err := r.db.QueryRow(`
		INSERT INTO accounts (name, website, industry, phone, email, num_employees,
			annual_revenue, account_type, owner_id, address, city, state, country, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,COALESCE(NULLIF($8,''),'Prospect'),$9,$10,$11,$12,$13,$14)
		RETURNING id
	`, req.Name, req.Website, req.Industry, req.Phone, req.Email, req.NumEmployees,
		req.AnnualRevenue, req.AccountType, req.OwnerID,
		req.Address, req.City, req.State, req.Country, req.Description,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// Update updates an account
func (r *AccountRepository) Update(id string, req *models.UpdateAccountRequest) (*models.Account, error) {
	_, err := r.db.Exec(`
		UPDATE accounts SET name=$1, website=$2, industry=$3, phone=$4, email=$5,
			num_employees=$6, annual_revenue=$7, account_type=$8, owner_id=$9,
			address=$10, city=$11, state=$12, country=$13, description=$14, updated_at=NOW()
		WHERE id=$15 AND deleted_at IS NULL
	`, req.Name, req.Website, req.Industry, req.Phone, req.Email, req.NumEmployees,
		req.AnnualRevenue, req.AccountType, req.OwnerID,
		req.Address, req.City, req.State, req.Country, req.Description, id,
	)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// SoftDelete marks an account as deleted
func (r *AccountRepository) SoftDelete(id string) error {
	_, err := r.db.Exec("UPDATE accounts SET deleted_at=$1 WHERE id=$2", time.Now(), id)
	return err
}

// Count returns total non-deleted accounts
func (r *AccountRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM accounts WHERE deleted_at IS NULL").Scan(&count)
	return count, err
}

// Search searches accounts by term
func (r *AccountRepository) Search(term string, limit int) ([]*models.Account, error) {
	params := models.AccountListParams{Search: term, Page: 1, PageSize: limit}
	accounts, _, err := r.List(params)
	return accounts, err
}

// ListSimple returns accounts for dropdown selection
func (r *AccountRepository) ListSimple() ([]*models.Account, error) {
	rows, err := r.db.Query(`SELECT id, name FROM accounts WHERE deleted_at IS NULL ORDER BY name LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []*models.Account
	for rows.Next() {
		a := &models.Account{}
		rows.Scan(&a.ID, &a.Name)
		accounts = append(accounts, a)
	}
	return accounts, nil
}
