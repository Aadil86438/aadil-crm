package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"crm/models"
)

// DealRepository handles all deal database operations
type DealRepository struct {
	db *sql.DB
}

// NewDealRepository creates a new DealRepository
func NewDealRepository(db *sql.DB) *DealRepository {
	return &DealRepository{db: db}
}

func scanDeal(rows interface{ Scan(...interface{}) error }) (*models.Deal, error) {
	d := &models.Deal{}
	var ownerID, ownerName, accountID, accountName, contactID, contactName sql.NullString
	var amount sql.NullFloat64
	var closeDate sql.NullTime

	err := rows.Scan(
		&d.ID, &d.Name, &amount, &accountID, &accountName,
		&contactID, &contactName, &d.Stage, &d.Probability,
		&closeDate, &d.LeadSource, &ownerID, &ownerName,
		&d.Description, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ownerID.Valid {
		d.OwnerID = &ownerID.String
	}
	d.OwnerName = ownerName.String
	if accountID.Valid {
		d.AccountID = &accountID.String
	}
	d.AccountName = accountName.String
	if contactID.Valid {
		d.ContactID = &contactID.String
	}
	d.ContactName = contactName.String
	if amount.Valid {
		d.Amount = &amount.Float64
	}
	if closeDate.Valid {
		d.ExpectedCloseDate = &closeDate.Time
	}
	return d, nil
}

const dealSelectBase = `
	SELECT d.id, d.name, d.amount, d.account_id, a.name as account_name,
		d.contact_id, CONCAT(c.first_name, ' ', c.last_name) as contact_name,
		d.stage, d.probability, d.expected_close_date, d.lead_source,
		d.owner_id, u.name as owner_name, d.description, d.created_at, d.updated_at
	FROM deals d
	LEFT JOIN accounts a ON d.account_id = a.id
	LEFT JOIN contacts c ON d.contact_id = c.id
	LEFT JOIN users u ON d.owner_id = u.id
`

// List returns paginated deals
func (r *DealRepository) List(params models.DealListParams) ([]*models.Deal, int, error) {
	args := []interface{}{}
	conditions := []string{"d.deleted_at IS NULL"}
	argIdx := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("d.name ILIKE $%d", argIdx))
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}
	if params.Stage != "" {
		conditions = append(conditions, fmt.Sprintf("d.stage = $%d", argIdx))
		args = append(args, params.Stage)
		argIdx++
	}
	if params.AccountID != "" {
		conditions = append(conditions, fmt.Sprintf("d.account_id = $%d", argIdx))
		args = append(args, params.AccountID)
		argIdx++
	}
	if params.OwnerID != "" {
		conditions = append(conditions, fmt.Sprintf("d.owner_id = $%d", argIdx))
		args = append(args, params.OwnerID)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")
	var total int
	r.db.QueryRow("SELECT COUNT(*) FROM deals d "+where, args...).Scan(&total)

	sortCol := "d.created_at"
	allowedSort := map[string]string{
		"name": "d.name", "amount": "d.amount", "stage": "d.stage",
		"expected_close_date": "d.expected_close_date", "created_at": "d.created_at",
	}
	if col, ok := allowedSort[params.SortColumn]; ok {
		sortCol = col
	}
	sortDir := "DESC"
	if strings.ToUpper(params.SortDir) == "ASC" {
		sortDir = "ASC"
	}

	offset := (params.Page - 1) * params.PageSize
	query := dealSelectBase + where + fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortCol, sortDir, argIdx, argIdx+1)
	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var deals []*models.Deal
	for rows.Next() {
		d, err := scanDeal(rows)
		if err != nil {
			return nil, 0, err
		}
		deals = append(deals, d)
	}
	return deals, total, nil
}

// FindByID finds a deal by ID
func (r *DealRepository) FindByID(id string) (*models.Deal, error) {
	row := r.db.QueryRow(dealSelectBase+"WHERE d.id = $1 AND d.deleted_at IS NULL", id)
	d, err := scanDeal(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return d, err
}

// Create creates a new deal
func (r *DealRepository) Create(req *models.CreateDealRequest) (*models.Deal, error) {
	var id string
	err := r.db.QueryRow(`
		INSERT INTO deals (name, amount, account_id, contact_id, stage, probability,
			expected_close_date, lead_source, owner_id, description)
		VALUES ($1,$2,$3,$4,COALESCE(NULLIF($5,''),'Qualification'),$6,$7,$8,$9,$10)
		RETURNING id
	`, req.Name, req.Amount, req.AccountID, req.ContactID, req.Stage, req.Probability,
		req.ExpectedCloseDate, req.LeadSource, req.OwnerID, req.Description,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// Update updates a deal
func (r *DealRepository) Update(id string, req *models.UpdateDealRequest) (*models.Deal, error) {
	_, err := r.db.Exec(`
		UPDATE deals SET name=$1, amount=$2, account_id=$3, contact_id=$4, stage=$5,
			probability=$6, expected_close_date=$7, lead_source=$8, owner_id=$9,
			description=$10, updated_at=NOW()
		WHERE id=$11 AND deleted_at IS NULL
	`, req.Name, req.Amount, req.AccountID, req.ContactID, req.Stage, req.Probability,
		req.ExpectedCloseDate, req.LeadSource, req.OwnerID, req.Description, id,
	)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// UpdateStage updates only the stage of a deal
func (r *DealRepository) UpdateStage(id, stage string) (*models.Deal, error) {
	probability := stageProbability(stage)
	_, err := r.db.Exec(`
		UPDATE deals SET stage=$1, probability=$2, updated_at=NOW() WHERE id=$3 AND deleted_at IS NULL
	`, stage, probability, id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// SoftDelete marks a deal as deleted
func (r *DealRepository) SoftDelete(id string) error {
	_, err := r.db.Exec("UPDATE deals SET deleted_at=$1 WHERE id=$2", time.Now(), id)
	return err
}

// GetPipelineSummary returns aggregated pipeline stats
func (r *DealRepository) GetPipelineSummary() (*models.PipelineSummary, error) {
	s := &models.PipelineSummary{}
	err := r.db.QueryRow(`
		SELECT
			COUNT(*) as total_deals,
			COALESCE(SUM(amount), 0) as total_value,
			COALESCE(SUM(amount * probability / 100.0), 0) as weighted_value,
			COUNT(*) FILTER (WHERE stage = 'Closed Won') as won_deals,
			COALESCE(SUM(amount) FILTER (WHERE stage = 'Closed Won'), 0) as won_revenue,
			COUNT(*) FILTER (WHERE stage = 'Closed Lost') as lost_deals,
			COALESCE(SUM(amount) FILTER (WHERE stage = 'Closed Lost'), 0) as lost_revenue
		FROM deals WHERE deleted_at IS NULL
	`).Scan(&s.TotalDeals, &s.TotalValue, &s.WeightedValue, &s.WonDeals, &s.WonRevenue, &s.LostDeals, &s.LostRevenue)
	return s, err
}

// CountByStage returns deal counts per stage
func (r *DealRepository) CountByStage() (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT stage, COUNT(*) FROM deals WHERE deleted_at IS NULL GROUP BY stage
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int{}
	for rows.Next() {
		var stage string
		var count int
		rows.Scan(&stage, &count)
		result[stage] = count
	}
	return result, nil
}

// RecentDeals returns the most recent deals
func (r *DealRepository) RecentDeals(limit int) ([]*models.Deal, error) {
	params := models.DealListParams{Page: 1, PageSize: limit, SortColumn: "created_at", SortDir: "DESC"}
	deals, _, err := r.List(params)
	return deals, err
}

// Search searches deals by term
func (r *DealRepository) Search(term string, limit int) ([]*models.Deal, error) {
	params := models.DealListParams{Search: term, Page: 1, PageSize: limit}
	deals, _, err := r.List(params)
	return deals, err
}

// MonthlyRevenue returns monthly won revenue for the past 12 months
func (r *DealRepository) MonthlyRevenue() ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT TO_CHAR(DATE_TRUNC('month', created_at), 'Mon YYYY') as month,
			COALESCE(SUM(amount), 0) as revenue
		FROM deals
		WHERE deleted_at IS NULL AND stage = 'Closed Won'
			AND created_at >= NOW() - INTERVAL '12 months'
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var month string
		var revenue float64
		rows.Scan(&month, &revenue)
		result = append(result, map[string]interface{}{"month": month, "revenue": revenue})
	}
	return result, nil
}

// stageProbability returns the default probability for a stage
func stageProbability(stage string) int {
	probs := map[string]int{
		"Qualification": 10, "Needs Analysis": 20, "Proposal": 40,
		"Negotiation": 60, "Closed Won": 100, "Closed Lost": 0,
	}
	if p, ok := probs[stage]; ok {
		return p
	}
	return 10
}
