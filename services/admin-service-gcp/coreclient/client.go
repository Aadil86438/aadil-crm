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
	"strings"
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

// BaseURL returns the configured core service base URL
func (c *CoreClient) BaseURL() string {
	return c.baseURL
}

// ServiceKey returns the inter-service auth key
func (c *CoreClient) ServiceKey() string {
	return c.serviceKey
}

// doRequest makes an authenticated GET request to core-service-aws (with fallback between :8080 and :80)
func (c *CoreClient) doRequest(path string) ([]byte, error) {
	targetURLs := []string{fmt.Sprintf("%s%s", c.baseURL, path)}
	if strings.Contains(c.baseURL, ":8080") {
		targetURLs = append(targetURLs, fmt.Sprintf("%s%s", strings.Replace(c.baseURL, ":8080", "", 1), path))
	}

	var lastErr error
	for _, targetURL := range targetURLs {
		req, err := http.NewRequest(http.MethodGet, targetURL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("X-Internal-Service-Key", c.serviceKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request to %s failed: %w", targetURL, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("inter-service auth failed: check INTERNAL_SERVICE_KEY in both services")
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("core-service-aws (%s) returned status %d", targetURL, resp.StatusCode)
			continue
		}

		return io.ReadAll(resp.Body)
	}

	return nil, lastErr
}

// DoRequest performs an authenticated GET request and returns raw response bytes
func (c *CoreClient) DoRequest(path string) ([]byte, error) {
	return c.doRequest(path)
}

// DoPostRequest performs an authenticated POST request to core-service-aws (with fallback)
func (c *CoreClient) DoPostRequest(path string, body io.Reader) ([]byte, error) {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = io.ReadAll(body)
	}

	targetURLs := []string{fmt.Sprintf("%s%s", c.baseURL, path)}
	if strings.Contains(c.baseURL, ":8080") {
		targetURLs = append(targetURLs, fmt.Sprintf("%s%s", strings.Replace(c.baseURL, ":8080", "", 1), path))
	}

	var lastErr error
	for _, targetURL := range targetURLs {
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = strings.NewReader(string(bodyBytes))
		}

		req, err := http.NewRequest(http.MethodPost, targetURL, bodyReader)
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("X-Internal-Service-Key", c.serviceKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request to %s failed: %w", targetURL, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("inter-service auth failed: check INTERNAL_SERVICE_KEY in both services")
		}

		data, err := io.ReadAll(resp.Body)
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("core-service-aws (%s) returned status %d: %s", targetURL, resp.StatusCode, string(data))
			continue
		}

		return data, err
	}

	return nil, lastErr
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
