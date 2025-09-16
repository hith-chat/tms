package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

type StripeWebhookHandler struct {
	WebhookRepo   PaymentWebhookRepository
	CreditsRepo   repo.CreditsRepository
	TenantRepo    repo.TenantRepository
	WebhookSecret string
}

// PaymentWebhookRepository interface (simplified for this implementation)
type PaymentWebhookRepository interface {
	Create(event *db.PaymentWebhookEvent) error
	FindByPaymentEventID(eventID string) (*db.PaymentWebhookEvent, error)
	Update(event *db.PaymentWebhookEvent) error
	FindTenantByEmail(email string) (*db.Tenant, error)
}

func NewStripeWebhookHandler(webhookRepo PaymentWebhookRepository, creditsRepo repo.CreditsRepository, tenantRepo repo.TenantRepository, webhookSecret string) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		WebhookRepo:   webhookRepo,
		CreditsRepo:   creditsRepo,
		TenantRepo:    tenantRepo,
		WebhookSecret: webhookSecret,
	}
}

// StripeWebhookEvent represents the structure of a Stripe webhook event
type StripeWebhookEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object map[string]interface{} `json:"object"`
	} `json:"data"`
}

// HandleStripeWebhook processes incoming Stripe webhook events
func (h *StripeWebhookHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading Stripe webhook body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Verify the webhook signature
	signature := r.Header.Get("Stripe-Signature")
	if !h.verifySignature(body, signature) {
		log.Printf("Invalid Stripe webhook signature")
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	// Parse the webhook event
	var event StripeWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing Stripe webhook JSON: %v", err)
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Check if we've already processed this event (idempotency)
	existingEvent, err := h.WebhookRepo.FindByPaymentEventID(event.ID)
	if err != nil {
		log.Printf("Error checking for existing Stripe event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingEvent != nil {
		// Event already processed, return success to avoid retries
		log.Printf("Stripe event %s already processed, skipping", event.ID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Event already processed",
		})
		return
	}

	// Create webhook event record
	webhookEvent := &db.PaymentWebhookEvent{
		EventID:        event.ID,
		EventType:      event.Type,
		PayloadData:    json.RawMessage(body),
		Status:         "pending",
		PaymentGateway: "stripe",
	}

	// Extract common fields from the event data
	if objectID, ok := event.Data.Object["id"].(string); ok {
		webhookEvent.ObjectID = objectID
	}

	if currency, ok := event.Data.Object["currency"].(string); ok {
		webhookEvent.Currency = currency
	}

	// Process different event types
	switch event.Type {
	case "checkout.session.completed":
		h.processCheckoutSessionCompleted(webhookEvent, event)
	case "payment_intent.succeeded":
		h.processPaymentIntentSucceeded(webhookEvent, event)
	case "charge.succeeded":
		h.processChargeSucceeded(webhookEvent, event)
	default:
		// Mark as ignored for unhandled event types
		webhookEvent.MarkIgnored()
		log.Printf("Ignoring unhandled Stripe event type: %s", event.Type)
	}

	// Save the webhook event to database
	if err := h.WebhookRepo.Create(webhookEvent); err != nil {
		log.Printf("Error saving Stripe webhook event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update the event if it was processed
	if webhookEvent.Status != "pending" {
		if err := h.WebhookRepo.Update(webhookEvent); err != nil {
			log.Printf("Error updating Stripe webhook event: %v", err)
		}
	}

	// Return success immediately to Stripe
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Stripe webhook processed",
	})
}

// verifySignature verifies the Stripe webhook signature
func (h *StripeWebhookHandler) verifySignature(body []byte, signature string) bool {
	if h.WebhookSecret == "" {
		log.Printf("Warning: No Stripe webhook secret configured, skipping signature verification")
		return true // In development, you might want to skip verification
	}

	// Parse the signature header
	parts := strings.Split(signature, ",")
	var timestamp string
	var sigs []string

	for _, part := range parts {
		if strings.HasPrefix(part, "t=") {
			timestamp = strings.TrimPrefix(part, "t=")
		} else if strings.HasPrefix(part, "v1=") {
			sigs = append(sigs, strings.TrimPrefix(part, "v1="))
		}
	}

	if timestamp == "" || len(sigs) == 0 {
		return false
	}

	// Verify timestamp (should be within 5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix()-ts > 300 { // 5 minutes
		return false
	}

	// Create the signed payload
	signedPayload := timestamp + "." + string(body)

	// Compute the HMAC
	mac := hmac.New(sha256.New, []byte(h.WebhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Compare with provided signatures
	for _, sig := range sigs {
		if hmac.Equal([]byte(expectedSig), []byte(sig)) {
			return true
		}
	}

	return false
}

// processCheckoutSessionCompleted handles successful async payments from checkout sessions
func (h *StripeWebhookHandler) processCheckoutSessionCompleted(webhookEvent *db.PaymentWebhookEvent, event StripeWebhookEvent) {
	log.Printf("Processing checkout.session.completed for event %s", event.ID)

	// Extract customer email
	var tenantEmail string
	if customerDetails, ok := event.Data.Object["customer_details"].(map[string]interface{}); ok {
		if customerEmail, ok := customerDetails["email"].(string); ok {
			tenantEmail = customerEmail
		}
	}

	// Extract amount
	if amountTotal, ok := event.Data.Object["amount_total"].(float64); ok {
		amountInt := int64(amountTotal)
		webhookEvent.Amount = &amountInt
	}

	if tenantEmail != "" {
		webhookEvent.TenantEmail = tenantEmail

		// Find tenant by email
		tenant, err := h.WebhookRepo.FindTenantByEmail(tenantEmail)
		if err != nil {
			errorMsg := fmt.Sprintf("Error finding tenant by email %s: %v", tenantEmail, err)
			log.Printf("Error: %s", errorMsg)
			webhookEvent.MarkFailed(errorMsg)
			return
		}

		if tenant != nil {
			webhookEvent.TenantID = &tenant.ID

			if webhookEvent.Amount == nil || *webhookEvent.Amount <= 0 {
				webhookEvent.MarkFailed("Invalid amount in payment")
				return
			}

			// Calculate credits from payment (1:1 ratio for now)
			creditAmount := models.CalculateCreditsFromPayment(*webhookEvent.Amount)

			// Add credits to tenant account
			_, err := h.CreditsRepo.AddCredits(
				context.Background(),
				tenant.ID,
				creditAmount,
				models.TransactionTypePayment,
				models.PaymentGatewayStripe,
				event.ID,
				fmt.Sprintf("Stripe payment via checkout session: %s", webhookEvent.ObjectID),
			)
			if err != nil {
				errorMsg := fmt.Sprintf("Error adding credits to tenant %s: %v", tenant.ID, err)
				log.Printf("Error: %s", errorMsg)
				webhookEvent.MarkFailed(errorMsg)
				return
			}

			log.Printf("Successfully added %d credits to tenant %s (%s) via Stripe payment", creditAmount, tenant.ID, tenantEmail)
			webhookEvent.MarkProcessed()
		} else {
			errorMsg := fmt.Sprintf("Tenant not found with email: %s", tenantEmail)
			log.Printf("Error: %s", errorMsg)
			webhookEvent.MarkFailed(errorMsg)
		}
	} else {
		errorMsg := "No customer email found in Stripe checkout session"
		log.Printf("Error: %s", errorMsg)
		webhookEvent.MarkFailed(errorMsg)
	}
}

// processPaymentIntentSucceeded handles successful payment intents
func (h *StripeWebhookHandler) processPaymentIntentSucceeded(webhookEvent *db.PaymentWebhookEvent, event StripeWebhookEvent) {
	log.Printf("Processing payment_intent.succeeded for event %s", event.ID)

	// Extract amount
	if amount, ok := event.Data.Object["amount"].(float64); ok {
		amountInt := int64(amount)
		webhookEvent.Amount = &amountInt
	}

	// For payment_intent.succeeded, we need additional logic to find the tenant
	// This is a simplified implementation - in practice, you'd need to associate
	// payment intents with tenants during creation
	webhookEvent.MarkProcessed()
}

// processChargeSucceeded handles successful charges
func (h *StripeWebhookHandler) processChargeSucceeded(webhookEvent *db.PaymentWebhookEvent, event StripeWebhookEvent) {
	log.Printf("Processing charge.succeeded for event %s", event.ID)

	// Extract amount
	if amount, ok := event.Data.Object["amount"].(float64); ok {
		amountInt := int64(amount)
		webhookEvent.Amount = &amountInt
	}

	webhookEvent.MarkProcessed()
}
