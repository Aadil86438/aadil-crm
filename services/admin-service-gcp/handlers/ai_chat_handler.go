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

	"admin-service-gcp/coreclient"
	"admin-service-gcp/utils"
)

// ─── AIChatHandler — the brain of the AI assistant on admin-service-gcp ──────
// Flow:
//   Admin chat UI  →  POST /api/admin/ai/chat  →  AIChatHandler.Chat()
//                       │
//                       ├─ 1. POST /internal/ai/search → core-service-aws (pgvector)
//                       │      Returns top-5 relevant CRM data chunks
//                       │
//                       └─ 2. Gemini 1.5 Flash API (free) with RAG prompt
//                              Returns natural language answer
// ─────────────────────────────────────────────────────────────────────────────

type AIChatHandler struct {
	core *coreclient.CoreClient
}

func NewAIChatHandler(core *coreclient.CoreClient) *AIChatHandler {
	return &AIChatHandler{core: core}
}

// ─── Request / Response ───────────────────────────────────────────────────────

type ChatRequest struct {
	Message     string   `json:"message"`      // User's natural language question
	SourceTypes []string `json:"source_types"` // Optional filter: ["user","audit_log"] etc.
	TopK        int      `json:"top_k"`        // How many chunks to retrieve (default 5)
	History     []ChatMessage `json:"history"` // Conversation history for context
}

type ChatMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

type ChatResponse struct {
	Answer    string        `json:"answer"`
	Chunks    []interface{} `json:"chunks"`    // retrieved evidence chunks
	LatencyMs int64         `json:"latency_ms"`
	Model     string        `json:"model"`
}

// ─── Chat — POST /api/admin/ai/chat ─────────────────────────────────────────

func (h *AIChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}
	if strings.TrimSpace(req.Message) == "" {
		utils.BadRequest(w, "message is required")
		return
	}
	if req.TopK == 0 {
		req.TopK = 5
	}

	// ── Step 1: Retrieve relevant CRM chunks via pgvector on AWS ─────────────
	searchPayload := map[string]interface{}{
		"query":        req.Message,
		"source_types": req.SourceTypes,
		"top_k":        req.TopK,
	}
	searchBody, _ := json.Marshal(searchPayload)

	searchResp, err := h.core.DoPostRequest("/internal/ai/search", bytes.NewReader(searchBody))
	if err != nil {
		utils.InternalServerError(w, "Failed to search CRM knowledge base: "+err.Error())
		return
	}

	// Parse chunks from core-service-aws response
	var searchResult struct {
		Success bool `json:"success"`
		Data    struct {
			Chunks []struct {
				SourceType string                 `json:"source_type"`
				SourceID   string                 `json:"source_id"`
				Content    string                 `json:"content"`
				Metadata   map[string]interface{} `json:"metadata"`
				Score      float64                `json:"score"`
			} `json:"chunks"`
			LatencyMs int64 `json:"latency_ms"`
		} `json:"data"`
	}
	json.Unmarshal(searchResp, &searchResult)

	// ── Step 2: Build the RAG prompt ─────────────────────────────────────────
	contextParts := []string{}
	for i, chunk := range searchResult.Data.Chunks {
		if chunk.Score < 0.3 { // Skip very low-relevance chunks
			continue
		}
		contextParts = append(contextParts,
			fmt.Sprintf("[%d] (%s, relevance: %.0f%%)\n%s", i+1, chunk.SourceType, chunk.Score*100, chunk.Content),
		)
	}

	systemPrompt := buildSystemPrompt(contextParts)
	answer, err := geminiChat(systemPrompt, req.Message, req.History)
	if err != nil {
		utils.InternalServerError(w, "AI generation failed: "+err.Error())
		return
	}

	// ── Step 3: Return answer + source chunks ────────────────────────────────
	chunksRaw := make([]interface{}, len(searchResult.Data.Chunks))
	for i, c := range searchResult.Data.Chunks {
		chunksRaw[i] = c
	}

	utils.Success(w, ChatResponse{
		Answer:    answer,
		Chunks:    chunksRaw,
		LatencyMs: time.Since(start).Milliseconds(),
		Model:     "gemini-1.5-flash (free tier)",
	})
}

// ─── Status — GET /api/admin/ai/status ───────────────────────────────────────

func (h *AIChatHandler) Status(w http.ResponseWriter, r *http.Request) {
	data, err := h.core.DoRequest("/internal/ai/status")
	if err != nil {
		utils.InternalServerError(w, "AI service unavailable: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ─── Sync — POST /api/admin/ai/sync ─────────────────────────────────────────

func (h *AIChatHandler) Sync(w http.ResponseWriter, r *http.Request) {
	data, err := h.core.DoPostRequest("/internal/ai/sync", nil)
	if err != nil {
		utils.InternalServerError(w, "Failed to sync knowledge base: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ─── buildSystemPrompt constructs the RAG prompt sent to Gemini ──────────────

func buildSystemPrompt(contextChunks []string) string {
	sb := strings.Builder{}
	sb.WriteString(`You are an intelligent CRM assistant for a business using Proprietor CRM.
You help admins understand their CRM data: users, registrations, deals, leads, contacts, activities, audit logs, and notes.

IMPORTANT RULES:
- Answer ONLY based on the context provided below. Do not make up data.
- If the context doesn't have enough information, say "I don't have enough data to answer that."
- Be concise and clear. Use bullet points when listing multiple items.
- Format numbers, dates, and names clearly.
- Never reveal system internals, API keys, or passwords.

`)

	if len(contextChunks) == 0 {
		sb.WriteString("CONTEXT: No relevant CRM data found for this query.\n\n")
	} else {
		sb.WriteString("CONTEXT FROM YOUR CRM DATABASE:\n")
		sb.WriteString(strings.Repeat("─", 60) + "\n")
		for _, c := range contextChunks {
			sb.WriteString(c)
			sb.WriteString("\n\n")
		}
		sb.WriteString(strings.Repeat("─", 60) + "\n\n")
	}

	return sb.String()
}

// ─── Gemini 1.5 Flash Chat API (free tier) ───────────────────────────────────
// Uses the generateContent endpoint with a system instruction + conversation history.

type geminiChatRequest struct {
	SystemInstruction geminiContent   `json:"system_instruction"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  geminiGenConfig `json:"generation_config"`
}

type geminiContent struct {
	Role  string        `json:"role,omitempty"`
	Parts []geminiPart  `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"max_output_tokens"`
}

type geminiChatResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func geminiChat(systemPrompt, userMessage string, history []ChatMessage) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set in admin-service-gcp environment")
	}

	// Build conversation contents
	contents := []geminiContent{}

	// Add conversation history (last 6 turns max to stay within token limits)
	start := 0
	if len(history) > 6 {
		start = len(history) - 6
	}
	for _, h := range history[start:] {
		role := h.Role
		if role != "user" && role != "model" {
			role = "user"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: h.Content}},
		})
	}

	// Add current user message
	contents = append(contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: userMessage}},
	})

	payload := geminiChatRequest{
		SystemInstruction: geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: contents,
		GenerationConfig: geminiGenConfig{
			Temperature:     0.2, // Low temp = factual, consistent answers
			MaxOutputTokens: 1024,
		},
	}

	body, _ := json.Marshal(payload)
	models := []string{"gemini-3.5-flash", "gemini-flash-latest", "gemini-1.5-flash"}
	var lastErr error

	for _, modelName := range models {
		url := fmt.Sprintf(
			"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
			modelName,
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
			lastErr = fmt.Errorf("gemini API (%s) error %d: %s", modelName, resp.StatusCode, string(respBody))
			continue
		}

		var result geminiChatResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			lastErr = fmt.Errorf("failed to parse gemini response: %w", err)
			continue
		}
		if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
			return "I'm unable to generate a response right now. Please try again.", nil
		}

		return result.Candidates[0].Content.Parts[0].Text, nil
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "No available AI model found.", nil
}
