package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bareuptime/tms/internal/auth"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/gin-gonic/gin"
)

// ApiKeyAuthMiddleware handles API key authentication
// This middleware checks for x-api-key header and validates the API key
// It sets tenant_id, agent_id, and project_id in the context if the key is valid
func ApiKeyAuthMiddleware(apiKeyRepo repo.ApiKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for x-api-key header
		apiKey := c.GetHeader("x-api-key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "x-api-key header required"})
			c.Abort()
			return
		}

		// Hash the API key for database lookup
		keyHash := repo.HashApiKey(apiKey)

		// Look up the API key in the database
		apiKeyRecord, err := apiKeyRepo.GetByHash(c.Request.Context(), keyHash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// Check if the API key is active
		if !apiKeyRecord.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is inactive"})
			c.Abort()
			return
		}

		// Check if the API key has expired
		if apiKeyRecord.ExpiresAt != nil && apiKeyRecord.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key has expired"})
			c.Abort()
			return
		}

		// Set context values from the API key record
		c.Set("tenant_id", apiKeyRecord.TenantID.String())
		c.Set("agent_id", apiKeyRecord.AgentID.String())
		c.Set("project_id", apiKeyRecord.ProjectID.String())
		c.Set("api_key_auth", true) // Flag to indicate this is API key auth
		c.Set("api_key_id", apiKeyRecord.ID.String())

		// Update last used timestamp asynchronously to avoid blocking the request
		go func() {
			_ = apiKeyRepo.UpdateLastUsed(c.Request.Context(), apiKeyRecord.ID)
		}()

		c.Next()
	}
}

// ApiKeyOrJWTAuthMiddleware allows either API key or JWT authentication
// This is useful for endpoints that should support both authentication methods
func ApiKeyOrJWTAuthMiddleware(apiKeyRepo repo.ApiKeyRepository, jwtAuth *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for public endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/v1/public/") {
			c.Next()
			return
		}

		// Check for API key first
		apiKey := c.GetHeader("x-api-key")
		if apiKey != "" {
			fmt.Println("API Key Auth Middleware: x-api-key header found")
			// Use API key authentication
			handleApiKeyAuth(c, apiKeyRepo, apiKey)
			return
		}

		// Fall back to JWT authentication
		authHeader := c.GetHeader("Authorization")
		queryTokenParam := c.Query("token")
		if authHeader != "" || queryTokenParam != "" {
			// Use existing JWT auth logic
			handleJWTAuth(c, jwtAuth, authHeader, queryTokenParam)
			return
		}

		// No authentication provided
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required: provide either x-api-key header or Authorization header"})
		c.Abort()
	}
}

// handleApiKeyAuth handles the API key authentication logic
func handleApiKeyAuth(c *gin.Context, apiKeyRepo repo.ApiKeyRepository, apiKey string) {
	// Hash the API key for database lookup
	keyHash := repo.HashApiKey(apiKey)

	// Look up the API key in the database
	apiKeyRecord, err := apiKeyRepo.GetByHash(c.Request.Context(), keyHash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		c.Abort()
		return
	}

	// Check if the API key is active
	if !apiKeyRecord.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is inactive"})
		c.Abort()
		return
	}

	// Check if the API key has expired
	if apiKeyRecord.ExpiresAt != nil && apiKeyRecord.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API key has expired"})
		c.Abort()
		return
	}

	// Set context values from the API key record
	c.Set("tenant_id", apiKeyRecord.TenantID.String())
	c.Set("agent_id", apiKeyRecord.AgentID.String())
	c.Set("project_id", apiKeyRecord.ProjectID.String())
	c.Set("api_key_auth", true) // Flag to indicate this is API key auth
	c.Set("api_key_id", apiKeyRecord.ID.String())

	// Update last used timestamp asynchronously to avoid blocking the request
	go func() {
		_ = apiKeyRepo.UpdateLastUsed(c.Request.Context(), apiKeyRecord.ID)
	}()

	fmt.Println("API Key Auth Middleware: API key validated successfully")

	c.Next()
}

// handleJWTAuth handles the JWT authentication logic (extracted from AuthMiddleware)
func handleJWTAuth(c *gin.Context, jwtAuth *auth.Service, authHeader, queryTokenParam string) {
	// Check Bearer prefix
	const bearerPrefix = "Bearer "
	var token string
	if strings.HasPrefix(authHeader, bearerPrefix) {
		token = authHeader[len(bearerPrefix):]
	} else {
		token = queryTokenParam
	}
	if !strings.HasPrefix(authHeader, bearerPrefix) && queryTokenParam == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		c.Abort()
		return
	}

	// Validate token
	claims, err := jwtAuth.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	// Check token type (should be access token for API endpoints)
	if claims.TokenType != "access" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
		c.Abort()
		return
	}

	// Validate mandatory fields
	if claims.AgentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id missing from token"})
		c.Abort()
		return
	}

	if claims.TenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id missing from token"})
		c.Abort()
		return
	}

	if claims.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email missing from token"})
		c.Abort()
		return
	}

	if claims.RoleBindings == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role_bindings missing from token"})
		c.Abort()
		return
	}

	// Store validated claims in context
	c.Set("tenant_id", claims.TenantID)
	c.Set("email", claims.Email)
	c.Set("agent_id", claims.AgentID)
	c.Set("role_bindings", claims.RoleBindings)
	c.Set("claims", claims)
	c.Set("is_tenant_admin", claims.IsTenantAdmin)
	c.Set("api_key_auth", false) // Flag to indicate this is JWT auth

	c.Next()
}
