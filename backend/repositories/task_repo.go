package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"crm/models"
)

// TaskRepository handles all task database operations
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new TaskRepository
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func scanTask(rows interface{ Scan(...interface{}) error }) (*models.Task, error) {
	t := &models.Task{}
	var ownerID, ownerName sql.NullString
	var relLeadID, relLeadFirst, relLeadLast sql.NullString
	var relContactID, relContactFirst, relContactLast sql.NullString
	var relAccountID, relAccountName sql.NullString
	var relDealID, relDealName sql.NullString
	var dueDate sql.NullTime

	err := rows.Scan(
		&t.ID, &t.Subject, &t.Description, &dueDate, &t.Priority, &t.Status,
		&ownerID, &ownerName,
		&relLeadID, &relLeadFirst, &relLeadLast,
		&relContactID, &relContactFirst, &relContactLast,
		&relAccountID, &relAccountName,
		&relDealID, &relDealName,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ownerID.Valid {
		t.OwnerID = &ownerID.String
	}
	t.OwnerName = ownerName.String
	if dueDate.Valid {
		t.DueDate = &dueDate.Time
	}
	if relLeadID.Valid {
		t.RelatedLeadID = &relLeadID.String
		t.RelatedLeadName = relLeadFirst.String + " " + relLeadLast.String
	}
	if relContactID.Valid {
		t.RelatedContactID = &relContactID.String
		t.RelatedContactName = relContactFirst.String + " " + relContactLast.String
	}
	if relAccountID.Valid {
		t.RelatedAccountID = &relAccountID.String
		t.RelatedAccountName = relAccountName.String
	}
	if relDealID.Valid {
		t.RelatedDealID = &relDealID.String
		t.RelatedDealName = relDealName.String
	}
	return t, nil
}

const taskSelectBase = `
	SELECT t.id, t.subject, t.description, t.due_date, t.priority, t.status,
		t.owner_id, u.name as owner_name,
		t.related_lead_id, l.first_name, l.last_name,
		t.related_contact_id, c.first_name, c.last_name,
		t.related_account_id, a.name,
		t.related_deal_id, d.name,
		t.created_at, t.updated_at
	FROM tasks t
	LEFT JOIN users u ON t.owner_id = u.id
	LEFT JOIN leads l ON t.related_lead_id = l.id
	LEFT JOIN contacts c ON t.related_contact_id = c.id
	LEFT JOIN accounts a ON t.related_account_id = a.id
	LEFT JOIN deals d ON t.related_deal_id = d.id
`

// List returns paginated tasks
func (r *TaskRepository) List(params models.TaskListParams) ([]*models.Task, int, error) {
	args := []interface{}{}
	conditions := []string{"t.deleted_at IS NULL"}
	argIdx := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("t.subject ILIKE $%d", argIdx))
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}
	if params.Status != "" {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argIdx))
		args = append(args, params.Status)
		argIdx++
	}
	if params.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("t.priority = $%d", argIdx))
		args = append(args, params.Priority)
		argIdx++
	}
	if params.OwnerID != "" {
		conditions = append(conditions, fmt.Sprintf("t.owner_id = $%d", argIdx))
		args = append(args, params.OwnerID)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")
	var total int
	r.db.QueryRow("SELECT COUNT(*) FROM tasks t "+where, args...).Scan(&total)

	sortDir := "ASC"
	if strings.ToUpper(params.SortDir) == "DESC" {
		sortDir = "DESC"
	}

	offset := (params.Page - 1) * params.PageSize
	query := taskSelectBase + where + fmt.Sprintf(" ORDER BY t.due_date %s NULLS LAST, t.created_at DESC LIMIT $%d OFFSET $%d", sortDir, argIdx, argIdx+1)
	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, nil
}

// FindByID finds a task by ID
func (r *TaskRepository) FindByID(id string) (*models.Task, error) {
	row := r.db.QueryRow(taskSelectBase+"WHERE t.id = $1 AND t.deleted_at IS NULL", id)
	t, err := scanTask(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

// Create creates a new task
func (r *TaskRepository) Create(req *models.CreateTaskRequest) (*models.Task, error) {
	var id string
	err := r.db.QueryRow(`
		INSERT INTO tasks (subject, description, due_date, priority, status, owner_id,
			related_lead_id, related_contact_id, related_account_id, related_deal_id)
		VALUES ($1,$2,$3,COALESCE(NULLIF($4,''),'Medium'),COALESCE(NULLIF($5,''),'Not Started'),$6,$7,$8,$9,$10)
		RETURNING id
	`, req.Subject, req.Description, req.DueDate, req.Priority, req.Status, req.OwnerID,
		req.RelatedLeadID, req.RelatedContactID, req.RelatedAccountID, req.RelatedDealID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// Update updates a task
func (r *TaskRepository) Update(id string, req *models.UpdateTaskRequest) (*models.Task, error) {
	_, err := r.db.Exec(`
		UPDATE tasks SET subject=$1, description=$2, due_date=$3, priority=$4, status=$5,
			owner_id=$6, related_lead_id=$7, related_contact_id=$8, related_account_id=$9,
			related_deal_id=$10, updated_at=NOW()
		WHERE id=$11 AND deleted_at IS NULL
	`, req.Subject, req.Description, req.DueDate, req.Priority, req.Status, req.OwnerID,
		req.RelatedLeadID, req.RelatedContactID, req.RelatedAccountID, req.RelatedDealID, id,
	)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// SoftDelete marks a task as deleted
func (r *TaskRepository) SoftDelete(id string) error {
	_, err := r.db.Exec("UPDATE tasks SET deleted_at=$1 WHERE id=$2", time.Now(), id)
	return err
}

// CountDueToday returns count of tasks due today
func (r *TaskRepository) CountDueToday() (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM tasks
		WHERE deleted_at IS NULL AND status NOT IN ('Completed', 'Deferred')
		AND due_date::date = CURRENT_DATE
	`).Scan(&count)
	return count, err
}

// CountUpcoming returns count of upcoming tasks (not completed, due in future)
func (r *TaskRepository) CountUpcoming() (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM tasks
		WHERE deleted_at IS NULL AND status NOT IN ('Completed', 'Deferred')
		AND due_date > NOW()
	`).Scan(&count)
	return count, err
}
