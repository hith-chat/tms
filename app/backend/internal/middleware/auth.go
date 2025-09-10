package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/bareuptime/tms/internal/auth"
	"github.com/bareuptime/tms/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware handles JWT authentication
func AuthMiddleware(jwtAuth *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for public endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/v1/public/") {
			c.Next()
			return
		}

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		queryTokenParam := c.Query("token")
		if authHeader == "" && queryTokenParam == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header or query token required"})
			c.Abort()
			return
		}

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

		// Validate UUID format for agent_id
		if _, err := uuid.Parse(claims.AgentID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent_id format in token"})
			c.Abort()
			return
		}

		// Validate UUID format for tenant_id
		if _, err := uuid.Parse(claims.TenantID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format in token"})
			c.Abort()
			return
		}

		// Validate that role_bindings has at least one role
		for projectID, roles := range claims.RoleBindings {
			if len(roles) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "no roles found for project " + projectID})
				c.Abort()
				return
			}
		}

		// Store validated claims in context
		c.Set("tenant_id", claims.TenantID)
		c.Set("email", claims.Email)
		c.Set("agent_id", claims.AgentID)
		c.Set("role_bindings", claims.RoleBindings)
		c.Set("claims", claims)
		c.Set("is_tenant_admin", claims.IsTenantAdmin)

		c.Next()
	}
}

// TicketAccessMiddleware ensures agents can only access tickets they have permission for
// Rules: tenant_admin and project_admin can access all tickets in their scope
// Regular agents can only access tickets assigned to them
func TicketAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid claims found"})
			c.Abort()
			return
		}

		// Get IDs from URL parameters
		projectID := c.Param("project_id")
		ticketID := c.Param("ticket_id")

		if projectID == "" || ticketID == "" {
			// If no ticket_id in URL, this middleware doesn't apply
			c.Next()
			return
		}

		// Validate UUID formats
		if _, err := uuid.Parse(projectID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id format"})
			c.Abort()
			return
		}

		if _, err := uuid.Parse(ticketID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket_id format"})
			c.Abort()
			return
		}

		// If user is tenant admin, allow access to all tickets
		if claims.IsTenantAdmin {
			c.Next()
			return
		}

		// Check if user has project_admin role for this project
		roles, projectExists := claims.RoleBindings[projectID]
		if projectExists {
			for _, role := range roles {
				if role == "project_admin" {
					c.Next()
					return
				}
			}
		}

		// For regular agents, we need to check if they are assigned to the ticket
		// This will be handled at the repository level using agent_id filter
		// Store agent_id in context for repository to use
		c.Set("enforce_agent_assignment", true)

		c.Next()
	}
}

// TicketReassignmentMiddleware ensures only tenant_admin and project_admin can reassign tickets
func TicketReassignmentMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid claims found"})
			c.Abort()
			return
		}

		// Check if this is a reassignment request by looking at the request body
		if c.Request.Method == "PATCH" || c.Request.Method == "PUT" {
			// We need to check if assignee_agent_id is being modified
			// This is a simplified check - in production you might want to be more thorough
			c.Set("allow_reassignment", claims.IsTenantAdmin)

			// Check project admin role
			projectID := c.Param("project_id")
			if projectID != "" {
				if roles, exists := claims.RoleBindings[projectID]; exists {
					for _, role := range roles {
						if role == "project_admin" {
							c.Set("allow_reassignment", true)
							break
						}
					}
				}
			}
		}

		c.Next()
	}
}

// TenantMiddleware sets up tenant context for RLS
func TenantMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			// Try to get from URL params for public endpoints
			tenantID = c.Param("tenant_id")
		}

		if tenantID != "" {
			// Validate tenant ID format
			if _, err := uuid.Parse(tenantID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
				c.Abort()
				return
			}

			// Set RLS context in database
			ctx := context.WithValue(c.Request.Context(), "tenant_id", tenantID)
			c.Request = c.Request.WithContext(ctx)

			// Execute SET statement for RLS
			if db != nil {
				_, err := db.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set tenant context"})
					c.Abort()
					return
				}
			}

			c.Set("tenant_id", tenantID)
		}

		c.Next()
	}
}

// CORSMiddleware handles CORS headers with configurable origins
func CORSMiddleware(corsConfig *config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// Determine allowed origin

		allowedOrigin := ""
		if (len(corsConfig.AllowedOrigins) == 0) ||
			(len(corsConfig.AllowedOrigins) == 1 && corsConfig.AllowedOrigins[0] == "*") || (len(corsConfig.AllowedOrigins) == 1 && corsConfig.AllowedOrigins[0] == "") {
			// If allowing all origins and not using credentials, use wildcard
			if !corsConfig.AllowCredentials {
				allowedOrigin = "*"
			} else {
				// If using credentials, echo back the request origin if it's valid
				allowedOrigin = origin
			}
		} else {
			// Check if the origin is in the allowed list
			for _, allowedOrig := range corsConfig.AllowedOrigins {
				if allowedOrig == origin {
					allowedOrigin = origin
					break
				}
			}
		}

		// Set CORS headers
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Session-Token, Cache-Control, Pragma")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Session-Token, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

		if corsConfig.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ErrorHandlerMiddleware handles panics and errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// ValidationMiddleware handles request validation
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This will be used by handlers to validate requests
		c.Next()
	}
}

// Helper functions to extract values from context

// GetTenantID extracts tenant ID from context
func GetTenantID(c *gin.Context) uuid.UUID {
	if tenantID, exists := c.Get("tenant_id"); exists {
		if id, ok := tenantID.(string); ok {
			tenantUUID, _ := uuid.Parse(id)
			return tenantUUID
		}
	}
	// raise error
	err := errors.New("tenant ID not found")
	panic(err)
}

func GetAgentID(c *gin.Context) uuid.UUID {
	if agentID, exists := c.Get("agent_id"); exists {
		if id, ok := agentID.(string); ok {
			agentUUID, _ := uuid.Parse(id)
			return agentUUID
		}
	}
	// raise error
	err := errors.New("agent ID not found")
	panic(err)
}

func GetProjectID(c *gin.Context) uuid.UUID {
	if projectID, exists := c.Params.Get("project_id"); exists {
		projectUUID, _ := uuid.Parse(projectID)
		return projectUUID
	}
	// raise error
	err := errors.New("project ID not found")
	panic(err)
}

func GetWidgetID(c *gin.Context) uuid.UUID {
	if widgetID, exists := c.Params.Get("widget_id"); exists {
		widgetUUID, _ := uuid.Parse(widgetID)
		return widgetUUID
	}
	// raise error
	err := errors.New("widget ID not found")
	panic(err)
}

func GetSessionID(c *gin.Context) uuid.UUID {
	if sessionID, exists := c.Params.Get("session_id"); exists {
		sessionUUID, _ := uuid.Parse(sessionID)
		return sessionUUID
	}
	// raise error
	err := errors.New("session ID not found")
	panic(err)
}

func GetSessionToken(c *gin.Context) string {
	if sessionToken, exists := c.Params.Get("session_token"); exists {
		return sessionToken
	}
	// raise error
	err := errors.New("session token not found")
	panic(err)
}

func GetEmail(c *gin.Context) uuid.UUID {
	if email, exists := c.Get("email"); exists {
		if id, ok := email.(string); ok {
			emailUUID, _ := uuid.Parse(id)
			return emailUUID
		}
	}
	// raise error
	err := errors.New("email not found")
	panic(err)
}

func IsTenantAdmin(c *gin.Context) bool {
	if isAdmin, exists := c.Get("is_tenant_admin"); exists {
		if admin, ok := isAdmin.(bool); ok {
			return admin
		}
	}
	return false
}

func GetRoleBindings(c *gin.Context) []string {
	if roleBindings, exists := c.Get("role_bindings"); exists {
		if roles, ok := roleBindings.([]string); ok {
			return roles
		}
	}
	return []string{}
}

// GetClaims extracts JWT claims from context
func GetClaims(c *gin.Context) *auth.Claims {
	if claims, exists := c.Get("claims"); exists {
		if cl, ok := claims.(*auth.Claims); ok {
			return cl
		}
	}
	return nil
}

// TenantAdminMiddleware ensures only tenant admins can access the endpoint
func TenantAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid claims found"})
			c.Abort()
			return
		}

		if !claims.IsTenantAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tenant admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ProjectAdminMiddleware ensures only project admins can access the endpoint
func ProjectAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid claims found"})
			c.Abort()
			return
		}

		// Get project_id from URL parameter
		projectID := c.Param("project_id")
		if projectID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "project_id parameter is required"})
			c.Abort()
			return
		}

		// Validate project ID format
		if _, err := uuid.Parse(projectID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id format"})
			c.Abort()
			return
		}

		// Check if user is tenant admin (tenant admins have access to all projects)
		if claims.IsTenantAdmin {
			c.Next()
			return
		}

		// Check if project exists in role bindings
		roles, projectExists := claims.RoleBindings[projectID]
		if !projectExists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: no access to this project"})
			c.Abort()
			return
		}

		// Check if user has project_admin role for this project
		hasProjectAdminRole := false
		for _, role := range roles {
			if role == "project_admin" {
				hasProjectAdminRole = true
				break
			}
		}

		if !hasProjectAdminRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: project_admin role required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func ProjectAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid claims found"})
			c.Abort()
			return
		}

		// Get project_id from URL parameter
		projectID := c.Param("project_id")
		if projectID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "project_id parameter is required"})
			c.Abort()
			return
		}

		// Validate project ID format
		if _, err := uuid.Parse(projectID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id format"})
			c.Abort()
			return
		}

		// Check if user is tenant admin (tenant admins have access to all projects)
		if claims.IsTenantAdmin {
			c.Next()
			return
		}

		// Check if project exists in role bindings
		_, projectExists := claims.RoleBindings[projectID]
		if !projectExists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: no access to this project"})
			c.Abort()
			return
		}

		c.Next()
	}
}
