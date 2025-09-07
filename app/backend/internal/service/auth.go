package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bareuptime/tms/internal/auth"
	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/rbac"
	"github.com/bareuptime/tms/internal/redis"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService struct {
	agentRepo     repo.AgentRepository
	rbacService   *rbac.Service
	authService   *auth.Service
	redisService  *redis.Service
	resendService *ResendService
	featureFlags  *FeatureFlags
	domainRepo    *repo.DomainValidationRepo
	tenantRepo    repo.TenantRepository
	projectRepo   repo.ProjectRepository
}

// FeatureFlags represents the feature configuration
type FeatureFlags struct {
	RequireCorporateEmail bool
}

// NewAuthService creates a new auth service
func NewAuthService(agentRepo repo.AgentRepository, rbacService *rbac.Service, authService *auth.Service, redisService *redis.Service, resendService *ResendService, featureFlags *FeatureFlags, tenantRepo repo.TenantRepository, domainRepo *repo.DomainValidationRepo, projectRepo repo.ProjectRepository) *AuthService {
	return &AuthService{
		agentRepo:     agentRepo,
		rbacService:   rbacService,
		authService:   authService,
		redisService:  redisService,
		resendService: resendService,
		featureFlags:  featureFlags,
		tenantRepo:    tenantRepo,
		domainRepo:    domainRepo,
		projectRepo:   projectRepo,
	}
}

// Personal/Consumer email domains that should be blocked for corporate signup
var blockedEmailDomains = map[string]bool{
	// Google
	"gmail.com":      true,
	"googlemail.com": true,

	// Microsoft
	"hotmail.com": true,
	"outlook.com": true,
	"live.com":    true,
	"msn.com":     true,

	// Yahoo
	"yahoo.com":      true,
	"yahoo.co.uk":    true,
	"yahoo.ca":       true,
	"yahoo.co.in":    true,
	"yahoo.com.au":   true,
	"yahoo.fr":       true,
	"yahoo.de":       true,
	"yahoo.it":       true,
	"yahoo.es":       true,
	"ymail.com":      true,
	"rocketmail.com": true,

	// Apple
	"icloud.com": true,
	"me.com":     true,
	"mac.com":    true,

	// AOL
	"aol.com": true,
	"aim.com": true,

	// Other common personal email providers
	"protonmail.com": true,
	"proton.me":      true,
	"tutanota.com":   true,
	"fastmail.com":   true,
	"mailbox.org":    true,
	"posteo.de":      true,
	"hushmail.com":   true,
	"mailfence.com":  true,

	// Common disposable email domains
	"guerrillamail.com":      true,
	"10minutemail.com":       true,
	"tempmail.org":           true,
	"mailinator.com":         true,
	"yopmail.com":            true,
	"dispostable.com":        true,
	"throwaway.email":        true,
	"emailondeck.com":        true,
	"getnada.com":            true,
	"temp-mail.org":          true,
	"fakeinbox.com":          true,
	"sharklasers.com":        true,
	"guerrillamailblock.com": true,
	"pokemail.net":           true,
	"spam4.me":               true,
	"maildrop.cc":            true,
	"mohmal.com":             true,
	"nada.email":             true,
	"tempail.com":            true,
	"disposablemail.com":     true,
	"0-mail.com":             true,
	"1secmail.com":           true,
	"2prong.com":             true,
	"3d-painting.com":        true,
	"4warding.com":           true,
	"7tags.com":              true,
	"9ox.net":                true,
	"aaathats3as.com":        true,
	"abyssmail.com":          true,
	"afrobacon.com":          true,
	"ajaxapp.net":            true,
	"amilegit.com":           true,
	"amiri.net":              true,
	"amiriindustries.com":    true,
	"anonmails.de":           true,
	"anonymbox.com":          true,
}

// Suspicious patterns that indicate disposable or temporary emails
var suspiciousPatterns = []string{
	"temp", "throw", "fake", "disposable", "trash", "delete",
	"remove", "destroy", "kill", "burn", "10min", "20min",
	"minute", "hour", "day", "week", "short", "quick", "fast",
	"instant", "now", "asap", "test", "demo", "sample",
}

// isValidCorporateEmail checks if an email address is from a corporate domain
func (s *AuthService) isValidCorporateEmail(email string) error {
	// Extract domain from email
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email format")
	}

	domain := strings.ToLower(parts[1])

	// Check if domain is in blocked list
	if blockedEmailDomains[domain] {
		return fmt.Errorf("personal email addresses (e.g., Gmail, Yahoo, Hotmail) are not allowed. Please use your company email address")
	}

	// Check for suspicious patterns in domain
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(domain, pattern) {
			return fmt.Errorf("temporary or disposable email addresses are not allowed. Please use your company email address")
		}
	}

	// Additional validation: domain should have at least one dot and be longer than 4 chars
	if len(domain) < 4 || !strings.Contains(domain, ".") {
		return fmt.Errorf("please enter a valid company email address")
	}

	return nil
}

// convertRoleBindings converts role bindings to map format
func (s *AuthService) convertRoleBindings(roleBindings []*db.RoleBinding) map[string][]string {
	result := make(map[string][]string)
	for _, binding := range roleBindings {
		projectKey := ""
		if binding.ProjectID != nil {
			projectKey = binding.ProjectID.String()
		}
		if result[projectKey] == nil {
			result[projectKey] = []string{}
		}
		result[projectKey] = append(result[projectKey], binding.Role.String())
	}
	return result
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string              `json:"access_token"`
	RefreshToken string              `json:"refresh_token"`
	Agent        *db.Agent           `json:"agent"`
	RoleBindings map[string][]string `json:"role_bindings"`
}

// Login authenticates an agent and returns tokens
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {

	// Get agent by email
	agent, err := s.agentRepo.GetByEmailWithoutTenantID(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Verify password
	if agent.PasswordHash == nil {
		return nil, fmt.Errorf("account not configured for password login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*agent.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Get role bindings
	roleBindings, err := s.rbacService.GetAgentRoleBindings(ctx, agent.ID, agent.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", err)
	}

	fmt.Println("Role bindings:", roleBindings)

	// Generate tokens
	accessToken, err := s.authService.GenerateAccessToken(
		agent.ID.String(),
		agent.TenantID.String(),
		agent.Email,
		s.convertRoleBindings(roleBindings),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.authService.GenerateRefreshToken(
		agent.ID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Remove password hash from response
	agent.PasswordHash = nil

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Agent:        agent,
		RoleBindings: s.convertRoleBindings(roleBindings),
	}, nil
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken generates new access token from refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*LoginResponse, error) {
	// Validate refresh token
	claims, err := s.authService.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Get agent
	agentID, err := uuid.Parse(claims.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent ID in token")
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID in token")
	}

	agent, err := s.agentRepo.GetByID(ctx, tenantID, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}

	// Check if agent is active
	if agent.Status != "active" {
		return nil, fmt.Errorf("account is not active")
	}

	// Get role bindings
	roleBindings, err := s.rbacService.GetAgentRoleBindings(ctx, agent.ID, agent.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", err)
	}

	// Generate new access token
	accessToken, err := s.authService.GenerateAccessToken(
		agent.ID.String(),
		agent.TenantID.String(),
		agent.Email,
		s.convertRoleBindings(roleBindings),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Remove password hash from response
	agent.PasswordHash = nil

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken, // Return the same refresh token
		Agent:        agent,
		RoleBindings: s.convertRoleBindings(roleBindings),
	}, nil
}

// MagicLinkRequest represents a magic link request
type MagicLinkRequest struct {
	Email    string `json:"email" validate:"required,email"`
	TenantID string `json:"tenant_id" validate:"required"`
}

// SendMagicLink sends a magic link for passwordless login
func (s *AuthService) SendMagicLink(ctx context.Context, req MagicLinkRequest) error {
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Get agent by email
	agent, err := s.agentRepo.GetByEmail(ctx, tenantID, req.Email)
	if err != nil {
		// Don't reveal if email exists or not
		return nil
	}

	// Check if agent is active
	if agent.Status != "active" {
		return nil
	}

	// Generate magic link token
	token, err := s.authService.GenerateMagicLinkToken(
		agent.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to generate magic link token: %w", err)
	}

	// TODO: Send email with magic link
	// For now, just log the token (in production, this should send an email)
	_ = token

	return nil
}

// ConsumeMagicLinkRequest represents a magic link consumption request
type ConsumeMagicLinkRequest struct {
	Token string `json:"token" validate:"required"`
}

// ConsumeMagicLink exchanges magic link token for access token
func (s *AuthService) ConsumeMagicLink(ctx context.Context, req ConsumeMagicLinkRequest) (*LoginResponse, error) {
	// Validate magic link token
	claims, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid magic link token: %w", err)
	}

	if claims.TokenType != "magic_link" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Get agent
	agentID, err := uuid.Parse(claims.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent ID in token")
	}

	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID in token")
	}

	agent, err := s.agentRepo.GetByID(ctx, tenantID, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found")
	}

	// Check if agent is active
	if agent.Status != "active" {
		return nil, fmt.Errorf("account is not active")
	}

	// Get role bindings
	roleBindings, err := s.rbacService.GetAgentRoleBindings(ctx, agent.ID, agent.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", err)
	}

	// Generate tokens
	accessToken, err := s.authService.GenerateAccessToken(
		agent.ID.String(),
		agent.TenantID.String(),
		agent.Email,
		s.convertRoleBindings(roleBindings),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.authService.GenerateRefreshToken(
		agent.ID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Remove password hash from response
	agent.PasswordHash = nil

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Agent:        agent,
		RoleBindings: s.convertRoleBindings(roleBindings),
	}, nil
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// SignUpRequest represents a signup request
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// VerifySignupOTPRequest represents OTP verification request
type VerifySignupOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// ResendSignupOTPRequest represents request to resend OTP
type ResendSignupOTPRequest struct {
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
}

// SignUp handles user registration with OTP verification
func (s *AuthService) SignUp(ctx context.Context, req SignUpRequest) error {

	// Validate corporate email domain if required
	if s.featureFlags.RequireCorporateEmail {
		if err := s.isValidCorporateEmail(req.Email); err != nil {
			return err
		}
	}

	domainNameFromEmail := strings.Split(req.Email, "@")[1]
	domainName, err := s.domainRepo.GetDomainByNameWithoutTenant(ctx, domainNameFromEmail)
	if err != nil {
		return fmt.Errorf("failed to get domain by name: %w", err)
	}

	if domainName != nil && domainName.Status == models.DomainValidationStatusVerified {
		return fmt.Errorf("domain is already registered, please verify your email or contact support")
	}

	// Check if agent already exists
	_, err = s.agentRepo.GetByEmailWithoutTenantID(ctx, req.Email)
	if err == nil {
		return fmt.Errorf("agent with email %s already exists", req.Email)
	}

	// Create Tenant

	// Check rate limiting
	attemptKey := fmt.Sprintf("signup_attempts:%s", req.Email)
	attempts, err := s.redisService.GetAttempts(ctx, attemptKey)
	if err != nil {
		return fmt.Errorf("failed to check signup attempts: %w", err)
	}

	if attempts >= 3 {
		return fmt.Errorf("too many signup attempts, please try again later")
	}

	// Generate OTP and store it with signup data
	otpKey := fmt.Sprintf("signup_otp:%s", req.Email)
	otp, err := s.redisService.GenerateAndStoreOTP(ctx, otpKey, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Store signup data temporarily
	signupDataKey := fmt.Sprintf("signup_data:%s", req.Email)
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	signupData := map[string]interface{}{
		"email":         req.Email,
		"password_hash": hashedPassword,
		"name":          req.Name,
		// "tenant_id":     tenant.ID.String(),
		// "project_id":    project.ID.String(),
	}

	err = s.redisService.GetClient().HMSet(ctx, signupDataKey, signupData).Err()
	if err != nil {
		return fmt.Errorf("failed to store signup data: %w", err)
	}

	err = s.redisService.GetClient().Expire(ctx, signupDataKey, 10*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration for signup data: %w", err)
	}

	// Send verification email
	err = s.resendService.SendSignupVerificationEmail(ctx, req.Email, otp)
	if err != nil {
		// Clean up on email failure
		s.redisService.DeleteOTP(ctx, otpKey)
		s.redisService.GetClient().Del(ctx, signupDataKey)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// VerifySignupOTP verifies the OTP and creates the agent account
func (s *AuthService) VerifySignupOTP(ctx context.Context, req VerifySignupOTPRequest) (*LoginResponse, error) {
	// Verify OTP
	otpKey := fmt.Sprintf("signup_otp:%s", req.Email)
	isValid, err := s.redisService.VerifyOTP(ctx, otpKey, req.OTP)
	if err != nil {
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}

	if !isValid {
		// Increment failed attempts
		attemptKey := fmt.Sprintf("signup_verify_attempts:%s", req.Email)
		s.redisService.IncrementAttempts(ctx, attemptKey, 1*time.Hour)
		return nil, fmt.Errorf("invalid or expired verification code")
	}

	// Get signup data
	signupDataKey := fmt.Sprintf("signup_data:%s", req.Email)
	signupData, err := s.redisService.GetClient().HGetAll(ctx, signupDataKey).Result()
	if err != nil || len(signupData) == 0 {
		return nil, fmt.Errorf("signup session expired or not found")
	}

	domainNameFromEmail := strings.Split(req.Email, "@")[1]

	tenantID, _ := uuid.NewUUID()
	tenant := &db.Tenant{
		ID:        tenantID,
		Name:      domainNameFromEmail,
		KMSKeyID:  domainNameFromEmail,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.tenantRepo.Create(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create Project
	projectID, _ := uuid.NewUUID()
	project := &db.Project{
		ID:        projectID,
		Key:       "default",
		Name:      "Default",
		Status:    "active",
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Create agent account
	passwordHashStr := signupData["password_hash"]
	agent := &db.Agent{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        signupData["email"],
		Name:         signupData["name"],
		Status:       "active",
		PasswordHash: &passwordHashStr,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	fmt.Println("user obj", agent)

	// Create agent in database
	err = s.agentRepo.Create(ctx, agent)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	fmt.Println("let me assign roles")

	// Assign default agent role (no project ID for global role)
	err = s.rbacService.AssignRole(ctx, agent.ID, agent.TenantID, projectID, models.RoleTenantAdmin)
	if err != nil {
		// Log error but don't fail - agent is created, role can be assigned later
		fmt.Printf("Warning: failed to assign default role to agent %s: %v\n", agent.ID, err)
	}

	// Get role bindings for the new agent
	roleBindings, err := s.rbacService.GetAgentRoleBindings(ctx, agent.ID, agent.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", err)
	}

	fmt.Println("assigning bindings", roleBindings)

	// Generate tokens
	accessToken, err := s.authService.GenerateAccessToken(
		agent.ID.String(),
		agent.TenantID.String(),
		agent.Email,
		s.convertRoleBindings(roleBindings),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.authService.GenerateRefreshToken(
		agent.ID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Clean up temporary data
	// s.redisService.GetClient().Del(ctx, signupDataKey)
	// s.redisService.DeleteOTP(ctx, otpKey)

	// Remove password hash from response
	agent.PasswordHash = nil

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Agent:        agent,
		RoleBindings: s.convertRoleBindings(roleBindings),
	}, nil
}

// ResendSignupOTP resends the verification OTP
func (s *AuthService) ResendSignupOTP(ctx context.Context, req ResendSignupOTPRequest) error {
	// Check if signup data exists
	signupDataKey := fmt.Sprintf("signup_data:%s:%s", req.TenantID, req.Email)
	exists, err := s.redisService.GetClient().Exists(ctx, signupDataKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check signup data: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("no pending signup found for this email")
	}

	// Check rate limiting for resend
	resendKey := fmt.Sprintf("signup_resend:%s:%s", req.TenantID, req.Email)
	attempts, err := s.redisService.GetAttempts(ctx, resendKey)
	if err != nil {
		return fmt.Errorf("failed to check resend attempts: %w", err)
	}

	if attempts >= 3 {
		return fmt.Errorf("too many resend attempts, please try again later")
	}

	// Generate new OTP
	otpKey := fmt.Sprintf("signup_otp:%s:%s", req.TenantID, req.Email)
	otp, err := s.redisService.GenerateAndStoreOTP(ctx, otpKey, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Send verification email
	err = s.resendService.SendSignupVerificationEmail(ctx, req.Email, otp)
	if err != nil {
		s.redisService.DeleteOTP(ctx, otpKey)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	// Increment resend attempts
	s.redisService.IncrementAttempts(ctx, resendKey, 15*time.Minute)

	return nil
}
