package handlers

import (
	"fmt"
	"net/http"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService           *service.AuthService
	publicService         *service.PublicService
	validator             *validator.Validate
	AiAgentLoginAccessKey string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, publicService *service.PublicService, aiAgentLoginAccessKey string) *AuthHandler {
	return &AuthHandler{
		authService:           authService,
		publicService:         publicService,
		validator:             validator.New(),
		AiAgentLoginAccessKey: aiAgentLoginAccessKey,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         User   `json:"user"`
}

// User represents the user data returned in login response
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	loginReq := service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := h.authService.Login(c.Request.Context(), loginReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Determine primary role (use tenant_admin if available, otherwise first role found)
	primaryRole := models.RoleAgent.String() // default
	for _, roles := range response.RoleBindings {
		for _, role := range roles {
			if role == models.RoleTenantAdmin.String() {
				primaryRole = role
				break
			}
			if primaryRole == models.RoleAgent.String() { // Only set if we haven't found a better role
				primaryRole = role
			}
		}
		if primaryRole == models.RoleTenantAdmin.String() {
			break
		}
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		User: User{
			ID:       response.Agent.ID.String(),
			Email:    response.Agent.Email,
			Name:     response.Agent.Name,
			Role:     primaryRole,
			TenantID: response.Agent.TenantID.String(),
		},
	})
}

// Login handles user login
func (h *AuthHandler) AiAgentLogin(c *gin.Context) {
	var req LoginRequest
	s2sKey := c.GetHeader("X-S2S-KEY")
	if s2sKey == "" || s2sKey != h.AiAgentLoginAccessKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing S2S key"})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	loginReq := service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	response, err := h.authService.AiAgentLogin(c.Request.Context(), loginReq, tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Determine primary role (use tenant_admin if available, otherwise first role found)
	primaryRole := models.RoleAgent.String() // default
	for _, roles := range response.RoleBindings {
		for _, role := range roles {
			if role == models.RoleTenantAdmin.String() {
				primaryRole = role
				break
			}
			if primaryRole == models.RoleAgent.String() { // Only set if we haven't found a better role
				primaryRole = role
			}
		}
		if primaryRole == models.RoleTenantAdmin.String() {
			break
		}
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		User: User{
			ID:       response.Agent.ID.String(),
			Email:    response.Agent.Email,
			Name:     response.Agent.Name,
			Role:     primaryRole,
			TenantID: response.Agent.TenantID.String(),
		},
	})
}

// RefreshRequest represents a refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse represents a refresh token response
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	refreshReq := service.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), refreshReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, RefreshResponse{
		AccessToken: response.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600, // 1 hour
	})
}

// GenerateMagicLinkRequest represents a magic link generation request
type GenerateMagicLinkRequest struct {
	TicketID uuid.UUID `json:"ticket_id" validate:"required,uuid"`
}

// GenerateMagicLinkResponse represents a magic link generation response
type GenerateMagicLinkResponse struct {
	MagicLink string `json:"magic_link"`
	ExpiresIn int    `json:"expires_in"`
}

// LogoutResponse represents a logout response
type LogoutResponse struct {
	Message string `json:"message"`
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// For stateless JWT, logout is handled client-side by removing the token
	// In a production system, you might want to maintain a blacklist of revoked tokens
	c.JSON(http.StatusOK, LogoutResponse{
		Message: "Logged out successfully",
	})
}

// MeResponse represents the current user information
type MeResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	TenantID string `json:"tenant_id"`
}

// Me returns current user information
func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, MeResponse{
		ID:       claims.Subject,
		Email:    claims.Email,
		Name:     claims.Email, // Use email as name since Name field doesn't exist
		TenantID: claims.TenantID,
	})
}

// SignUpRequest represents a signup request
type SignUpRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2"`
}

// SignUpResponse represents the signup response (OTP sent)
type SignUpResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

// VerifySignupOTPRequest represents OTP verification request
type VerifySignupOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

// VerifySignupOTPResponse represents successful signup completion
type VerifySignupOTPResponse struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

// ResendSignupOTPRequest represents request to resend OTP
type ResendSignupOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// SignUp handles user registration with OTP verification
func (h *AuthHandler) SignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	signupReq := service.SignUpRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	err := h.authService.SignUp(c.Request.Context(), signupReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SignUpResponse{
		Message: "Verification code sent to your email",
		Email:   req.Email,
	})
}

// VerifySignupOTP handles OTP verification and completes registration
func (h *AuthHandler) VerifySignupOTP(c *gin.Context) {
	var req VerifySignupOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	verifyReq := service.VerifySignupOTPRequest{
		Email: req.Email,
		OTP:   req.OTP,
	}
	fmt.Print("Verifying signup OTP for email:", req.Email)

	response, err := h.authService.VerifySignupOTP(c.Request.Context(), verifyReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get primary role for display
	primaryRole := models.RoleAgent.String() // Default to agent role
	for _, roles := range response.RoleBindings {
		for _, role := range roles {
			primaryRole = role
			break
		}
		if primaryRole == models.RoleTenantAdmin.String() {
			break
		}
	}
	fmt.Println("Primary role assigned:", primaryRole)

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
		User: User{
			ID:       response.Agent.ID.String(),
			Email:    response.Agent.Email,
			Name:     response.Agent.Name,
			Role:     primaryRole,
			TenantID: response.Agent.TenantID.String(),
		},
	})
}

// ResendSignupOTP handles resending OTP for signup verification
func (h *AuthHandler) ResendSignupOTP(c *gin.Context) {
	var req ResendSignupOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	resendReq := service.ResendSignupOTPRequest{
		Email:    req.Email,
		TenantID: tenantID,
	}

	err := h.authService.ResendSignupOTP(c.Request.Context(), resendReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification code resent to your email",
		"email":   req.Email,
	})
}
