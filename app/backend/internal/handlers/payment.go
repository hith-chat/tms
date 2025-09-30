// Package handlers provides HTTP request handlers for the BareUptime application.
// This file contains the payment handler that manages payment session creation
// with automatic gateway selection based on user location.
package handlers

import (
	"net/http"
	"strings"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	paymentService *service.PaymentService
}

// NewPaymentHandler creates a new instance of PaymentHandler
//
// Parameters:
//   - paymentService: Service for payment operations and gateway selection
//
// Returns:
//   - *PaymentHandler: Configured handler instance ready for routing
//
// Example usage:
//
//	paymentHandler := NewPaymentHandler(paymentService)
//	router.POST("/payments/create-session", paymentHandler.CreatePaymentSession)
func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentSession handles POST /payments/create-session
//
// This endpoint creates a payment session with automatic gateway selection:
// 1. Extracts user's IP address from request
// 2. Uses IP geolocation to determine optimal payment gateway
// 3. Creates payment session with appropriate gateway
// 4. Returns payment URL for user redirect
//
// Request Body (JSON):
//   - amount: Payment amount (float64, required, min=1)
//   - currency: Currency code (string, required)
//   - type: Payment type (string, required, e.g., "ai_messages")
//   - gateway: Force specific gateway (string, optional)
//   - success_url: Success redirect URL (string, required, valid URL)
//   - cancel_url: Cancel redirect URL (string, required, valid URL)
//
// Response (JSON):
//   - payment_url: URL to redirect user for payment
//   - session_id: Payment session identifier
//   - gateway: Selected payment gateway
//   - amount: Final amount in local currency
//   - currency: Final currency code
//   - expires_at: Session expiration timestamp
//
// Gateway Selection:
//   - India (IP geolocation): Cashfree
//   - All other countries: Stripe
//   - Manual override via "gateway" parameter
//
// Authentication: JWT required (user must be authenticated)
// Authorization: User can only create payments for their own tenant
//
// Example request:
//
//	POST /v1/payments/create-session
//	Authorization: Bearer <jwt_token>
//	Content-Type: application/json
//
//	{
//	  "amount": 10.0,
//	  "currency": "USD",
//	  "type": "ai_messages",
//	  "success_url": "https://app.example.com/success",
//	  "cancel_url": "https://app.example.com/cancel"
//	}
//
// Example response:
//
//	{
//	  "payment_url": "https://checkout.stripe.com/c/pay/cs_123...",
//	  "session_id": "stripe_session_abc123",
//	  "gateway": "stripe",
//	  "amount": 10.0,
//	  "currency": "USD",
//	  "expires_at": "2023-12-25T12:00:00Z"
//	}
//
// CreatePaymentSession creates a payment session with automatic gateway selection
// @Summary Create payment session
// @Description Create a payment session with automatic gateway selection based on user location
// @Tags payments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Param payment body object{amount=number,currency=string,type=string} true "Payment session request"
// @Success 200 {object} object{payment_url=string,session_id=string,gateway=string,amount=number,currency=string,expires_at=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/payments/create-session [post]
func (h *PaymentHandler) CreatePaymentSession(c *gin.Context) {
	// Extract tenant ID from authenticated user context
	tenantUUID := middleware.GetTenantID(c)
	if tenantUUID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Tenant ID not found in context",
		})
		return
	}

	// Parse request body
	var req service.PaymentSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate payment type (extend this list as needed)
	validTypes := map[string]bool{
		"ai_messages":  true,
		"credits":      true,
		"subscription": true,
	}
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid payment type. Supported types: ai_messages, credits, subscription",
		})
		return
	}

	// Extract client IP address for geolocation
	clientIP := h.getClientIP(c)

	// Create payment session
	response, err := h.paymentService.CreatePaymentSession(
		c.Request.Context(),
		&req,
		clientIP,
		tenantUUID,
	)
	if err != nil {
		// Log error (in production, you might want structured logging)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to create payment session: " + err.Error(),
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, response)
}

// GetPaymentGatewayPreview handles GET /payments/gateway-preview
//
// This endpoint allows users to preview which payment gateway would be selected
// for their location without creating an actual payment session.
//
// Query Parameters:
//   - ip: Optional IP address to check (defaults to request IP)
//
// Response (JSON):
//   - gateway: Recommended payment gateway
//   - country_code: Detected country code
//   - currency: Recommended currency for the region
//
// Authentication: JWT required
//
// Example request:
//
//	GET /v1/payments/gateway-preview
//	Authorization: Bearer <jwt_token>
//
// Example response:
//
//	{
//	  "gateway": "cashfree",
//	  "country_code": "IN",
//	  "currency": "INR"
//	}
//
// GetPaymentGatewayPreview gets payment gateway based on location
// @Summary Get payment gateway preview
// @Description Get the recommended payment gateway based on user's location
// @Tags payments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tenant_id header string true "Tenant ID"
// @Success 200 {object} object{gateway=string,country=string,currency=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/payments/gateway-preview [get]
func (h *PaymentHandler) GetPaymentGatewayPreview(c *gin.Context) {
	// Extract client IP (use query param if provided, otherwise detect)
	clientIP := c.Query("ip")
	if clientIP == "" {
		clientIP = h.getClientIP(c)
	}

	// Get gateway recommendation
	gateway, countryCode, err := h.paymentService.GetPaymentGatewayForIP(
		c.Request.Context(),
		clientIP,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to determine payment gateway: " + err.Error(),
		})
		return
	}

	// Determine recommended currency based on country
	currency := "USD" // Default
	if countryCode == "IN" {
		currency = "INR"
	}

	c.JSON(http.StatusOK, gin.H{
		"gateway":      gateway,
		"country_code": countryCode,
		"currency":     currency,
	})
}

// getClientIP extracts the real client IP address from the request
//
// This method checks various headers in order of preference:
// 1. X-Forwarded-For (proxy/load balancer)
// 2. X-Real-IP (nginx reverse proxy)
// 3. RemoteAddr (direct connection)
//
// Parameters:
//   - c: Gin context containing the HTTP request
//
// Returns:
//   - string: Client IP address
//
// Note: In production, ensure your load balancer/reverse proxy is configured
// to set these headers correctly and that you trust the source.
func (h *PaymentHandler) getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (most common with load balancers)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (nginx reverse proxy)
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr (direct connection)
	// This includes the port, so we need to extract just the IP
	remoteAddr := c.Request.RemoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		return remoteAddr[:idx]
	}

	return remoteAddr
}

// ListPaymentMethods handles GET /payments/methods
//
// This endpoint returns available payment methods for the user's region.
//
// Response (JSON):
//   - methods: Array of available payment methods with details
//
// Authentication: JWT required
//
// Example response:
//
//	{
//	  "methods": [
//	    {
//	      "gateway": "stripe",
//	      "name": "Credit/Debit Card",
//	      "types": ["card", "google_pay", "apple_pay"],
//	      "currencies": ["USD", "EUR", "GBP"]
//	    }
//	  ]
//	}
func (h *PaymentHandler) ListPaymentMethods(c *gin.Context) {
	clientIP := h.getClientIP(c)

	gateway, countryCode, err := h.paymentService.GetPaymentGatewayForIP(
		c.Request.Context(),
		clientIP,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to determine payment methods: " + err.Error(),
		})
		return
	}

	var methods []map[string]interface{}

	switch gateway {
	case service.PaymentGatewayCashfree:
		methods = []map[string]interface{}{
			{
				"gateway":    "cashfree",
				"name":       "UPI",
				"types":      []string{"upi"},
				"currencies": []string{"INR"},
			},
			{
				"gateway":    "cashfree",
				"name":       "Credit/Debit Card",
				"types":      []string{"card"},
				"currencies": []string{"INR"},
			},
			{
				"gateway":    "cashfree",
				"name":       "Net Banking",
				"types":      []string{"netbanking"},
				"currencies": []string{"INR"},
			},
		}
	case service.PaymentGatewayStripe:
		methods = []map[string]interface{}{
			{
				"gateway":    "stripe",
				"name":       "Credit/Debit Card",
				"types":      []string{"card"},
				"currencies": []string{"USD", "EUR", "GBP", "INR"},
			},
			{
				"gateway":    "stripe",
				"name":       "Digital Wallets",
				"types":      []string{"google_pay", "apple_pay"},
				"currencies": []string{"USD", "EUR", "GBP"},
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"methods":      methods,
		"country_code": countryCode,
		"recommended":  gateway,
	})
}
