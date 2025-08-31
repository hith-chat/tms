package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/repo"
)

// BrandGreetingService generates brand-aware greeting responses
type BrandGreetingService struct {
	settingsRepo *repo.SettingsRepository
}

// NewBrandGreetingService creates a new brand greeting service
func NewBrandGreetingService(settingsRepo *repo.SettingsRepository) *BrandGreetingService {
	return &BrandGreetingService{
		settingsRepo: settingsRepo,
	}
}

// BrandInfo represents the brand information used for greetings
type BrandInfo struct {
	CompanyName string `json:"company_name"`
	About       string `json:"about"`
	SupportURL  string `json:"support_url"`
}

// GreetingResponse represents a generated greeting response
type GreetingResponse struct {
	Message     string    `json:"message"`
	BrandInfo   BrandInfo `json:"brand_info"`
	Template    string    `json:"template"`
	GeneratedAt time.Time `json:"generated_at"`
}

// GenerateGreetingResponse creates a personalized greeting response using brand settings
func (b *BrandGreetingService) GenerateGreetingResponse(ctx context.Context, tenantID, projectID uuid.UUID, customerMessage string) (*GreetingResponse, error) {
	// Get brand settings
	brandInfo, err := b.getBrandInfo(ctx, tenantID, projectID)
	if err != nil {
		// Use default greeting if brand settings are not available
		return b.generateDefaultGreeting(customerMessage), nil
	}

	// Generate personalized greeting based on brand info
	greeting := b.createBrandedGreeting(brandInfo, customerMessage)

	return &GreetingResponse{
		Message:     greeting,
		BrandInfo:   *brandInfo,
		Template:    "branded_greeting",
		GeneratedAt: time.Now(),
	}, nil
}

// getBrandInfo retrieves brand information from settings
func (b *BrandGreetingService) getBrandInfo(ctx context.Context, tenantID, projectID uuid.UUID) (*BrandInfo, error) {
	settings, _, err := b.settingsRepo.GetSetting(ctx, tenantID, projectID, "branding_settings")
	if err != nil {
		return nil, err
	}

	brandInfo := &BrandInfo{}

	if companyName, ok := settings["company_name"].(string); ok {
		brandInfo.CompanyName = companyName
	}

	if about, ok := settings["about"].(string); ok {
		brandInfo.About = about
	}

	if supportURL, ok := settings["support_url"].(string); ok {
		brandInfo.SupportURL = supportURL
	}

	return brandInfo, nil
}

// createBrandedGreeting generates a greeting response using brand information
func (b *BrandGreetingService) createBrandedGreeting(brandInfo *BrandInfo, customerMessage string) string {
	timeOfDay := b.getTimeOfDay()

	// Choose greeting based on available brand information
	var greeting string

	if brandInfo.CompanyName != "" && brandInfo.About != "" {
		// Full brand info available
		greeting = b.generateFullBrandGreeting(brandInfo, timeOfDay)
	} else if brandInfo.CompanyName != "" {
		// Only company name available
		greeting = b.generateCompanyNameGreeting(brandInfo.CompanyName, timeOfDay)
	} else if brandInfo.About != "" {
		// Only about info available
		greeting = b.generateAboutBasedGreeting(brandInfo.About, timeOfDay)
	} else {
		// No brand info available
		greeting = b.generateGenericGreeting(timeOfDay)
	}

	return greeting
}

// generateFullBrandGreeting creates a greeting with company name and about information
func (b *BrandGreetingService) generateFullBrandGreeting(brandInfo *BrandInfo, timeOfDay string) string {
	templates := []string{
		fmt.Sprintf("%s! Welcome to %s. %s How can we help you today?",
			timeOfDay, brandInfo.CompanyName, strings.TrimSpace(brandInfo.About)),

		fmt.Sprintf("Hello! Thanks for reaching out to %s. %s - What can we assist you with?",
			brandInfo.CompanyName, strings.TrimSpace(brandInfo.About)),

		fmt.Sprintf("Hi there! You've reached %s support. %s How may we help you?",
			brandInfo.CompanyName, strings.TrimSpace(brandInfo.About)),
	}

	// Use a simple hash to consistently pick the same template for the same company
	index := len(brandInfo.CompanyName) % len(templates)
	return templates[index]
}

// generateCompanyNameGreeting creates a greeting with just company name
func (b *BrandGreetingService) generateCompanyNameGreeting(companyName, timeOfDay string) string {
	templates := []string{
		fmt.Sprintf("%s! Welcome to %s. How can we assist you today?", timeOfDay, companyName),
		fmt.Sprintf("Hello! You've reached %s support. What can we help you with?", companyName),
		fmt.Sprintf("Hi there! Thanks for contacting %s. How may we help you?", companyName),
	}

	index := len(companyName) % len(templates)
	return templates[index]
}

// generateAboutBasedGreeting creates a greeting using about information
func (b *BrandGreetingService) generateAboutBasedGreeting(about, timeOfDay string) string {
	templates := []string{
		fmt.Sprintf("%s! %s How can we help you today?", timeOfDay, strings.TrimSpace(about)),
		fmt.Sprintf("Hello! %s What can we assist you with?", strings.TrimSpace(about)),
		fmt.Sprintf("Hi there! %s How may we help you?", strings.TrimSpace(about)),
	}

	index := len(about) % len(templates)
	return templates[index]
}

// generateGenericGreeting creates a generic greeting when no brand info is available
func (b *BrandGreetingService) generateGenericGreeting(timeOfDay string) string {
	templates := []string{
		fmt.Sprintf("%s! Welcome to our support. How can we help you today?", timeOfDay),
		"Hello! Thanks for reaching out. What can we assist you with?",
		"Hi there! How may we help you today?",
	}

	index := int(time.Now().Unix()) % len(templates)
	return templates[index]
}

// generateDefaultGreeting creates a default greeting when brand settings are unavailable
func (b *BrandGreetingService) generateDefaultGreeting(customerMessage string) *GreetingResponse {
	timeOfDay := b.getTimeOfDay()
	greeting := fmt.Sprintf("%s! Thanks for reaching out. How can we help you today?", timeOfDay)

	return &GreetingResponse{
		Message: greeting,
		BrandInfo: BrandInfo{
			CompanyName: "",
			About:       "",
			SupportURL:  "",
		},
		Template:    "default_greeting",
		GeneratedAt: time.Now(),
	}
}

// getTimeOfDay returns an appropriate time-based greeting
func (b *BrandGreetingService) getTimeOfDay() string {
	now := time.Now()
	hour := now.Hour()

	switch {
	case hour >= 5 && hour < 12:
		return "Good morning"
	case hour >= 12 && hour < 17:
		return "Good afternoon"
	case hour >= 17 && hour < 22:
		return "Good evening"
	default:
		return "Hello"
	}
}

// IsValidBrandInfo checks if brand information is sufficient for branded greetings
func (b *BrandGreetingService) IsValidBrandInfo(brandInfo *BrandInfo) bool {
	return brandInfo != nil && (brandInfo.CompanyName != "" || brandInfo.About != "")
}

// GetGreetingTemplates returns available greeting templates for testing/preview
func (b *BrandGreetingService) GetGreetingTemplates() map[string][]string {
	return map[string][]string{
		"full_brand": {
			"{timeOfDay}! Welcome to {companyName}. {about} How can we help you today?",
			"Hello! Thanks for reaching out to {companyName}. {about} What can we assist you with?",
			"Hi there! You've reached {companyName} support. {about} How may we help you?",
		},
		"company_only": {
			"{timeOfDay}! Welcome to {companyName}. How can we assist you today?",
			"Hello! You've reached {companyName} support. What can we help you with?",
			"Hi there! Thanks for contacting {companyName}. How may we help you?",
		},
		"about_only": {
			"{timeOfDay}! {about} How can we help you today?",
			"Hello! {about} What can we assist you with?",
			"Hi there! {about} How may we help you?",
		},
		"generic": {
			"{timeOfDay}! Welcome to our support. How can we help you today?",
			"Hello! Thanks for reaching out. What can we assist you with?",
			"Hi there! How may we help you today?",
		},
	}
}

// PreviewGreeting generates a preview of what the greeting would look like
func (b *BrandGreetingService) PreviewGreeting(ctx context.Context, tenantID, projectID uuid.UUID) (*GreetingResponse, error) {
	return b.GenerateGreetingResponse(ctx, tenantID, projectID, "Hello")
}
