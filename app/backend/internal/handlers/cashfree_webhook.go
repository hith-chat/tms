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

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

type CashfreeWebhookHandler struct {
	WebhookRepo   PaymentWebhookRepository
	CreditsRepo   repo.CreditsRepository
	TenantRepo    repo.TenantRepository
	WebhookSecret string
}

func NewCashfreeWebhookHandler(webhookRepo PaymentWebhookRepository, creditsRepo repo.CreditsRepository, tenantRepo repo.TenantRepository, webhookSecret string) *CashfreeWebhookHandler {
	return &CashfreeWebhookHandler{
		WebhookRepo:   webhookRepo,
		CreditsRepo:   creditsRepo,
		TenantRepo:    tenantRepo,
		WebhookSecret: webhookSecret,
	}
}

// CashfreeWebhookEvent represents the structure of a Cashfree webhook event
type CashfreeWebhookEvent struct {
	Type      string          `json:"type"`
	EventTime string          `json:"event_time"`
	Data      json.RawMessage `json:"data"`
}

// CashfreeFormOrderData represents the data structure for PAYMENT_FORM_ORDER_WEBHOOK
type CashfreeFormOrderData struct {
	Form struct {
		FormID       *string `json:"form_id"`
		CfFormID     int     `json:"cf_form_id"`
		FormURL      string  `json:"form_url"`
		FormCurrency string  `json:"form_currency"`
	} `json:"form"`
	Order struct {
		OrderAmount     string `json:"order_amount"`
		OrderID         string `json:"order_id"`
		OrderStatus     string `json:"order_status"`
		TransactionID   int    `json:"transaction_id"`
		CustomerDetails struct {
			CustomerPhone  string `json:"customer_phone"`
			CustomerEmail  string `json:"customer_email"`
			CustomerName   string `json:"customer_name"`
			CustomerFields []struct {
				Title string `json:"title"`
				Value string `json:"value"`
			} `json:"customer_fields"`
		} `json:"customer_details"`
		AmountDetails []struct {
			Title          string `json:"title"`
			Value          string `json:"value"`
			Quantity       int    `json:"quantity,omitempty"`
			SelectedOption string `json:"selectedoption,omitempty"`
		} `json:"amount_details"`
	} `json:"order"`
}

// CashfreePaymentSuccessData represents the data structure for PAYMENT_SUCCESS_WEBHOOK
type CashfreePaymentSuccessData struct {
	Order struct {
		OrderID       string                 `json:"order_id"`
		OrderAmount   json.Number            `json:"order_amount"`
		OrderCurrency string                 `json:"order_currency"`
		OrderTags     map[string]interface{} `json:"order_tags"`
	} `json:"order"`
	Payment struct {
		CfPaymentID     int64       `json:"cf_payment_id"`
		PaymentStatus   string      `json:"payment_status"`
		PaymentAmount   json.Number `json:"payment_amount"`
		PaymentCurrency string      `json:"payment_currency"`
		PaymentMessage  string      `json:"payment_message"`
		PaymentTime     string      `json:"payment_time"`
		BankReference   string      `json:"bank_reference"`
		AuthID          *string     `json:"auth_id"`
		PaymentMethod   struct {
			UPI *struct {
				Channel *string `json:"channel"`
				UPIID   string  `json:"upi_id"`
			} `json:"upi,omitempty"`
		} `json:"payment_method"`
		PaymentGroup string `json:"payment_group"`
	} `json:"payment"`
	CustomerDetails struct {
		CustomerName  string  `json:"customer_name"`
		CustomerID    *string `json:"customer_id"`
		CustomerEmail string  `json:"customer_email"`
		CustomerPhone string  `json:"customer_phone"`
	} `json:"customer_details"`
	PaymentGatewayDetails struct {
		GatewayName             string  `json:"gateway_name"`
		GatewayOrderID          *string `json:"gateway_order_id"`
		GatewayPaymentID        *string `json:"gateway_payment_id"`
		GatewayStatusCode       *string `json:"gateway_status_code"`
		GatewayOrderReferenceID *string `json:"gateway_order_reference_id"`
		GatewaySettlement       string  `json:"gateway_settlement"`
	} `json:"payment_gateway_details"`
	PaymentOffers interface{} `json:"payment_offers"`
}

// HandleCashfreeWebhook processes incoming Cashfree webhook events
// @Summary Handle Cashfree webhook
// @Description Process incoming Cashfree payment webhook events for payment status updates
// @Tags webhooks
// @Accept json
// @Produce json
// @Param X-CF-Signature header string true "Cashfree webhook signature for verification"
// @Param webhook body CashfreeWebhookEvent true "Cashfree webhook event payload"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized - Invalid signature"
// @Failure 500 {string} string "Internal Server Error"
// @Router /webhooks/cashfree [post]
func (h *CashfreeWebhookHandler) HandleCashfreeWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading Cashfree webhook body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Verify the webhook signature (if Cashfree provides signature verification)
	signature := r.Header.Get("X-Cashfree-Signature")
	if signature != "" && !h.verifySignature(body, signature) {
		log.Printf("Invalid Cashfree webhook signature")
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	// Parse the webhook event
	var event CashfreeWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing Cashfree webhook JSON: %v", err)
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Parse event data based on event type and create webhook event record
	var webhookEvent *db.PaymentWebhookEvent
	var eventID string

	switch event.Type {
	case "PAYMENT_FORM_ORDER_WEBHOOK":
		var formOrderData CashfreeFormOrderData
		if err := json.Unmarshal(event.Data, &formOrderData); err != nil {
			log.Printf("Error parsing PAYMENT_FORM_ORDER_WEBHOOK data: %v", err)
			http.Error(w, "Error parsing webhook data", http.StatusBadRequest)
			return
		}

		eventID = fmt.Sprintf("cf_%s_%d", formOrderData.Order.OrderID, formOrderData.Order.TransactionID)
		webhookEvent = &db.PaymentWebhookEvent{
			EventID:        eventID,
			EventType:      event.Type,
			PayloadData:    json.RawMessage(body),
			Status:         "pending",
			ObjectID:       formOrderData.Order.OrderID,
			Currency:       formOrderData.Form.FormCurrency,
			TenantEmail:    formOrderData.Order.CustomerDetails.CustomerEmail,
			PaymentGateway: "cashfree",
		}

		// Convert amount to cents
		if formOrderData.Order.OrderAmount != "" {
			orderAmount, err := strconv.ParseFloat(formOrderData.Order.OrderAmount, 64)
			if err != nil {
				log.Printf("Error parsing order amount '%s': %v", formOrderData.Order.OrderAmount, err)
			} else if orderAmount > 0 {
				amountInCents := int64(orderAmount * 100)
				webhookEvent.Amount = &amountInCents
			}
		}

	case "PAYMENT_SUCCESS_WEBHOOK":
		var successData CashfreePaymentSuccessData
		if err := json.Unmarshal(event.Data, &successData); err != nil {
			log.Printf("Error parsing PAYMENT_SUCCESS_WEBHOOK data: %v", err)
			http.Error(w, "Error parsing webhook data", http.StatusBadRequest)
			return
		}

		eventID = fmt.Sprintf("cf_%s_%d", successData.Order.OrderID, successData.Payment.CfPaymentID)
		webhookEvent = &db.PaymentWebhookEvent{
			EventID:        eventID,
			EventType:      event.Type,
			PayloadData:    json.RawMessage(body),
			Status:         "pending",
			ObjectID:       successData.Order.OrderID,
			Currency:       successData.Order.OrderCurrency,
			TenantEmail:    successData.CustomerDetails.CustomerEmail,
			PaymentGateway: "cashfree",
		}

		// Convert amount to cents
		if successData.Payment.PaymentAmount != "" {
			paymentAmount, err := successData.Payment.PaymentAmount.Float64()
			if err != nil {
				log.Printf("Error parsing payment amount '%s': %v", successData.Payment.PaymentAmount, err)
			} else if paymentAmount > 0 {
				amountInCents := int64(paymentAmount * 100)
				webhookEvent.Amount = &amountInCents
			}
		}

	default:
		log.Printf("Ignoring unhandled Cashfree event type: %s", event.Type)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Event type ignored",
		})
		return
	}

	// Check if we've already processed this event (idempotency)
	existingEvent, err := h.WebhookRepo.FindByPaymentEventID(eventID)
	if err != nil {
		log.Printf("Error checking for existing Cashfree event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingEvent != nil {
		// Event already processed, return success to avoid retries
		log.Printf("Cashfree event %s already processed, skipping", eventID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Event already processed",
		})
		return
	}

	// Process different event types
	switch event.Type {
	case "PAYMENT_FORM_ORDER_WEBHOOK":
		var formOrderData CashfreeFormOrderData
		if err := json.Unmarshal(event.Data, &formOrderData); err == nil {
			h.processPaymentFormOrderWebhook(webhookEvent, formOrderData)
		}
	case "PAYMENT_SUCCESS_WEBHOOK":
		var successData CashfreePaymentSuccessData
		if err := json.Unmarshal(event.Data, &successData); err == nil {
			h.processPaymentSuccessWebhook(webhookEvent, successData)
		}
	default:
		// Mark as ignored for unhandled event types
		webhookEvent.MarkIgnored()
		log.Printf("Ignoring unhandled Cashfree event type: %s", event.Type)
	}

	// Save the webhook event to database
	if err := h.WebhookRepo.Create(webhookEvent); err != nil {
		log.Printf("Error saving Cashfree webhook event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update the event if it was processed
	if webhookEvent.Status != "pending" {
		if err := h.WebhookRepo.Update(webhookEvent); err != nil {
			log.Printf("Error updating Cashfree webhook event: %v", err)
		}
	}

	// Return success immediately to Cashfree
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Cashfree webhook processed",
	})
}

// verifySignature verifies the Cashfree webhook signature
func (h *CashfreeWebhookHandler) verifySignature(body []byte, signature string) bool {
	if h.WebhookSecret == "" {
		log.Printf("Warning: No Cashfree webhook secret configured, skipping signature verification")
		return true // In development, you might want to skip verification
	}

	// Parse the signature header (format may vary based on Cashfree's implementation)
	// This is a basic HMAC SHA256 verification - adjust based on Cashfree's actual signature format
	parts := strings.Split(signature, "=")
	if len(parts) != 2 {
		return false
	}

	// Compute the HMAC
	mac := hmac.New(sha256.New, []byte(h.WebhookSecret))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Compare with provided signature
	return hmac.Equal([]byte(expectedSig), []byte(parts[1]))
}

// processPaymentFormOrderWebhook handles PAYMENT_FORM_ORDER_WEBHOOK events
func (h *CashfreeWebhookHandler) processPaymentFormOrderWebhook(webhookEvent *db.PaymentWebhookEvent, formOrderData CashfreeFormOrderData) {
	log.Printf("Processing PAYMENT_FORM_ORDER_WEBHOOK for order %s", formOrderData.Order.OrderID)

	// Check if the order was paid successfully
	if formOrderData.Order.OrderStatus == "PAID" {
		tenantEmail := formOrderData.Order.CustomerDetails.CustomerEmail

		if tenantEmail != "" {
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

				// Check if amount is sufficient
				if webhookEvent.Amount == nil || *webhookEvent.Amount <= 0 {
					webhookEvent.MarkFailed("Invalid amount in payment")
					return
				}

				// Calculate credits from payment
				creditAmount := models.CalculateCreditsFromPayment(*webhookEvent.Amount)

				// Add credits to tenant account
				_, err := h.CreditsRepo.AddCredits(
					context.Background(),
					tenant.ID,
					creditAmount,
					models.TransactionTypePayment,
					models.PaymentGatewayCashfree,
					webhookEvent.EventID,
					fmt.Sprintf("Cashfree payment via form order: %s", formOrderData.Order.OrderID),
				)
				if err != nil {
					errorMsg := fmt.Sprintf("Error adding credits to tenant %s: %v", tenant.ID, err)
					log.Printf("Error: %s", errorMsg)
					webhookEvent.MarkFailed(errorMsg)
					return
				}

				log.Printf("Successfully added %d credits to tenant %s (%s) via Cashfree payment", creditAmount, tenant.ID, tenantEmail)
				webhookEvent.MarkProcessed()
			} else {
				errorMsg := fmt.Sprintf("Tenant not found with email: %s", tenantEmail)
				log.Printf("Error: %s", errorMsg)
				webhookEvent.MarkFailed(errorMsg)
			}
		} else {
			errorMsg := "No customer email found in Cashfree webhook"
			log.Printf("Error: %s", errorMsg)
			webhookEvent.MarkFailed(errorMsg)
		}
	} else {
		// Order was not paid successfully
		log.Printf("Order %s status is %s, not processing", formOrderData.Order.OrderID, formOrderData.Order.OrderStatus)
		webhookEvent.MarkIgnored()
	}
}

// processPaymentSuccessWebhook handles PAYMENT_SUCCESS_WEBHOOK events
func (h *CashfreeWebhookHandler) processPaymentSuccessWebhook(webhookEvent *db.PaymentWebhookEvent, successData CashfreePaymentSuccessData) {
	log.Printf("Processing PAYMENT_SUCCESS_WEBHOOK for order %s", successData.Order.OrderID)

	// Check if the payment was successful
	if successData.Payment.PaymentStatus == "SUCCESS" {
		tenantEmail := successData.CustomerDetails.CustomerEmail

		if tenantEmail != "" {
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

				// Check if amount is sufficient
				if webhookEvent.Amount == nil || *webhookEvent.Amount <= 0 {
					webhookEvent.MarkFailed("Invalid amount in payment")
					return
				}

				// Calculate credits from payment
				creditAmount := models.CalculateCreditsFromPayment(*webhookEvent.Amount)

				// Add credits to tenant account
				_, err := h.CreditsRepo.AddCredits(
					context.Background(),
					tenant.ID,
					creditAmount,
					models.TransactionTypePayment,
					models.PaymentGatewayCashfree,
					webhookEvent.EventID,
					fmt.Sprintf("Cashfree payment success: %s", successData.Order.OrderID),
				)
				if err != nil {
					errorMsg := fmt.Sprintf("Error adding credits to tenant %s: %v", tenant.ID, err)
					log.Printf("Error: %s", errorMsg)
					webhookEvent.MarkFailed(errorMsg)
					return
				}

				log.Printf("Successfully added %d credits to tenant %s (%s) via Cashfree payment", creditAmount, tenant.ID, tenantEmail)
				webhookEvent.MarkProcessed()
			} else {
				errorMsg := fmt.Sprintf("Tenant not found with email: %s", tenantEmail)
				log.Printf("Error: %s", errorMsg)
				webhookEvent.MarkFailed(errorMsg)
			}
		} else {
			errorMsg := "No customer email found in Cashfree webhook"
			log.Printf("Error: %s", errorMsg)
			webhookEvent.MarkFailed(errorMsg)
		}
	} else {
		// Payment was not successful
		log.Printf("Payment for order %s status is %s, not processing", successData.Order.OrderID, successData.Payment.PaymentStatus)
		webhookEvent.MarkIgnored()
	}
}
