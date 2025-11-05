package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/mail"
	"github.com/bareuptime/tms/internal/models"
)

type DomainNameService struct {
	domainRepo   DomainValidationRepository
	emailService *mail.Service
	dnsLookup    func(string) ([]string, error)
}

func NewDomainValidationService(domainRepo DomainValidationRepository, emailService *mail.Service) *DomainNameService {
	return &DomainNameService{
		domainRepo:   domainRepo,
		emailService: emailService,
		dnsLookup:    net.LookupTXT,
	}
}

// DomainValidationRepository captures the data access needs for domain validation workflows.
type DomainValidationRepository interface {
	CreateDomainValidation(ctx context.Context, validation *models.EmailDomain) error
	GetDomainByID(ctx context.Context, tenantID uuid.UUID, validationID uuid.UUID) (*models.EmailDomain, error)
	UpdateDomainValidation(ctx context.Context, validation *models.EmailDomain) error
	ListDomainNames(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.EmailDomain, error)
	DeleteDomainName(ctx context.Context, tenantID, projectID, domainID uuid.UUID) error
}

// CreateDomainValidation initiates domain validation process
func (s *DomainNameService) CreateDomainValidation(ctx context.Context, tenantID, projectID uuid.UUID, domain string) (*models.EmailDomain, error) {
	// Validate domain format
	if err := validateDomainFormat(domain); err != nil {
		return nil, fmt.Errorf("invalid domain format: %w", err)
	}

	// Generate validation token based on method
	var token string
	var err error
	token, err = generateDNSToken()

	if err != nil {
		return nil, fmt.Errorf("failed to generate validation token: %w", err)
	}

	// Create validation record
	validation := &models.EmailDomain{
		ID:              uuid.New(),
		TenantID:        tenantID,
		ProjectID:       projectID,
		Domain:          strings.ToLower(domain),
		ValidationToken: token,
		Status:          models.DomainValidationStatusPending,
		ExpiresAt:       time.Now().Add(24 * time.Hour), // 24 hours expiry
		Metadata:        make(models.JSONMap),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Set method-specific metadata

	validation.Metadata["dns_record"] = fmt.Sprintf("_tms-validation.%s", domain)
	validation.Metadata["dns_value"] = token

	err = s.domainRepo.CreateDomainValidation(ctx, validation)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain validation: %w", err)
	}

	return validation, nil
}

// VerifyDomainName verifies domain ownership using the provided token/proof
func (s *DomainNameService) VerifyDomainName(ctx context.Context, tenantID uuid.UUID, validationID uuid.UUID, proof string) error {
	validation, err := s.domainRepo.GetDomainByID(ctx, tenantID, validationID)
	if err != nil {
		return fmt.Errorf("failed to get domain validation: %w", err)
	}

	if validation == nil {
		return fmt.Errorf("domain validation not found")
	}

	if validation.Status == models.DomainValidationStatusVerified {
		return fmt.Errorf("domain is already verified")
	}

	if time.Now().After(validation.ExpiresAt) {
		validation.Status = models.DomainValidationStatusExpired
		s.domainRepo.UpdateDomainValidation(ctx, validation)
		return fmt.Errorf("validation has expired")
	}

	verified, verifyErr := s.verifyDNSRecord(validation.Domain, validation.ValidationToken)

	if verifyErr != nil {
		validation.Status = models.DomainValidationStatusFailed
		validation.Metadata["error"] = verifyErr.Error()
		s.domainRepo.UpdateDomainValidation(ctx, validation)
		return fmt.Errorf("verification failed: %w", verifyErr)
	}

	if verified {
		now := time.Now()
		validation.Status = models.DomainValidationStatusVerified
		validation.VerifiedAt = &now
		err = s.domainRepo.UpdateDomainValidation(ctx, validation)
		if err != nil {
			return fmt.Errorf("failed to update validation status: %w", err)
		}
		return nil
	} else {
		validation.Status = models.DomainValidationStatusFailed
		validation.Metadata["error"] = "Invalid verification proof provided"
		s.domainRepo.UpdateDomainValidation(ctx, validation)
		return fmt.Errorf("invalid verification proof")
	}
}

// GetDomainNames lists all domain validations for a project
func (s *DomainNameService) GetDomainNames(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.EmailDomain, error) {
	return s.domainRepo.ListDomainNames(ctx, tenantID, projectID)
}

// DeleteDomainName deletes a domain validation
func (s *DomainNameService) DeleteDomainName(ctx context.Context, tenantID, projectID, domainID uuid.UUID) error {
	return s.domainRepo.DeleteDomainName(ctx, tenantID, projectID, domainID)
}

// Helper functions
func generateDNSToken() (string, error) {
	// Generate a random DNS-safe token
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 32)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[num.Int64()]
	}

	return string(result), nil
}

func validateDomainFormat(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Basic domain validation
	if len(domain) > 253 {
		return fmt.Errorf("domain too long")
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return fmt.Errorf("domain must have at least two parts")
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return fmt.Errorf("invalid domain part length")
		}
	}

	return nil
}

func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

func (s *DomainNameService) verifyDNSRecord(domain, expectedValue string) (bool, error) {
	recordName := fmt.Sprintf("_tms-validation.%s", domain)

	lookup := s.dnsLookup
	if lookup == nil {
		lookup = net.LookupTXT
	}

	txtRecords, err := lookup(recordName)
	if err != nil {
		return false, fmt.Errorf("failed to lookup DNS TXT record: %w", err)
	}

	for _, record := range txtRecords {
		if record == expectedValue {
			return true, nil
		}
	}

	return false, fmt.Errorf("DNS TXT record not found or does not match expected value")
}
