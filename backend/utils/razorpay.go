package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"crm/config"
)

// RazorpayOrderRequest is the payload sent to Razorpay Create Order API
type RazorpayOrderRequest struct {
	Amount   int    `json:"amount"`   // Amount in paise (49900 = ₹499)
	Currency string `json:"currency"` // INR
	Receipt  string `json:"receipt"`  // Unique receipt ID
}

// RazorpayOrderResponse is Razorpay's response after order creation
type RazorpayOrderResponse struct {
	ID       string `json:"id"`
	Entity   string `json:"entity"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
	Receipt  string `json:"receipt"`
	Status   string `json:"status"`
}

// CreateRazorpayOrder calls Razorpay Orders API to create a new order
func CreateRazorpayOrder(cfg *config.Config, amountPaise int, receiptID string) (*RazorpayOrderResponse, error) {
	if cfg.Razorpay.KeyID == "" || cfg.Razorpay.KeySecret == "" {
		return nil, fmt.Errorf("razorpay credentials not configured")
	}

	orderReq := RazorpayOrderRequest{
		Amount:   amountPaise,
		Currency: "INR",
		Receipt:  receiptID,
	}

	body, err := json.Marshal(orderReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.razorpay.com/v1/orders", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cfg.Razorpay.KeyID, cfg.Razorpay.KeySecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("razorpay API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read razorpay response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("razorpay API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var orderResp RazorpayOrderResponse
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to parse razorpay response: %w", err)
	}

	return &orderResp, nil
}

// VerifyRazorpaySignature verifies the payment signature using HMAC SHA256
func VerifyRazorpaySignature(orderID, paymentID, signature, secret string) bool {
	// Razorpay signature = HMAC-SHA256(order_id + "|" + payment_id, key_secret)
	data := orderID + "|" + paymentID
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
