package handlers

import (
	"encoding/json"
	"net/http"

	"admin-service-gcp/database"
	"admin-service-gcp/models"
	"admin-service-gcp/repositories"
	"admin-service-gcp/utils"
)

// AdminPanelHandler handles the admin approval & devops panel
type AdminPanelHandler struct {
	userRepo *repositories.UserRepository
}

// NewAdminPanelHandler creates a new AdminPanelHandler
func NewAdminPanelHandler(userRepo *repositories.UserRepository) *AdminPanelHandler {
	return &AdminPanelHandler{userRepo: userRepo}
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

// ListUsers handles GET /api/admin/all — lists registered system users
func (h *AdminPanelHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	users, _, err := h.userRepo.ListUsers(1, 1000)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch users")
		return
	}
	if users == nil {
		users = []*models.User{}
	}
	utils.Success(w, users)
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


