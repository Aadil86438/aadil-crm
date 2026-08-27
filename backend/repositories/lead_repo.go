package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"crm/models"
)

// LeadRepository handles all lead database operations
type LeadRepository struct {
	db *sql.DB
}

// NewLeadRepository creates a new LeadRepository
func NewLeadRepository(db *sql.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

func scanLead(rows interface {
	Scan(...interface{}) error
}) (*models.Lead, error) {
	l := &models.Lead{}
	var ownerName, ownerID sql.NullString
	var annualRevenue sql.NullFloat64
	var numEmployees sql.NullInt64
	var convertedAt sql.NullTime

	err := rows.Scan(
		&l.ID, &l.FirstName, &l.LastName, &l.Company, &l.Email, &l.Phone, &l.Mobile,
		&l.Website, &l.LeadSource, &l.LeadStatus, &l.Industry, &l.JobTitle,
		&annualRevenue, &numEmployees, &l.Rating, &l.Address, &l.City, &l.State,
		&l.Country, &l.Description, &ownerID, &ownerName, &l.IsConverted,
		&convertedAt, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ownerID.Valid {
		l.OwnerID = &ownerID.String
	}
	l.OwnerName = ownerName.String
	if annualRevenue.Valid {
		l.AnnualRevenue = &annualRevenue.Float64
	}
	if numEmployees.Valid {
		n := int(numEmployees.Int64)
		l.NumEmployees = &n
	}
	if convertedAt.Valid {
		l.ConvertedAt = &convertedAt.Time
	}
	return l, nil
}

const leadSelectBase = `
	SELECT l.id, l.first_name, l.last_name, l.company, l.email, l.phone, l.mobile,
		l.website, l.lead_source, l.lead_status, l.industry, l.job_title,
		l.annual_revenue, l.num_employees, l.rating, l.address, l.city, l.state,
		l.country, l.description, l.owner_id, u.name as owner_name, l.is_converted,
		l.converted_at, l.created_at, l.updated_at
	FROM leads l
	LEFT JOIN users u ON l.owner_id = u.id
`

// List returns a paginated list of leads
func (r *LeadRepository) List(params models.LeadListParams) ([]*models.Lead, int, error) {
	args := []interface{}{}
	conditions := []string{"l.deleted_at IS NULL"}
	argIdx := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(l.first_name ILIKE $%d OR l.last_name ILIKE $%d OR l.company ILIKE $%d OR l.email ILIKE $%d)",
			argIdx, argIdx+1, argIdx+2, argIdx+3,
		))
		search := "%" + params.Search + "%"
		args = append(args, search, search, search, search)
		argIdx += 4
	}
	if params.Status != "" {
		conditions = append(conditions, fmt.Sprintf("l.lead_status = $%d", argIdx))
		args = append(args, params.Status)
		argIdx++
	}
	if params.Source != "" {
		conditions = append(conditions, fmt.Sprintf("l.lead_source = $%d", argIdx))
		args = append(args, params.Source)
		argIdx++
	}
	if params.OwnerID != "" {
		conditions = append(conditions, fmt.Sprintf("l.owner_id = $%d", argIdx))
		args = append(args, params.OwnerID)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count
	var total int
	countQuery := "SELECT COUNT(*) FROM leads l " + where
	r.db.QueryRow(countQuery, args...).Scan(&total)

	// Data
	sortCol := "l.created_at"
	allowedSort := map[string]string{
		"first_name": "l.first_name", "last_name": "l.last_name",
		"company": "l.company", "email": "l.email",
		"lead_status": "l.lead_status", "created_at": "l.created_at",
	}
	if col, ok := allowedSort[params.SortColumn]; ok {
		sortCol = col
	}
	sortDir := "DESC"
	if strings.ToUpper(params.SortDir) == "ASC" {
		sortDir = "ASC"
	}

	offset := (params.Page - 1) * params.PageSize
	query := leadSelectBase + where + fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var leads []*models.Lead
	for rows.Next() {
		l, err := scanLead(rows)
		if err != nil {
			return nil, 0, err
		}
		leads = append(leads, l)
	}
	return leads, total, nil
}

// FindByID finds a lead by ID
func (r *LeadRepository) FindByID(id string) (*models.Lead, error) {
	row := r.db.QueryRow(leadSelectBase+"WHERE l.id = $1 AND l.deleted_at IS NULL", id)
	l, err := scanLead(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return l, err
}

// Create creates a new lead
func (r *LeadRepository) Create(req *models.CreateLeadRequest) (*models.Lead, error) {
	var id string
	err := r.db.QueryRow(`
		INSERT INTO leads (first_name, last_name, company, email, phone, mobile, website,
			lead_source, lead_status, industry, job_title, annual_revenue, num_employees,
			rating, address, city, state, country, description, owner_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,COALESCE(NULLIF($9,''),'New'),$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		RETURNING id
	`, req.FirstName, req.LastName, req.Company, req.Email, req.Phone, req.Mobile, req.Website,
		req.LeadSource, req.LeadStatus, req.Industry, req.JobTitle,
		req.AnnualRevenue, req.NumEmployees, req.Rating,
		req.Address, req.City, req.State, req.Country, req.Description, req.OwnerID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// Update updates a lead
func (r *LeadRepository) Update(id string, req *models.UpdateLeadRequest) (*models.Lead, error) {
	_, err := r.db.Exec(`
		UPDATE leads SET first_name=$1, last_name=$2, company=$3, email=$4, phone=$5, mobile=$6,
			website=$7, lead_source=$8, lead_status=$9, industry=$10, job_title=$11,
			annual_revenue=$12, num_employees=$13, rating=$14, address=$15, city=$16,
			state=$17, country=$18, description=$19, owner_id=$20, updated_at=NOW()
		WHERE id=$21 AND deleted_at IS NULL
	`, req.FirstName, req.LastName, req.Company, req.Email, req.Phone, req.Mobile,
		req.Website, req.LeadSource, req.LeadStatus, req.Industry, req.JobTitle,
		req.AnnualRevenue, req.NumEmployees, req.Rating,
		req.Address, req.City, req.State, req.Country, req.Description, req.OwnerID, id,
	)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// SoftDelete marks a lead as deleted
func (r *LeadRepository) SoftDelete(id string) error {
	_, err := r.db.Exec("UPDATE leads SET deleted_at=$1 WHERE id=$2", time.Now(), id)
	return err
}

// MarkConverted marks a lead as converted
func (r *LeadRepository) MarkConverted(id string) error {
	_, err := r.db.Exec(`
		UPDATE leads SET is_converted=TRUE, converted_at=NOW(), lead_status='Converted', updated_at=NOW()
		WHERE id=$1
	`, id)
	return err
}

// IsAlreadyConverted checks if a lead is already converted
func (r *LeadRepository) IsAlreadyConverted(id string) (bool, error) {
	var isConverted bool
	err := r.db.QueryRow("SELECT is_converted FROM leads WHERE id=$1 AND deleted_at IS NULL", id).Scan(&isConverted)
	return isConverted, err
}

// CountByStatus returns count of leads grouped by status
func (r *LeadRepository) CountByStatus() (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT lead_status, COUNT(*) FROM leads WHERE deleted_at IS NULL GROUP BY lead_status
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int{}
	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		result[status] = count
	}
	return result, nil
}

// CountBySource returns count of leads grouped by source
func (r *LeadRepository) CountBySource() (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT COALESCE(lead_source, 'Unknown'), COUNT(*) FROM leads WHERE deleted_at IS NULL GROUP BY lead_source
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int{}
	for rows.Next() {
		var source string
		var count int
		rows.Scan(&source, &count)
		result[source] = count
	}
	return result, nil
}

// RecentLeads returns the most recent leads
func (r *LeadRepository) RecentLeads(limit int) ([]*models.Lead, error) {
	params := models.LeadListParams{Page: 1, PageSize: limit, SortColumn: "created_at", SortDir: "DESC"}
	leads, _, err := r.List(params)
	return leads, err
}

// Search searches leads by term
func (r *LeadRepository) Search(term string, limit int) ([]*models.Lead, error) {
	params := models.LeadListParams{Search: term, Page: 1, PageSize: limit}
	leads, _, err := r.List(params)
	return leads, err
}
