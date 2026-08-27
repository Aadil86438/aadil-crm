package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"crm/models"
)

// NoteRepository handles all note database operations
type NoteRepository struct {
	db *sql.DB
}

// NewNoteRepository creates a new NoteRepository
func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

const noteSelectBase = `
	SELECT n.id, n.title, n.body, n.owner_id, u.name as owner_name,
		n.related_lead_id, n.related_contact_id, n.related_account_id, n.related_deal_id,
		n.created_at, n.updated_at
	FROM notes n
	LEFT JOIN users u ON n.owner_id = u.id
`

func scanNote(rows interface{ Scan(...interface{}) error }) (*models.Note, error) {
	n := &models.Note{}
	var ownerID, ownerName, title sql.NullString
	var relLeadID, relContactID, relAccountID, relDealID sql.NullString

	err := rows.Scan(
		&n.ID, &title, &n.Body, &ownerID, &ownerName,
		&relLeadID, &relContactID, &relAccountID, &relDealID,
		&n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	n.Title = title.String
	if ownerID.Valid {
		n.OwnerID = &ownerID.String
	}
	n.OwnerName = ownerName.String
	if relLeadID.Valid {
		n.RelatedLeadID = &relLeadID.String
	}
	if relContactID.Valid {
		n.RelatedContactID = &relContactID.String
	}
	if relAccountID.Valid {
		n.RelatedAccountID = &relAccountID.String
	}
	if relDealID.Valid {
		n.RelatedDealID = &relDealID.String
	}
	return n, nil
}

// ListByEntity lists notes for a given entity
func (r *NoteRepository) ListByEntity(entityType, entityID string) ([]*models.Note, error) {
	colMap := map[string]string{
		"lead": "n.related_lead_id", "contact": "n.related_contact_id",
		"account": "n.related_account_id", "deal": "n.related_deal_id",
	}
	col, ok := colMap[entityType]
	if !ok {
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}

	rows, err := r.db.Query(
		noteSelectBase+fmt.Sprintf("WHERE %s = $1 AND n.deleted_at IS NULL ORDER BY n.created_at DESC", col),
		entityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		n, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// FindByID finds a note by ID
func (r *NoteRepository) FindByID(id string) (*models.Note, error) {
	row := r.db.QueryRow(noteSelectBase+"WHERE n.id = $1 AND n.deleted_at IS NULL", id)
	n, err := scanNote(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return n, err
}

// Create creates a new note
func (r *NoteRepository) Create(req *models.CreateNoteRequest, ownerID string) (*models.Note, error) {
	var id string
	err := r.db.QueryRow(`
		INSERT INTO notes (title, body, owner_id, related_lead_id, related_contact_id, related_account_id, related_deal_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id
	`, nullableString(req.Title), req.Body, ownerID, req.RelatedLeadID, req.RelatedContactID,
		req.RelatedAccountID, req.RelatedDealID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// Update updates a note
func (r *NoteRepository) Update(id string, req *models.UpdateNoteRequest) (*models.Note, error) {
	_, err := r.db.Exec(`
		UPDATE notes SET title=$1, body=$2, updated_at=NOW() WHERE id=$3 AND deleted_at IS NULL
	`, nullableString(req.Title), req.Body, id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

// SoftDelete marks a note as deleted
func (r *NoteRepository) SoftDelete(id string) error {
	_, err := r.db.Exec("UPDATE notes SET deleted_at=$1 WHERE id=$2", time.Now(), id)
	return err
}

func nullableString(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates a new AuditRepository
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create records an audit log entry
func (r *AuditRepository) Create(req *models.CreateAuditLogRequest) error {
	_, err := r.db.Exec(`
		INSERT INTO audit_logs (user_id, user_email, action, entity, entity_id, metadata, ip_address)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, req.UserID, req.UserEmail, req.Action, req.Entity, req.EntityID, nil, req.IPAddress)
	return err
}

// List returns paginated audit logs
func (r *AuditRepository) List(page, pageSize int) ([]*models.AuditLog, int, error) {
	offset := (page - 1) * pageSize
	rows, err := r.db.Query(`
		SELECT id, user_id, user_email, action, entity, entity_id, ip_address, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		al := &models.AuditLog{}
		var userID, entityID sql.NullString
		err := rows.Scan(&al.ID, &userID, &al.UserEmail, &al.Action, &al.Entity, &entityID, &al.IPAddress, &al.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		if userID.Valid {
			al.UserID = &userID.String
		}
		if entityID.Valid {
			al.EntityID = &entityID.String
		}
		logs = append(logs, al)
	}

	var total int
	r.db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&total)
	return logs, total, nil
}
