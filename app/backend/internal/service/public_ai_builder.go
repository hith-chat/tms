package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/repo"
)

// PublicTenantID is the fixed tenant ID for all public AI widget builder projects
var PublicTenantID = uuid.MustParse("79634696-84ca-11f0-b71d-063b1faa412d")

// PublicProjectID is the fixed project ID for all public AI widget builder projects
var PublicProjectID = uuid.MustParse("ec33f64a-3560-4bbe-80aa-f97c5409c9f0")

// PublicAIBuilderService handles public AI widget builder requests
type PublicAIBuilderService struct {
	projectRepo        repo.ProjectRepository
	aiBuilderService   *AIBuilderService
	webScrapingService *WebScrapingService
}

// NewPublicAIBuilderService creates a new public AI builder service
func NewPublicAIBuilderService(
	projectRepo repo.ProjectRepository,
	aiBuilderService *AIBuilderService,
	webScrapingService *WebScrapingService,
) *PublicAIBuilderService {
	return &PublicAIBuilderService{
		projectRepo:        projectRepo,
		aiBuilderService:   aiBuilderService,
		webScrapingService: webScrapingService,
	}
}

// BuildPublicWidget creates a public AI widget for the given URL
// It uses the fixed public tenant and project IDs
// Returns streaming events via the events channel
func (s *PublicAIBuilderService) BuildPublicWidget(
	ctx context.Context,
	targetURL string,
	depth int,
	events chan<- AIBuilderEvent,
) (uuid.UUID, error) {
	if depth <= 0 {
		depth = 3 // Default depth
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Invalid URL provided",
			Detail:  err.Error(),
		})
		return uuid.Nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Invalid URL provided",
			Detail:  "URL must include both scheme (http/https) and host",
		})
		return uuid.Nil, fmt.Errorf("invalid URL: missing scheme or host")
	}

	if parsedURL.Scheme != "https" {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Invalid URL provided",
			Detail:  "Only HTTPS URLs are supported",
		})
		return uuid.Nil, fmt.Errorf("invalid URL: only HTTPS URLs are supported")
	}

	domain := parsedURL.Host
	// Remove www. prefix for consistency
	domain = strings.TrimPrefix(domain, "www.")

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "builder_started",
		Stage:   "initialization",
		Message: fmt.Sprintf("Starting public AI widget builder for %s", domain),
		Data: map[string]any{
			"domain":     domain,
			"url":        targetURL,
			"tenant_id":  PublicTenantID.String(),
			"project_id": PublicProjectID.String(),
		},
	})

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "project_ready",
		Stage:   "initialization",
		Message: fmt.Sprintf("Using public project with ID: %s", PublicProjectID),
		Data: map[string]any{
			"project_id": PublicProjectID.String(),
			"tenant_id":  PublicTenantID.String(),
			"build_id":   uuid.New().String(),
		},
	})

	// Run the AI builder service
	err = s.aiBuilderService.Run(ctx, PublicTenantID, PublicProjectID, targetURL, depth, false, events)
	if err != nil {
		logger.GetTxLogger(ctx).Error().
			Str("project_id", PublicProjectID.String()).
			Err(err).
			Msg("AI builder failed for public project")
		return PublicProjectID, err
	}

	return PublicProjectID, nil
}

// createPublicProject creates a new public project for the given domain
func (s *PublicAIBuilderService) createPublicProject(
	ctx context.Context,
	domain string,
	targetURL string,
	events chan<- AIBuilderEvent,
) (*db.Project, error) {
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "project_creation_started",
		Stage:   "initialization",
		Message: "Creating public project",
	})

	// Generate project key from domain
	// e.g., "example.com" -> "EXAMPLE-COM-PUBLIC"
	projectKey := s.generateProjectKey(domain)

	// Generate project name
	projectName := fmt.Sprintf("%s Public Widget", domain)

	// Calculate expiration time (6 hours from now)
	expiresAt := time.Now().Add(6 * time.Hour)

	project := &db.Project{
		ID:        uuid.New(),
		TenantID:  PublicTenantID,
		Key:       projectKey,
		Name:      projectName,
		Status:    "active",
		IsPublic:  true,
		ExpiresAt: &expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.projectRepo.Create(ctx, project)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Failed to create public project",
			Detail:  err.Error(),
		})
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	logger.GetTxLogger(ctx).Info().
		Str("project_id", project.ID.String()).
		Str("domain", domain).
		Time("expires_at", expiresAt).
		Msg("Created public project")

	return project, nil
}

// generateProjectKey generates a project key from a domain
// e.g., "example.com" -> "EXAMPLE-COM-PUBLIC"
// e.g., "docs.anthropic.com" -> "DOCS-ANTHROPIC-COM-PUBLIC"
func (s *PublicAIBuilderService) generateProjectKey(domain string) string {
	// Remove any port numbers
	domain = strings.Split(domain, ":")[0]

	// Replace dots with dashes
	key := strings.ReplaceAll(domain, ".", "-")

	// Remove any non-alphanumeric characters except dashes
	reg := regexp.MustCompile("[^a-zA-Z0-9-]+")
	key = reg.ReplaceAllString(key, "")

	// Convert to uppercase and add PUBLIC suffix
	key = strings.ToUpper(key) + "-PUBLIC"

	// Ensure key doesn't exceed 50 characters (database limit)
	if len(key) > 50 {
		key = key[:47] + "PUB" // Truncate and ensure PUBLIC suffix
	}

	return key
}

// emit sends an event with a timestamp
func (s *PublicAIBuilderService) emit(ctx context.Context, events chan<- AIBuilderEvent, event AIBuilderEvent) {
	event.Timestamp = time.Now()
	select {
	case <-ctx.Done():
	case events <- event:
	}
}

// ExtractURLsDebug extracts all URLs from a website for debugging purposes
func (s *PublicAIBuilderService) ExtractURLsDebug(ctx context.Context, targetURL string, depth int, events chan<- URLExtractionEvent) error {
	return s.webScrapingService.ExtractURLsWithStream(ctx, targetURL, depth, events)
}
