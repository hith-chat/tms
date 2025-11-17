package handlers

import (
	"net/http"

	"github.com/bareuptime/tms/internal/logger"
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
	googleOAuthService    *service.GoogleOAuthService
	validator             *validator.Validate
	AiAgentLoginAccessKey string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, publicService *service.PublicService, googleOAuthService *service.GoogleOAuthService, aiAgentLoginAccessKey string) *AuthHandler {
	logger.Infof("Initializing AuthHandler - operation: init_handler, ai_agent_key_configured: %v", aiAgentLoginAccessKey != "")

	return &AuthHandler{
		authService:           authService,
		publicService:         publicService,
		googleOAuthService:    googleOAuthService,
		validator:             validator.New(),
		AiAgentLoginAccessKey: aiAgentLoginAccessKey,
	}
}

// LoginRequest represents a login request
// @Description User login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// LoginResponse represents a login response
// @Description Successful login response with tokens and user information
type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int    `json:"expires_in" example:"3600"`
	User         User   `json:"user"`
}

// User represents the user data returned in login response
// @Description User information
type User struct {
	ID       string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email    string `json:"email" example:"user@example.com"`
	Name     string `json:"name" example:"John Doe"`
	Role     string `json:"role" example:"agent"`
	TenantID string `json:"tenant_id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Successfully authenticated"
// @Failure 400 {object} map[string]interface{} "Invalid request or validation failed"
// @Failure 401 {object} map[string]interface{} "Authentication failed"
// @Router /v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnfCtx(c.Request.Context(), "Invalid request body for login: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		logger.WarnfCtx(c.Request.Context(), "Login validation failed for email %s: %v", req.Email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	logger.InfofCtx(c.Request.Context(), "Processing login request for email: %s", req.Email)

	loginReq := service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := h.authService.Login(c.Request.Context(), loginReq)
	if err != nil {
		logger.ErrorfCtx(c.Request.Context(), err, "Login authentication failed for email %s: %v", req.Email, err)
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

	logger.InfofCtx(c.Request.Context(), "Login successful - user_id: %s, tenant_id: %s, primary_role: %s",
		response.Agent.ID.String(), response.Agent.TenantID.String(), primaryRole)

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

// AiAgentLogin handles AI agent authentication
// @Summary AI Agent login
// @Description Authenticate AI agent with tenant and project context
// @Tags Auth
// @Accept json
// @Produce json
// @Param X-S2S-KEY header string true "Service-to-service authentication key"
// @Param tenant_id path string true "Tenant ID"
// @Param project_id path string true "Project ID"
// @Param loginRequest body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Successfully authenticated"
// @Failure 400 {object} map[string]interface{} "Invalid request or validation failed"
// @Failure 401 {object} map[string]interface{} "Authentication failed"
// @Router /v1/auth/ai-agent/tenant/{tenant_id}/project/{project_id}/login [post]
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
	tenantIDstr, errT := c.Params.Get("tenant_id")
	logger.InfofCtx(c.Request.Context(), "Tenant ID from param: %s, Found: %v", tenantIDstr, errT)
	if !errT || tenantIDstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}
	tenantID, _ := uuid.Parse(tenantIDstr)
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
// @Description Token refresh request payload
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshResponse represents a refresh token response
// @Description Token refresh response with new access token
type RefreshResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int    `json:"expires_in" example:"3600"`
}

// Refresh handles token refresh
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refreshRequest body RefreshRequest true "Refresh token"
// @Success 200 {object} RefreshResponse "Successfully refreshed token"
// @Failure 400 {object} map[string]interface{} "Invalid request or validation failed"
// @Failure 401 {object} map[string]interface{} "Invalid refresh token"
// @Router /v1/auth/refresh [post]
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
// @Description User registration request payload
type SignUpRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
	Name     string `json:"name" validate:"required,min=2" example:"John Doe"`
}

// SignUpResponse represents the signup response (OTP sent)
// @Description Signup response indicating OTP has been sent
type SignUpResponse struct {
	Message string `json:"message" example:"Verification code sent to your email"`
	Email   string `json:"email" example:"user@example.com"`
}

// VerifySignupOTPRequest represents OTP verification request
// @Description OTP verification request payload
type VerifySignupOTPRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
	OTP   string `json:"otp" validate:"required,len=6" example:"123456"`
}

// VerifySignupOTPResponse represents successful signup completion
// @Description Successful signup verification response
type VerifySignupOTPResponse struct {
	Message string `json:"message" example:"Account verified successfully"`
	User    User   `json:"user"`
}

// ResendSignupOTPRequest represents request to resend OTP
// @Description Request to resend OTP for signup verification
type ResendSignupOTPRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
}

// SignUp handles user registration with OTP verification
// @Summary User registration
// @Description Register a new user account with email verification
// @Tags Auth
// @Accept json
// @Produce json
// @Param signupRequest body SignUpRequest true "User registration details"
// @Success 200 {object} SignUpResponse "OTP sent to email for verification"
// @Failure 400 {object} map[string]interface{} "Invalid request or validation failed"
// @Router /v1/auth/signup [post]
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
// @Summary Verify signup OTP
// @Description Verify the OTP sent during signup to complete user registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param verifyOTPRequest body VerifySignupOTPRequest true "OTP verification details"
// @Success 200 {object} LoginResponse "Account verified and user logged in"
// @Failure 400 {object} map[string]interface{} "Invalid request, validation failed, or invalid OTP"
// @Router /v1/auth/verify-signup-otp [post]
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
	logger.InfofCtx(c.Request.Context(), "Verifying signup OTP for email: %s", req.Email)

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
	logger.InfofCtx(c.Request.Context(), "Primary role assigned: %s", primaryRole)

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
// @Summary Resend signup OTP
// @Description Resend the OTP for signup verification
// @Tags Auth
// @Accept json
// @Produce json
// @Param resendOTPRequest body ResendSignupOTPRequest true "Email to resend OTP to"
// @Success 200 {object} map[string]interface{} "OTP resent successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or validation failed"
// @Router /v1/auth/resend-signup-otp [post]
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

// GoogleOAuthLogin initiates Google OAuth flow
// @Summary Initiate Google OAuth login
// @Description Redirects user to Google OAuth consent screen
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]interface{} "OAuth URL returned"
// @Router /v1/auth/google/login [get]
func (h *AuthHandler) GoogleOAuthLogin(c *gin.Context) {
	if h.googleOAuthService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google OAuth is not configured"})
		return
	}

	// Generate state token for CSRF protection
	state, err := h.googleOAuthService.GenerateStateToken()
	if err != nil {
		logger.ErrorfCtx(c.Request.Context(), err, "Failed to generate OAuth state token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate OAuth flow"})
		return
	}

	// Get OAuth URL
	authURL := h.googleOAuthService.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// GoogleOAuthCallback handles Google OAuth callback
// @Summary Handle Google OAuth callback
// @Description Processes OAuth callback from Google and completes login
// @Tags Auth
// @Accept json
// @Produce json
// @Param code query string true "OAuth authorization code"
// @Param state query string true "OAuth state token"
// @Success 200 {object} LoginResponse "Successfully authenticated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Authentication failed"
// @Router /v1/auth/google/callback [get]
func (h *AuthHandler) GoogleOAuthCallback(c *gin.Context) {
	if h.googleOAuthService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google OAuth is not configured"})
		return
	}

	// Get code and state from query parameters
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code or state parameter"})
		return
	}

	// Validate state token
	if !h.googleOAuthService.ValidateStateToken(state) {
		logger.WarnfCtx(c.Request.Context(), "Invalid OAuth state token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state token"})
		return
	}

	// Handle Google login
	response, err := h.googleOAuthService.HandleGoogleLogin(c.Request.Context(), code)
	if err != nil {
		logger.ErrorfCtx(c.Request.Context(), err, "Google OAuth login failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Determine primary role
	primaryRole := models.RoleAgent.String()
	for _, roles := range response.RoleBindings {
		for _, role := range roles {
			if role == models.RoleTenantAdmin.String() {
				primaryRole = role
				break
			}
			if primaryRole == models.RoleAgent.String() {
				primaryRole = role
			}
		}
		if primaryRole == models.RoleTenantAdmin.String() {
			break
		}
	}

	logger.InfofCtx(c.Request.Context(), "Google OAuth login successful - user_id: %s, tenant_id: %s, primary_role: %s",
		response.Agent.ID.String(), response.Agent.TenantID.String(), primaryRole)

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: User{
			ID:       response.Agent.ID.String(),
			Email:    response.Agent.Email,
			Name:     response.Agent.Name,
			Role:     primaryRole,
			TenantID: response.Agent.TenantID.String(),
		},
	})
}
