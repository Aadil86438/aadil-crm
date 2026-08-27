package utils

import (
	"net/http"
	"strconv"
	"strings"
)

// ParsePagination extracts pagination params from query string
func ParsePagination(r *http.Request) (page, pageSize int) {
	page = 1
	pageSize = 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	return
}

// ParseSort extracts sort params from query string
// Returns column and direction (ASC/DESC)
func ParseSort(r *http.Request, allowed []string) (column, direction string) {
	column = "created_at"
	direction = "DESC"

	if s := r.URL.Query().Get("sort"); s != "" {
		for _, a := range allowed {
			if strings.EqualFold(s, a) {
				column = a
				break
			}
		}
	}
	if d := r.URL.Query().Get("order"); d != "" {
		if strings.EqualFold(d, "asc") {
			direction = "ASC"
		} else {
			direction = "DESC"
		}
	}
	return
}

// Offset calculates the SQL OFFSET value
func Offset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// IsValidEmail does a basic email validation
func IsValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	validRoles := []string{"admin", "manager", "sales_user"}
	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}
