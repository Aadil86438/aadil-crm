package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"crm/models"
)

// ActivityRepository handles all activity database operations
type ActivityRepository struct {
	db *sql.DB
}

// NewActivityRepository creates a new ActivityRepository
func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func scanActivity(rows interface{ Scan(...interface{}) error }) (*models.Activity, error) {
	a := &models.Activity{}
	var ownerID, ownerName sql.NullString
	var relLeadID, relLeadFirst, relLeadLast sql.NullString
	var relContactID, relContactFirst, relContactLast sql.NullString
	var relAccountID, relAccountName sql.NullString
	var relDealID, relDealName sql.NullString
	var dueDate, completedAt sql.NullTime

	err := rows.Scan(
		&a.ID, &a.Type, &a.Subject, &a.Description, &a.Status,
		&dueDate, &completedAt,
		&ownerID, &ownerName,
		&relLeadID, &relLeadFirst, &relLeadLast,
		&relContactID, &relContactFirst, &relContactLast,
		&relAccountID, &relAccountName,
		&relDealID, &relDealName,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ownerID.Valid {
		a.OwnerID = &ownerID.String
	}
	a.OwnerName = ownerName.String
	if dueDate.Valid {
		a.DueDate = &dueDate.Time
	}
	if completedAt.Valid {
		a.CompletedAt = &completedAt.Time
	}
	if relLeadID.Valid {
		a.RelatedLeadID = &relLeadID.String
		a.RelatedLeadName = relLeadFirst.String + " " + relLeadLast.String
	}
	if relContactID.Valid {
		a.RelatedContactID = &relContactID.String
		a.RelatedContactName = relContactFirst.String + " " + relContactLast.String
	}
	if relAccountID.Valid {
		a.RelatedAccountID = &relAccountID.String
		a.RelatedAccountName = relAccountName.String
	}
	if relDealID.Valid {
		a.RelatedDealID = &relDealID.String
		a.RelatedDealName = relDealName.String
	}
	return a, nil
}

const activitySelectBase = `
	SELECT a.id, a.type, a.subject, a.description, a.status, a.due_date, a.completed_at,
		a.owner_id, u.name as owner_name,
		a.related_lead_id, l.first_name, l.last_name,
		a.related_contact_id, c.first_name, c.last_name,
		a.related_account_id, acc.name,
		a.related_deal_id, d.name,
		a.created_at, a.updated_at
	FROM activities a
	LEFT JOIN users u ON a.owner_id = u.id
	LEFT JOIN leads l ON a.related_lead_id = l.id
	LEFT JOIN contacts c ON a.related_contact_id = c.id
	LEFT JOIN accounts acc ON a.related_account_id = acc.id
	LEFT JOIN deals d ON a.related_deal_id = d.id
`

// List returns activities with optional filters
func (r *ActivityRepository) List(params models.ActivityListParams) ([]*models.Activity, int, error) {
	args := []interface{}{}
	conditions := []string{"a.deleted_at IS NULL"}
	argIdx := 1

	if params.Type != "" {
		conditions = append(conditions, fmt.Sprintf("a.type = $%d", argIdx))
		args = append(args, params.Type)
		argIdx++
	}
	if params.RelatedLeadID != "" {
		conditions = append(conditions, fmt.Sprintf("a.related_lead_id = $%d", argIdx))
		args = append(args, params.RelatedLeadID)
		argIdx++
	}
	if params.RelatedContactID != "" {
		conditions = append(conditions, fmt.Sprintf("a.related_contact_id = $%d", argIdx))
		args = append(args, params.RelatedContactID)
		argIdx++
	}
	if params.RelatedAccountID != "" {
		conditions = append(conditions, fmt.Sprintf("a.related_account_id = $%d", argIdx))
		args = append(args, params.RelatedAccountID)
		argIdx++
	}
	if params.RelatedDealID != "" {
		conditions = append(conditions, fmt.Sprintf("a.related_deal_id = $%d", argIdx))
		args = append(args, params.RelatedDealID)
		argIdx++
	}
	if params.OwnerID != "" {
		conditions = append(conditions, fmt.Sprintf("a.owner_id = $%d", argIdx))
		args = append(args, params.OwnerID)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")
	var total int
	r.db.QueryRow("SELECT COUNT(*) FROM activities a "+where, args...).Scan(&total)

	offset := (params.Page - 1) * params.PageSize
	query := activitySelectBase + where + fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		act, err := scanActivity(rows)
		if err != nil {
			return nil, 0, err
		}
		activities = append(activities, act)
	}
	return activities, total, nil
}

// FindByID finds an activity by ID
func (r *ActivityRepository) FindByID(id string) (*models.Activity, error) {
	row := r.db.QueryRow(activitySelectBase+"WHERE a.id = $1 AND a.deleted_at IS NULL", id)
	act, err := scanActivity(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return act, err
}

// Create creates a new activity
func (r *ActivityRepository) Create(req *models.CreateActivityRequest) (*models.Activity, error) {
	var id string
	err := r.db.QueryRow(`
		INSERT INTO activities (type, subject, description, status, due_date, owner_id,
			related_lead_id, related_contact_id, related_account_id, related_deal_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id
	`, req.Type, req.Subject, req.Description, req.Status, req.DueDate, req.OwnerID,
		req.RelatedLeadID, req.RelatedContactID, req.RelatedAccountID, req.RelatedDealID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// SoftDelete marks an activity as deleted
func (r *ActivityRepository) SoftDelete(id string) error {
	_, err := r.db.Exec("UPDATE activities SET deleted_at=$1 WHERE id=$2", time.Now(), id)
	return err
}

// CountDueToday returns count of activities due today
func (r *ActivityRepository) CountDueToday() (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM activities
		WHERE deleted_at IS NULL AND due_date::date = CURRENT_DATE AND completed_at IS NULL
	`).Scan(&count)
	return count, err
}

// CountByType returns activity counts per type
func (r *ActivityRepository) CountByType() (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT type, COUNT(*) FROM activities WHERE deleted_at IS NULL GROUP BY type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int{}
	for rows.Next() {
		var t string
		var count int
		rows.Scan(&t, &count)
		result[t] = count
	}
	return result, nil
}

// GetCalendarEvents returns activities for a date range
func (r *ActivityRepository) GetCalendarEvents(start, end time.Time) ([]*models.Activity, error) {
	rows, err := r.db.Query(
		activitySelectBase+`WHERE a.deleted_at IS NULL AND a.due_date BETWEEN $1 AND $2 ORDER BY a.due_date`,
		start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		act, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		activities = append(activities, act)
	}
	return activities, nil
}
