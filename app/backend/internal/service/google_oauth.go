package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleOAuthService handles Google OAuth operations
type GoogleOAuthService struct {
	config       *oauth2.Config
	authService  *AuthService
	stateStorage map[string]time.Time // In production, use Redis
}

// GoogleUserInfo represents user information from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// NewGoogleOAuthService creates a new Google OAuth service
func NewGoogleOAuthService(cfg *config.GoogleOAuthConfig, authService *AuthService) *GoogleOAuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleOAuthService{
		config:       oauthConfig,
		authService:  authService,
		stateStorage: make(map[string]time.Time),
	}
}

// GenerateStateToken generates a random state token for CSRF protection
func (s *GoogleOAuthService) GenerateStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state token: %w", err)
	}
	
	state := base64.URLEncoding.EncodeToString(b)
	
	// Store state with expiration (5 minutes)
	s.stateStorage[state] = time.Now().Add(5 * time.Minute)
	
	// Clean up expired states
	go s.cleanupExpiredStates()
	
	return state, nil
}

// ValidateStateToken validates the state token
func (s *GoogleOAuthService) ValidateStateToken(state string) bool {
	expiry, exists := s.stateStorage[state]
	if !exists {
		return false
	}
	
	if time.Now().After(expiry) {
		delete(s.stateStorage, state)
		return false
	}
	
	delete(s.stateStorage, state)
	return true
}

// cleanupExpiredStates removes expired state tokens
func (s *GoogleOAuthService) cleanupExpiredStates() {
	now := time.Now()
	for state, expiry := range s.stateStorage {
		if now.After(expiry) {
			delete(s.stateStorage, state)
		}
	}
}

// GetAuthURL returns the Google OAuth authorization URL
func (s *GoogleOAuthService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges the authorization code for tokens
func (s *GoogleOAuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

// GetUserInfo fetches user information from Google
func (s *GoogleOAuthService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := s.config.Client(ctx, token)
	
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}
	
	return &userInfo, nil
}

// HandleGoogleLogin handles the complete Google OAuth login flow
func (s *GoogleOAuthService) HandleGoogleLogin(ctx context.Context, code string) (*LoginResponse, error) {
	// Exchange code for token
	token, err := s.ExchangeCode(ctx, code)
	if err != nil {
		logger.ErrorfCtx(ctx, err, "Failed to exchange OAuth code")
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}
	
	// Get user info from Google
	userInfo, err := s.GetUserInfo(ctx, token)
	if err != nil {
		logger.ErrorfCtx(ctx, err, "Failed to get user info from Google")
		return nil, fmt.Errorf("failed to get user information: %w", err)
	}
	
	// Validate email is verified
	if !userInfo.VerifiedEmail {
		return nil, fmt.Errorf("email not verified with Google")
	}
	
	logger.InfofCtx(ctx, "Google OAuth login attempt for email: %s", userInfo.Email)
	
	// Check if user exists
	agent, err := s.authService.agentRepo.GetByEmailWithoutTenantID(ctx, userInfo.Email)
	if err != nil {
		// User doesn't exist, create new account via signup flow
		logger.InfofCtx(ctx, "User not found, creating new account for: %s", userInfo.Email)
		
		// Validate corporate email if required
		if s.authService.featureFlags.RequireCorporateEmail {
			if err := s.authService.isValidCorporateEmail(userInfo.Email); err != nil {
				return nil, err
			}
		}
		
		// Create signup request
		signupReq := SignUpRequest{
			Email:    userInfo.Email,
			Password: "", // No password for OAuth users
			Name:     userInfo.Name,
		}
		
		// Create account without OTP verification (Google already verified email)
		response, err := s.authService.SignUpWithOAuth(ctx, signupReq, "google", userInfo.ID)
		if err != nil {
			logger.ErrorfCtx(ctx, err, "Failed to create account via Google OAuth")
			return nil, fmt.Errorf("failed to create account: %w", err)
		}
		
		return response, nil
	}
	
	// User exists, perform login
	logger.InfofCtx(ctx, "Existing user found, logging in: %s", userInfo.Email)
	
	// Get role bindings
	roleBindings, err := s.authService.rbacService.GetAgentRoleBindings(ctx, agent.ID, agent.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role bindings: %w", err)
	}
	
	// Generate tokens
	accessToken, err := s.authService.authService.GenerateAccessToken(
		agent.ID.String(),
		agent.TenantID.String(),
		agent.Email,
		s.authService.convertRoleBindings(roleBindings),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	
	refreshToken, err := s.authService.authService.GenerateRefreshToken(agent.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	// Remove password hash from response
	agent.PasswordHash = nil
	
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Agent:        agent,
		RoleBindings: s.authService.convertRoleBindings(roleBindings),
	}, nil
}
