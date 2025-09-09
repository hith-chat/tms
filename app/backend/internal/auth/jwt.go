package auth

import (
	"fmt"
	"time"

	"github.com/bareuptime/tms/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Claims is an alias for models.JWTClaims for backward compatibility
type Claims = models.JWTClaims

// Service provides authentication functionality
type Service struct {
	secretKey          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	magicLinkExpiry    time.Duration
	unauthTokenExpiry  time.Duration
}

// NewService creates a new auth service
func NewService(secretKey string, accessTokenExpiry, refreshTokenExpiry, magicLinkExpiry, unauthTokenExpiry int) *Service {
	return &Service{
		secretKey:          secretKey,
		accessTokenExpiry:  time.Duration(accessTokenExpiry) * time.Second,
		refreshTokenExpiry: time.Duration(refreshTokenExpiry) * time.Second,
		magicLinkExpiry:    time.Duration(magicLinkExpiry) * time.Second,
		unauthTokenExpiry:  time.Duration(unauthTokenExpiry) * time.Second,
	}
}

// HashPassword hashes a password using bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword verifies a password against its hash
func (s *Service) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateAccessToken generates an access token for an agent
func (s *Service) GenerateAccessToken(agentID, tenantID, email string, roleBindings map[string][]string, expiry *time.Duration) (string, error) {
	now := time.Now()
	jti := uuid.New().String()

	fmt.Println("Generating access token for agent:", roleBindings)

	// Check if roleBindings contains "tenant_admin" role in any of the role lists
	isTenantAdmin := false
	for _, roles := range roleBindings {
		for _, role := range roles {
			if role == models.RoleTenantAdmin.String() {
				isTenantAdmin = true
				break
			}
		}
		if isTenantAdmin {
			break
		}
	}

	expiresIn := s.accessTokenExpiry
	if expiry != nil {
		expiresIn = *expiry
	}

	// Create access token claims
	accessClaims := &models.JWTClaims{
		Sub:           agentID,
		Subject:       agentID, // For backward compatibility
		TenantID:      tenantID,
		AgentID:       agentID,
		Email:         email,
		RoleBindings:  roleBindings,
		TokenType:     "access",
		JTI:           jti,
		Exp:           now.Add(expiresIn).Unix(),
		Iat:           now.Unix(),
		IsTenantAdmin: isTenantAdmin,
	}

	// Generate access token
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	return accessTokenObj.SignedString([]byte(s.secretKey))
}

// GenerateRefreshToken generates a refresh token
func (s *Service) GenerateRefreshToken(agentID string) (string, error) {
	now := time.Now()
	jti := uuid.New().String()

	// Create refresh token claims (simpler)
	refreshClaims := &jwt.RegisteredClaims{
		Subject:   agentID,
		ID:        jti,
		ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	// Generate refresh token
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	return refreshTokenObj.SignedString([]byte(s.secretKey))
}

// GenerateTokens generates access and refresh tokens for an agent
func (s *Service) GenerateTokens(agent *models.Agent, roleBindings map[string][]string) (accessToken, refreshToken string, err error) {
	now := time.Now()
	jti := uuid.New().String()

	// Check if roleBindings contains "tenant_admin" role in any of the role lists
	isTenantAdmin := false
	for _, roles := range roleBindings {
		for _, role := range roles {
			if role == models.RoleTenantAdmin.String() {
				isTenantAdmin = true
				break
			}
		}
		if isTenantAdmin {
			break
		}
	}

	// Create access token claims
	accessClaims := &models.JWTClaims{
		Sub:           agent.ID.String(),
		Subject:       agent.ID.String(),
		TenantID:      agent.TenantID.String(),
		AgentID:       agent.ID.String(),
		Email:         agent.Email,
		RoleBindings:  roleBindings,
		TokenType:     "access",
		JTI:           jti,
		Exp:           now.Add(s.accessTokenExpiry).Unix(),
		Iat:           now.Unix(),
		IsTenantAdmin: isTenantAdmin,
	}

	// Generate access token
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token claims (simpler)
	refreshClaims := &jwt.RegisteredClaims{
		Subject:   agent.ID.String(),
		ID:        jti,
		ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	// Generate refresh token
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// GenerateMagicLinkToken generates a token for magic link authentication
func (s *Service) GenerateMagicLinkToken(agentEmail string) (string, error) {
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		Subject:   agentEmail,
		ID:        uuid.New().String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.magicLinkExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// GeneratePublicToken generates a token for unauthenticated ticket access
func (s *Service) GeneratePublicToken(customerID, ticketID uuid.UUID, scope []string) (string, error) {
	now := time.Now()
	claims := &models.PublicTokenClaims{
		Sub:        "public-ticket",
		CustomerID: customerID,
		TicketID:   ticketID,
		Scope:      scope,
		Exp:        now.Add(s.unauthTokenExpiry).Unix(),
		JTI:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// ValidateToken validates and parses a JWT token
func (s *Service) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// ValidatePublicToken validates and parses a public access token
func (s *Service) ValidatePublicToken(tokenString string) (*models.PublicTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.PublicTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*models.PublicTokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check expiration explicitly as an additional safety measure
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// ValidateMagicLinkToken validates a magic link token
func (s *Service) ValidateMagicLinkToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Check expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return "", fmt.Errorf("token expired")
	}

	return claims.Subject, nil
}

func (s *Service) ValidateChatToken(tokenString string, widgetID uuid.UUID) (*models.PublicChatClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.PublicChatClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Use widgetID as the secret key for chat tokens
		return []byte(widgetID.String()), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*models.PublicChatClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}
