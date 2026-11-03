package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"core-service-aws/repositories"
	"core-service-aws/utils"
)

// ─── AIHandler handles all AI/vector-search endpoints on core-service-aws ───
// Routes (all protected by InternalServiceAuth middleware):
//   POST /internal/ai/search  — vector similarity search across CRM data
//   POST /internal/ai/embed   — store a new or updated embedding
//   GET  /internal/ai/status  — embedding index health check
// ─────────────────────────────────────────────────────────────────────────────

type AIHandler struct {
	aiRepo *repositories.AIRepository
}

func NewAIHandler(aiRepo *repositories.AIRepository) *AIHandler {
	return &AIHandler{aiRepo: aiRepo}
}

// ─── Request / Response types ────────────────────────────────────────────────

type EmbedRequest struct {
	SourceType string                 `json:"source_type"` // "user", "deal", "audit_log", etc.
	SourceID   string                 `json:"source_id"`
	Content    string                 `json:"content"`    // plain-text representation of the record
	Metadata   map[string]interface{} `json:"metadata"`   // extra filter context
}

type SearchRequest struct {
	Query       string   `json:"query"`        // natural language question
	SourceTypes []string `json:"source_types"` // optional filter — nil = all
	TopK        int      `json:"top_k"`        // default 5
}

type SearchResponse struct {
	Answer  string                       `json:"answer"`   // synthesized by Gemini (set by caller)
	Chunks  []*repositories.EmbeddingRecord `json:"chunks"`   // raw retrieved chunks
	Latency int64                        `json:"latency_ms"`
}

// ─── Embed — POST /internal/ai/embed ────────────────────────────────────────
// Called by core-service-aws itself whenever a record is created/updated.
// Gets the Gemini embedding for `content` and stores it in pgvector.

func (h *AIHandler) Embed(w http.ResponseWriter, r *http.Request) {
	var req EmbedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	if req.SourceType == "" || req.SourceID == "" || req.Content == "" {
		utils.BadRequest(w, "source_type, source_id and content are required")
		return
	}

	embedding, err := geminiEmbed(req.Content)
	if err != nil {
		utils.InternalServerError(w, "Failed to generate embedding: "+err.Error())
		return
	}

	if err := h.aiRepo.UpsertEmbedding(req.SourceType, req.SourceID, req.Content, embedding, req.Metadata); err != nil {
		utils.InternalServerError(w, "Failed to store embedding: "+err.Error())
		return
	}

	utils.Success(w, map[string]interface{}{
		"message":     "embedding stored",
		"source_type": req.SourceType,
		"source_id":   req.SourceID,
		"dims":        len(embedding),
	})
}

// ─── Search — POST /internal/ai/search ──────────────────────────────────────
// Called by admin-service-gcp's AI chat handler.
// Embeds the query, runs pgvector similarity search, returns top chunks.

func (h *AIHandler) Search(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	if strings.TrimSpace(req.Query) == "" {
		utils.BadRequest(w, "query is required")
		return
	}
	if req.TopK == 0 {
		req.TopK = 5
	}

	// 1. Embed the query using Gemini
	queryVec, err := geminiEmbed(req.Query)
	if err != nil {
		utils.InternalServerError(w, "Failed to embed query: "+err.Error())
		return
	}

	// 2. pgvector cosine similarity search
	chunks, err := h.aiRepo.SearchSimilar(queryVec, req.SourceTypes, req.TopK)
	if err != nil {
		utils.InternalServerError(w, "Vector search failed: "+err.Error())
		return
	}

	utils.Success(w, SearchResponse{
		Chunks:  chunks,
		Latency: time.Since(start).Milliseconds(),
	})
}

// ─── Status — GET /internal/ai/status ───────────────────────────────────────

func (h *AIHandler) Status(w http.ResponseWriter, r *http.Request) {
	counts, err := h.aiRepo.CountEmbeddings()
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch embedding stats: "+err.Error())
		return
	}
	total := 0
	for _, v := range counts {
		total += v
	}
	utils.Success(w, map[string]interface{}{
		"status":      "ok",
		"total":       total,
		"by_type":     counts,
		"vector_dims": 768,
		"model":       "gemini-text-embedding-004",
	})
}

// ─── Sync — POST /internal/ai/sync ─────────────────────────────────────────
// Scans CRM records (users, audit logs), computes Gemini embeddings, and saves to pgvector.

func (h *AIHandler) Sync(w http.ResponseWriter, r *http.Request) {
	items, err := h.aiRepo.GetSyncItems()
	if err != nil {
		utils.InternalServerError(w, "Failed to load CRM entities: "+err.Error())
		return
	}

	synced := 0
	failed := 0
	for _, item := range items {
		vec, err := geminiEmbed(item.Content)
		if err != nil {
			failed++
			continue
		}
		if err := h.aiRepo.UpsertEmbedding(item.SourceType, item.SourceID, item.Content, vec, item.Metadata); err != nil {
			failed++
			continue
		}
		synced++
	}

	counts, _ := h.aiRepo.CountEmbeddings()
	utils.Success(w, map[string]interface{}{
		"message": "Knowledge base sync complete",
		"synced":  synced,
		"failed":  failed,
		"total":   len(items),
		"counts":  counts,
	})
}

// ─── Gemini Embedding API (free tier) ────────────────────────────────────────
// Calls the Gemini text-embedding-004 REST API directly (no SDK needed).
// Returns a 768-dimensional float32 vector.

type geminiEmbedRequest struct {
	Model                string             `json:"model"`
	Content              geminiEmbedContent `json:"content"`
	OutputDimensionality int                `json:"outputDimensionality,omitempty"`
}

type geminiEmbedContent struct {
	Parts []geminiEmbedPart `json:"parts"`
}

type geminiEmbedPart struct {
	Text string `json:"text"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

func geminiEmbed(text string) ([]float32, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}

	if len(text) > 8000 {
		text = text[:8000]
	}

	// Try gemini-embedding-001 with 768 dims, then text-embedding-004
	models := []struct {
		name string
		dims int
	}{
		{"gemini-embedding-001", 768},
		{"text-embedding-004", 0},
	}

	var lastErr error
	for _, m := range models {
		payload := geminiEmbedRequest{
			Model: fmt.Sprintf("models/%s", m.name),
			Content: geminiEmbedContent{
				Parts: []geminiEmbedPart{{Text: text}},
			},
			OutputDimensionality: m.dims,
		}

		body, _ := json.Marshal(payload)
		url := fmt.Sprintf(
			"https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
			m.name,
			apiKey,
		)

		resp, err := http.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = err
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("gemini embed (%s) error %d: %s", m.name, resp.StatusCode, string(respBody))
			continue
		}

		var result geminiEmbedResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			lastErr = fmt.Errorf("failed to parse embed response: %w", err)
			continue
		}

		if len(result.Embedding.Values) > 0 {
			return result.Embedding.Values, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("failed to generate embedding")
}
