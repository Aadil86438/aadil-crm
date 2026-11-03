package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// AIRepository handles all pgvector / AI embedding database operations on AWS PostgreSQL
type AIRepository struct {
	db *sql.DB
}

// NewAIRepository creates a new AIRepository
func NewAIRepository(db *sql.DB) *AIRepository {
	return &AIRepository{db: db}
}

// EmbeddingRecord holds a stored embedding and its metadata
type EmbeddingRecord struct {
	ID         string                 `json:"id"`
	SourceType string                 `json:"source_type"`
	SourceID   string                 `json:"source_id"`
	Content    string                 `json:"content"`
	Metadata   map[string]interface{} `json:"metadata"`
	Score      float64                `json:"score"` // cosine similarity score (1 = identical)
}

// UpsertEmbedding stores or updates an embedding for a CRM record.
// embedding is a slice of float32 from Gemini text-embedding-004 (768 dims).
func (r *AIRepository) UpsertEmbedding(sourceType, sourceID, content string, embedding []float32, metadata map[string]interface{}) error {
	// Convert embedding to PostgreSQL vector literal: '[0.1,0.2,...]'
	vecStr := floatsToVectorLiteral(embedding)

	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	_, err = r.db.Exec(`
		INSERT INTO ai_embeddings (source_type, source_id, content, embedding, metadata, updated_at)
		VALUES ($1, $2, $3, $4::vector, $5, NOW())
		ON CONFLICT (source_type, source_id)
		DO UPDATE SET
			content   = EXCLUDED.content,
			embedding = EXCLUDED.embedding,
			metadata  = EXCLUDED.metadata,
			updated_at = NOW()
	`, sourceType, sourceID, content, vecStr, metaJSON)

	return err
}

// SearchSimilar performs a cosine-similarity vector search.
// queryEmbedding: vector from embedding the user's question (768 dims)
// sourceTypes: optional filter e.g. []string{"user","audit_log"} — nil = search all
// topK: number of results to return
func (r *AIRepository) SearchSimilar(queryEmbedding []float32, sourceTypes []string, topK int) ([]*EmbeddingRecord, error) {
	vecStr := floatsToVectorLiteral(queryEmbedding)

	if topK <= 0 {
		topK = 5
	}
	if topK > 20 {
		topK = 20
	}

	var query string
	var args []interface{}

	if len(sourceTypes) > 0 {
		// Build parameterised IN clause
		placeholders := make([]string, len(sourceTypes))
		args = append(args, vecStr)
		for i, t := range sourceTypes {
			args = append(args, t)
			placeholders[i] = fmt.Sprintf("$%d", i+2)
		}
		args = append(args, topK)

		query = fmt.Sprintf(`
			SELECT id, source_type, source_id, content, metadata,
			       1 - (embedding <=> $1::vector) AS score
			FROM ai_embeddings
			WHERE source_type IN (%s)
			ORDER BY embedding <=> $1::vector
			LIMIT $%d
		`, strings.Join(placeholders, ","), len(args))
	} else {
		args = []interface{}{vecStr, topK}
		query = `
			SELECT id, source_type, source_id, content, metadata,
			       1 - (embedding <=> $1::vector) AS score
			FROM ai_embeddings
			ORDER BY embedding <=> $1::vector
			LIMIT $2
		`
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	defer rows.Close()

	var results []*EmbeddingRecord
	for rows.Next() {
		rec := &EmbeddingRecord{}
		var metaJSON []byte
		if err := rows.Scan(&rec.ID, &rec.SourceType, &rec.SourceID, &rec.Content, &metaJSON, &rec.Score); err != nil {
			continue
		}
		if metaJSON != nil {
			json.Unmarshal(metaJSON, &rec.Metadata)
		}
		results = append(results, rec)
	}

	return results, rows.Err()
}

// DeleteEmbedding removes the embedding for a specific record (call on delete)
func (r *AIRepository) DeleteEmbedding(sourceType, sourceID string) error {
	_, err := r.db.Exec(
		`DELETE FROM ai_embeddings WHERE source_type = $1 AND source_id = $2`,
		sourceType, sourceID,
	)
	return err
}

// CountEmbeddings returns how many embeddings are stored per source type
func (r *AIRepository) CountEmbeddings() (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT source_type, COUNT(*) FROM ai_embeddings GROUP BY source_type ORDER BY source_type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var sType string
		var count int
		if err := rows.Scan(&sType, &count); err == nil {
			counts[sType] = count
		}
	}
	return counts, rows.Err()
}

// SyncItem represents a chunk of CRM data ready for embedding
type SyncItem struct {
	SourceType string                 `json:"source_type"`
	SourceID   string                 `json:"source_id"`
	Content    string                 `json:"content"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// GetSyncItems fetches recent CRM records (users, audit logs, tasks, deals) to vectorize
func (r *AIRepository) GetSyncItems() ([]SyncItem, error) {
	var items []SyncItem

	// 1. Users
	uRows, err := r.db.Query(`
		SELECT id, name, email, role, status, COALESCE(approval_status, 'pending'), COALESCE(company_name, ''), COALESCE(crm_code, '')
		FROM users WHERE deleted_at IS NULL LIMIT 200
	`)
	if err == nil {
		defer uRows.Close()
		for uRows.Next() {
			var id, name, email, role, status, approvalStatus, company, crmCode string
			if err := uRows.Scan(&id, &name, &email, &role, &status, &approvalStatus, &company, &crmCode); err == nil {
				content := fmt.Sprintf("CRM User: %s (%s). Role: %s. Account Status: %s. Registration Approval: %s. Company: %s. CRM Code: %s.",
					name, email, role, status, approvalStatus, company, crmCode)
				items = append(items, SyncItem{
					SourceType: "user",
					SourceID:   id,
					Content:    content,
					Metadata: map[string]interface{}{
						"name":            name,
						"email":           email,
						"role":            role,
						"status":          status,
						"approval_status": approvalStatus,
						"company":         company,
						"crm_code":        crmCode,
					},
				})
			}
		}
	}

	// 2. Audit Logs
	aRows, err := r.db.Query(`
		SELECT id, COALESCE(user_email, 'system'), action, entity, COALESCE(ip_address, ''), created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT 100
	`)
	if err == nil {
		defer aRows.Close()
		for aRows.Next() {
			var id, email, action, entity, ip, createdAt string
			if err := aRows.Scan(&id, &email, &action, &entity, &ip, &createdAt); err == nil {
				content := fmt.Sprintf("Audit Activity Log: User %s performed action '%s' on %s at %s. Client IP: %s.",
					email, action, entity, createdAt, ip)
				items = append(items, SyncItem{
					SourceType: "audit_log",
					SourceID:   id,
					Content:    content,
					Metadata: map[string]interface{}{
						"user_email": email,
						"action":     action,
						"entity":     entity,
						"created_at": createdAt,
					},
				})
			}
		}
	}

	return items, nil
}

// ─── Helper ─────────────────────────────────────────────────────────────────

// floatsToVectorLiteral converts []float32 to PostgreSQL vector literal string e.g. "[0.1,0.2,0.3]"
func floatsToVectorLiteral(v []float32) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%.8f", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
