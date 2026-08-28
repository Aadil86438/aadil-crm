package handlers

import (
	"encoding/json"
	"net/http"

	"crm/database"
	"crm/models"
	"crm/repositories"
	"crm/utils"
)

// AdminPanelHandler handles the admin approval panel (code-gated, not JWT-based)
type AdminPanelHandler struct {
	regRepo  *repositories.RegistrationRepository
	userRepo *repositories.UserRepository
}

// NewAdminPanelHandler creates a new AdminPanelHandler
func NewAdminPanelHandler(regRepo *repositories.RegistrationRepository, userRepo *repositories.UserRepository) *AdminPanelHandler {
	return &AdminPanelHandler{regRepo: regRepo, userRepo: userRepo}
}

// adminCode is the hardcoded admin access code
const adminCode = "1101"

// VerifyCode handles POST /api/admin/verify — validates the admin code
func (h *AdminPanelHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	var req models.AdminVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Code != adminCode {
		utils.Forbidden(w, "Invalid admin code")
		return
	}

	// Generate a short-lived token for admin operations
	token, err := utils.GenerateToken("admin-panel", "admin@proprietor.app", "admin", "Admin Panel", 2*60*60*1e9) // 2 hours
	if err != nil {
		utils.InternalServerError(w, "Failed to generate session")
		return
	}

	utils.Success(w, map[string]string{
		"token":   token,
		"message": "Admin access granted",
	})
}

// ListPending handles GET /api/admin/pending
func (h *AdminPanelHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	regs, err := h.regRepo.ListPending()
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch pending requests")
		return
	}
	if regs == nil {
		regs = []*models.RegistrationRequest{}
	}
	utils.Success(w, regs)
}

// ListAll handles GET /api/admin/all
func (h *AdminPanelHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	regs, err := h.regRepo.ListAll()
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch registration requests")
		return
	}
	if regs == nil {
		regs = []*models.RegistrationRequest{}
	}
	utils.Success(w, regs)
}

// Approve handles POST /api/admin/approve/:id
func (h *AdminPanelHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/admin/approve/")
	if id == "" {
		utils.BadRequest(w, "Registration ID is required")
		return
	}

	// Get registration request
	reg, err := h.regRepo.FindByID(id)
	if err != nil || reg == nil {
		utils.NotFound(w, "Registration request not found")
		return
	}

	if reg.ApprovalStatus != "pending" {
		utils.BadRequest(w, "This request has already been processed")
		return
	}

	// Create user from registration data
	createReq := &models.CreateUserRequest{
		Name:  reg.Name,
		Email: reg.Email,
		Role:  "sales_user",
	}

	_, err = h.userRepo.Create(createReq, reg.PasswordHash)
	if err != nil {
		utils.InternalServerError(w, "Failed to create user account")
		return
	}

	// Mark registration as approved
	h.regRepo.Approve(id)

	utils.Success(w, map[string]string{
		"message": "User approved successfully. They can now sign in.",
	})
}

// Reject handles POST /api/admin/reject/:id
func (h *AdminPanelHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/admin/reject/")
	if id == "" {
		utils.BadRequest(w, "Registration ID is required")
		return
	}

	reg, err := h.regRepo.FindByID(id)
	if err != nil || reg == nil {
		utils.NotFound(w, "Registration request not found")
		return
	}

	if reg.ApprovalStatus != "pending" {
		utils.BadRequest(w, "This request has already been processed")
		return
	}

	h.regRepo.Reject(id)

	utils.Success(w, map[string]string{
		"message": "Registration request rejected.",
	})
}

// RedisKeyItem represents a key-value detail in Redis
type RedisKeyItem struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	TTL   int64       `json:"ttl"`
	Value interface{} `json:"value"`
}

// GetRedisData handles GET /api/admin/redis — retrieves Redis status and stored keys
func (h *AdminPanelHandler) GetRedisData(w http.ResponseWriter, r *http.Request) {
	if database.RedisClient == nil {
		utils.Success(w, map[string]interface{}{
			"connected": false,
			"message":   "Redis client is not connected",
			"keys":      []RedisKeyItem{},
		})
		return
	}

	ctx := r.Context()

	// Ping check
	pingErr := database.RedisClient.Ping(ctx).Err()
	connected := pingErr == nil

	// Get all keys
	keys, err := database.RedisClient.Keys(ctx, "*").Result()
	if err != nil {
		keys = []string{}
	}

	items := make([]RedisKeyItem, 0, len(keys))
	for _, k := range keys {
		kType, _ := database.RedisClient.Type(ctx, k).Result()
		ttlDuration, _ := database.RedisClient.TTL(ctx, k).Result()

		var val interface{}
		switch kType {
		case "string":
			v, err := database.RedisClient.Get(ctx, k).Result()
			if err == nil {
				// try to decode JSON if possible
				var decoded interface{}
				if json.Unmarshal([]byte(v), &decoded) == nil {
					val = decoded
				} else {
					val = v
				}
			}
		case "hash":
			v, err := database.RedisClient.HGetAll(ctx, k).Result()
			if err == nil {
				val = v
			}
		case "list":
			v, err := database.RedisClient.LRange(ctx, k, 0, 50).Result()
			if err == nil {
				val = v
			}
		case "set":
			v, err := database.RedisClient.SMembers(ctx, k).Result()
			if err == nil {
				val = v
			}
		default:
			val = "(binary/other)"
		}

		items = append(items, RedisKeyItem{
			Key:   k,
			Type:  kType,
			TTL:   int64(ttlDuration.Seconds()),
			Value: val,
		})
	}

	// Get basic Redis info
	infoRaw, _ := database.RedisClient.Info(ctx, "server", "memory", "clients").Result()

	utils.Success(w, map[string]interface{}{
		"connected":  connected,
		"key_count":  len(keys),
		"keys":       items,
		"raw_info":   infoRaw,
	})
}

// DeleteRedisKey handles DELETE /api/admin/redis?key=xxx — deletes a key from Redis
func (h *AdminPanelHandler) DeleteRedisKey(w http.ResponseWriter, r *http.Request) {
	if database.RedisClient == nil {
		utils.BadRequest(w, "Redis is not connected")
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		utils.BadRequest(w, "Key query parameter is required")
		return
	}

	ctx := r.Context()
	err := database.RedisClient.Del(ctx, key).Err()
	if err != nil {
		utils.InternalServerError(w, "Failed to delete key: "+err.Error())
		return
	}

	utils.Success(w, map[string]string{
		"message": "Key deleted successfully",
		"key":     key,
	})
}

