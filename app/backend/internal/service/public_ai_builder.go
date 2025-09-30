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
var PublicTenantID = uuid.MustParse("550e8400-e29b-41d4-a216-446655440000")

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
// It automatically creates a project, scrapes the website, and builds the widget
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
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Invalid URL provided",
			Detail:  err.Error(),
		})
		return uuid.Nil, fmt.Errorf("invalid URL: %w", err)
	}

	domain := parsedURL.Host
	// Remove www. prefix for consistency
	domain = strings.TrimPrefix(domain, "www.")

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "builder_started",
		Stage:   "initialization",
		Message: fmt.Sprintf("Starting public AI widget builder for %s", domain),
		Data: map[string]any{
			"domain": domain,
			"url":    targetURL,
		},
	})

	// Check if a public project already exists for this domain
	existingProject, err := s.projectRepo.GetActivePublicProjectByDomain(ctx, PublicTenantID, domain)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Failed to check for existing projects",
			Detail:  err.Error(),
		})
		return uuid.Nil, fmt.Errorf("failed to check existing projects: %w", err)
	}

	if existingProject != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "A public widget already exists for this domain",
			Detail:  fmt.Sprintf("Project %s already exists and has not expired yet", existingProject.ID),
			Data: map[string]any{
				"existing_project_id": existingProject.ID.String(),
				"expires_at":          existingProject.ExpiresAt,
			},
		})
		return uuid.Nil, fmt.Errorf("public widget already exists for domain %s", domain)
	}

	// Create a new public project
	project, err := s.createPublicProject(ctx, domain, targetURL, events)
	if err != nil {
		return uuid.Nil, err
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "project_created",
		Stage:   "initialization",
		Message: fmt.Sprintf("Public project created with ID: %s", project.ID),
		Data: map[string]any{
			"project_id": project.ID.String(),
			"build_id":   project.ID.String(), // project_id is the build_id
			"expires_at": project.ExpiresAt.Format(time.RFC3339),
		},
	})

	// Run the AI builder service
	err = s.aiBuilderService.Run(ctx, PublicTenantID, project.ID, targetURL, depth, events)
	if err != nil {
		logger.GetTxLogger(ctx).Error().
			Str("project_id", project.ID.String()).
			Err(err).
			Msg("AI builder failed for public project")
		return project.ID, err
	}

	return project.ID, nil
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