package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"crm/config"
	"crm/repositories"
	"crm/utils"
)

// PaymentHandler manages Razorpay payment endpoints
type PaymentHandler struct {
	Cfg     *config.Config
	RegRepo *repositories.RegistrationRepository
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(cfg *config.Config, regRepo *repositories.RegistrationRepository) *PaymentHandler {
	return &PaymentHandler{Cfg: cfg, RegRepo: regRepo}
}

// CreateOrderRequest is the frontend request to initiate payment
type CreateOrderRequest struct {
	RegistrationID string `json:"registration_id"`
}

// VerifyPaymentRequest is the frontend request after Razorpay checkout completion
type VerifyPaymentRequest struct {
	RegistrationID    string `json:"registration_id"`
	RazorpayOrderID   string `json:"razorpay_order_id"`
	RazorpayPaymentID string `json:"razorpay_payment_id"`
	RazorpaySignature string `json:"razorpay_signature"`
}

// CreateOrder handles POST /api/payment/create-order
// Creates a Razorpay order and returns order_id + key for frontend checkout
func (h *PaymentHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.RegistrationID == "" {
		utils.BadRequest(w, "Registration ID is required")
		return
	}

	// Verify registration exists
	reg, err := h.RegRepo.FindByID(req.RegistrationID)
	if err != nil || reg == nil {
		utils.NotFound(w, "Registration not found")
		return
	}

	if reg.PaymentStatus == "submitted" {
		utils.BadRequest(w, "Payment has already been completed for this registration")
		return
	}

	// Create Razorpay order for ₹499 (49900 paise)
	receiptID := fmt.Sprintf("reg_%s", req.RegistrationID[:8])
	order, err := utils.CreateRazorpayOrder(h.Cfg, 49900, receiptID)
	if err != nil {
		log.Printf("[PAYMENT] Failed to create Razorpay order: %v", err)
		utils.InternalServerError(w, "Failed to initiate payment. Please try again.")
		return
	}

	log.Printf("[PAYMENT] Razorpay order created: %s for registration %s", order.ID, req.RegistrationID)

	utils.Success(w, map[string]interface{}{
		"order_id":        order.ID,
		"amount":          order.Amount,
		"currency":        order.Currency,
		"razorpay_key_id": h.Cfg.Razorpay.KeyID,
		"registration_id": req.RegistrationID,
		"prefill": map[string]string{
			"name":  reg.Name,
			"email": reg.Email,
		},
	})
}

// VerifyPayment handles POST /api/payment/verify
// Cryptographically verifies Razorpay payment signature and marks registration as paid
func (h *PaymentHandler) VerifyPayment(w http.ResponseWriter, r *http.Request) {
	var req VerifyPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.RegistrationID == "" || req.RazorpayOrderID == "" || req.RazorpayPaymentID == "" || req.RazorpaySignature == "" {
		utils.BadRequest(w, "All payment verification fields are required")
		return
	}

	// Verify registration exists
	reg, err := h.RegRepo.FindByID(req.RegistrationID)
	if err != nil || reg == nil {
		utils.NotFound(w, "Registration not found")
		return
	}

	// Cryptographically verify the Razorpay payment signature
	isValid := utils.VerifyRazorpaySignature(
		req.RazorpayOrderID,
		req.RazorpayPaymentID,
		req.RazorpaySignature,
		h.Cfg.Razorpay.KeySecret,
	)

	if !isValid {
		log.Printf("[PAYMENT] Invalid signature for registration %s. Possible tampering!", req.RegistrationID)
		utils.Unauthorized(w, "Payment verification failed. Invalid signature.")
		return
	}

	// Update payment status with real Razorpay payment ID as transaction ID
	if err := h.RegRepo.SubmitPayment(req.RegistrationID, req.RazorpayPaymentID); err != nil {
		log.Printf("[PAYMENT] Failed to update payment status: %v", err)
		utils.InternalServerError(w, "Payment verified but failed to update status. Contact support.")
		return
	}

	log.Printf("[PAYMENT] Payment verified successfully for registration %s (Payment ID: %s)", req.RegistrationID, req.RazorpayPaymentID)

	utils.Success(w, map[string]string{
		"message":    "Payment verified successfully! Awaiting admin approval.",
		"status":     "submitted",
		"payment_id": req.RazorpayPaymentID,
	})
}

// GetRazorpayKey handles GET /api/payment/key — returns the public key for frontend
func (h *PaymentHandler) GetRazorpayKey(w http.ResponseWriter, r *http.Request) {
	utils.Success(w, map[string]string{
		"key_id": h.Cfg.Razorpay.KeyID,
	})
}
