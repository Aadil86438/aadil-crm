// Package coreclient handles all HTTP communication FROM admin-service-gcp TO core-service-aws.
// This is the inter-service communication layer — the KEY learning component of this microservices setup.
//
// How it works:
//   admin-service-gcp (GCP) ──── HTTPS + X-Internal-Service-Key header ────► core-service-aws (AWS)
//
// The X-Internal-Service-Key is a shared secret (like a password between the two services).
// It must be set identically in both services' .env files.
package coreclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CoreClient makes authenticated HTTP calls to core-service-aws
type CoreClient struct {
	baseURL     string
	serviceKey  string
	httpClient  *http.Client
}

// New creates a new CoreClient
// baseURL: the AWS core-service URL, e.g., "https://api.yourdomain.com"
// serviceKey: the shared INTERNAL_SERVICE_KEY secret
func New(baseURL, serviceKey string) *CoreClient {
	return &CoreClient{
		baseURL:    baseURL,
		serviceKey: serviceKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// doRequest makes an authenticated GET request to core-service-aws
func (c *CoreClient) doRequest(path string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// ← THIS IS HOW MICROSERVICES AUTHENTICATE WITH EACH OTHER
	// We pass the shared secret in a custom header instead of a user JWT
	req.Header.Set("X-Internal-Service-Key", c.serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to core-service-aws failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("inter-service auth failed: check INTERNAL_SERVICE_KEY in both services")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("core-service-aws returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// GetStats fetches dashboard stats from core-service-aws
func (c *CoreClient) GetStats() (map[string]interface{}, error) {
	body, err := c.doRequest("/internal/stats")
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetAuditLogs fetches audit logs from core-service-aws
func (c *CoreClient) GetAuditLogs(page, pageSize int) (map[string]interface{}, error) {
	path := fmt.Sprintf("/internal/audit-logs?page=%d&page_size=%d", page, pageSize)
	body, err := c.doRequest(path)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetUsers fetches the user list from core-service-aws
func (c *CoreClient) GetUsers() (map[string]interface{}, error) {
	body, err := c.doRequest("/internal/users")
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
