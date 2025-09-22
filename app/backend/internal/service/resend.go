package service

import (
	"context"
	"fmt"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/db"
	"github.com/resend/resend-go/v2"
)

// ResendService handles email sending via Resend
type ResendService struct {
	client      *resend.Client
	fromEmail   string
	fromName    string
	environment string
}

var _ EmailProvider = (*ResendService)(nil)

// NewResendService creates a new Resend service
func NewResendService(cfg *config.ResendConfig, environment string) *ResendService {
	apiKey := cfg.APIKey
	if apiKey == "" {
		// Fall back to a test key for development - should be configured properly
		apiKey = "re_123456789"
	}

	fromEmail := cfg.FromEmail
	if fromEmail == "" {
		fromEmail = "noreply@hith.support" // Default fallback
	}

	fromName := cfg.FromName
	if fromName == "" {
		fromName = "TMS"
	}

	client := resend.NewClient(apiKey)
	return &ResendService{
		client:      client,
		fromEmail:   fromEmail,
		fromName:    fromName,
		environment: environment,
	}
}

func (s *ResendService) senderAddress(fromName, fromEmail string) string {

	if fromName != "" && fromEmail != "" {
		return fmt.Sprintf("%s <%s>", fromName, fromEmail)
	}

	if s.fromName != "" {
		return fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	}
	return s.fromEmail
}

// SendSignupVerificationEmail sends a verification email for signup
func (s *ResendService) SendSignupVerificationEmail(ctx context.Context, toEmail, otp string) error {
	if s.environment == "development" {
		fmt.Println("Development mode: skipping email. Here is the OTP:", otp)
		return nil
	}

	fmt.Println("Sending signup verification email to:", toEmail, s.environment, otp)

	subject, htmlBody, textBody := buildSignupVerificationEmail(toEmail, otp)
	params := &resend.SendEmailRequest{
		From:    s.senderAddress("", ""),
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send verification email via Resend: %w", err)
	}

	return nil
}

// SendSignupWelcomeEmail sends a welcome email after successful signup verification
func (s *ResendService) SendSignupWelcomeEmail(ctx context.Context, toEmail, recipientName string) error {
	if s.environment == "development" {
		fmt.Printf("Development mode: would send signup welcome email to %s\n", toEmail)
		return nil
	}

	fmt.Printf("Sending signup welcome email to: %s\n", toEmail)

	subject, htmlBody, textBody := buildSignupWelcomeEmail(toEmail, recipientName)
	params := &resend.SendEmailRequest{
		From:    s.senderAddress("", ""),
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send signup welcome email via Resend: %w", err)
	}

	return nil
}

// SendTicketCreatedNotification sends email notification when a new ticket is created
func (s *ResendService) SendTicketCreatedNotification(ctx context.Context, ticket *db.Ticket, customer *db.Customer, toEmail, recipientName, recipientType string) error {
	if s.environment == "development" {
		fmt.Printf("Development mode: Would send ticket created notification to %s (%s)\n", toEmail, recipientType)
		return nil
	}

	subject, htmlBody, textBody := buildTicketCreatedEmail(ticket, customer, recipientName, recipientType, toEmail)
	params := &resend.SendEmailRequest{
		From:    s.senderAddress("", ""),
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send ticket created notification via Resend: %w", err)
	}

	return nil
}

// SendTicketUpdatedNotification sends email notification when a ticket is updated
func (s *ResendService) SendTicketUpdatedNotification(ctx context.Context, ticket *db.Ticket, customer *db.Customer, toEmail, recipientName, updateType, updateDetails string) error {
	if s.environment == "development" {
		fmt.Printf("Development mode: Would send ticket updated notification to %s\n", toEmail)
		return nil
	}

	subject, htmlBody, textBody := buildTicketUpdatedEmail(ticket, recipientName, updateType, updateDetails, toEmail)
	params := &resend.SendEmailRequest{
		From:    s.senderAddress("", ""),
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send ticket updated notification via Resend: %w", err)
	}

	return nil
}
