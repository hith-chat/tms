package service

import (
	"context"
	"fmt"

	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/db"
	maileroo "github.com/maileroo/maileroo-go-sdk/maileroo"
)

// MailerooService handles email sending via the Maileroo provider.
type MailerooService struct {
	client      *maileroo.Client
	fromEmail   string
	fromName    string
	environment string
}

var _ EmailProvider = (*MailerooService)(nil)

// NewMailerooService creates a new Maileroo email provider.
func NewMailerooService(cfg *config.MailerooConfig, environment string) (*MailerooService, error) {
	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 30
	}

	apiKey := cfg.APIKey
	if apiKey == "" {
		return nil, fmt.Errorf("maileroo api key is not configured")
	}

	fromEmail := cfg.FromEmail
	if fromEmail == "" {
		fromEmail = "noreply@hith.support"
	}

	fromName := cfg.FromName
	if fromName == "" {
		fromName = "TMS"
	}

	client, err := maileroo.NewClient(apiKey, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create Maileroo client: %w", err)
	}

	return &MailerooService{
		client:      client,
		fromEmail:   fromEmail,
		fromName:    fromName,
		environment: environment,
	}, nil
}

func (s *MailerooService) newSender(email, name string) maileroo.EmailAddress {
	if email != "" && name != "" {
		return maileroo.NewEmail(email, name)
	}
	return maileroo.NewEmail(s.fromEmail, s.fromName)
}

func (s *MailerooService) newRecipient(email, name string) maileroo.EmailAddress {
	if name == "" {
		name = email
	}
	return maileroo.NewEmail(email, name)
}

// SendSignupVerificationEmail sends a verification email for signup.
func (s *MailerooService) SendSignupVerificationEmail(ctx context.Context, toEmail, otp string) error {
	if s.environment == "development" {
		fmt.Println("Development mode: skipping email. Here is the OTP:", otp)
		return nil
	}

	fmt.Println("Sending signup verification email via Maileroo to:", toEmail, s.environment, otp)

	subject, htmlBody, textBody := buildSignupVerificationEmail(toEmail, otp)
	html := htmlBody
	text := textBody

	_, err := s.client.SendBasicEmail(ctx, maileroo.BasicEmailData{
		From:    s.newSender("", ""),
		To:      []maileroo.EmailAddress{s.newRecipient(toEmail, "")},
		Subject: subject,
		HTML:    &html,
		Plain:   &text,
	})
	if err != nil {
		return fmt.Errorf("failed to send verification email via Maileroo: %w", err)
	}

	return nil
}

// SendSignupWelcomeEmail sends a welcome email after successful signup verification.
func (s *MailerooService) SendSignupWelcomeEmail(ctx context.Context, toEmail, recipientName string) error {
	if s.environment == "development" {
		fmt.Printf("Development mode: would send signup welcome email via Maileroo to %s\n", toEmail)
		return nil
	}

	fmt.Printf("Sending signup welcome email via Maileroo to: %s\n", toEmail)

	subject, htmlBody, textBody := buildSignupWelcomeEmail(toEmail, recipientName)
	html := htmlBody
	text := textBody

	_, err := s.client.SendBasicEmail(ctx, maileroo.BasicEmailData{
		From:    s.newSender("", ""),
		To:      []maileroo.EmailAddress{s.newRecipient(toEmail, recipientName)},
		Subject: subject,
		HTML:    &html,
		Plain:   &text,
	})
	if err != nil {
		return fmt.Errorf("failed to send signup welcome email via Maileroo: %w", err)
	}

	return nil
}

// SendTicketCreatedNotification sends email notification when a new ticket is created.
func (s *MailerooService) SendTicketCreatedNotification(ctx context.Context, ticket *db.Ticket, customer *db.Customer, toEmail, recipientName, recipientType string) error {
	if s.environment == "development" {
		fmt.Printf("Development mode: Would send ticket created notification via Maileroo to %s (%s)\n", toEmail, recipientType)
		return nil
	}

	subject, htmlBody, textBody := buildTicketCreatedEmail(ticket, customer, recipientName, recipientType, toEmail)
	html := htmlBody
	text := textBody

	_, err := s.client.SendBasicEmail(ctx, maileroo.BasicEmailData{
		From:    s.newSender("", ""),
		To:      []maileroo.EmailAddress{s.newRecipient(toEmail, recipientName)},
		Subject: subject,
		HTML:    &html,
		Plain:   &text,
	})
	if err != nil {
		return fmt.Errorf("failed to send ticket created notification via Maileroo: %w", err)
	}

	return nil
}

// SendTicketUpdatedNotification sends email notification when a ticket is updated.
func (s *MailerooService) SendTicketUpdatedNotification(ctx context.Context, ticket *db.Ticket, customer *db.Customer, toEmail, recipientName, updateType, updateDetails string) error {
	if s.environment == "development" {
		fmt.Printf("Development mode: Would send ticket updated notification via Maileroo to %s\n", toEmail)
		return nil
	}

	subject, htmlBody, textBody := buildTicketUpdatedEmail(ticket, recipientName, updateType, updateDetails, toEmail)
	html := htmlBody
	text := textBody

	_, err := s.client.SendBasicEmail(ctx, maileroo.BasicEmailData{
		From:    s.newSender("", ""),
		To:      []maileroo.EmailAddress{s.newRecipient(toEmail, recipientName)},
		Subject: subject,
		HTML:    &html,
		Plain:   &text,
	})
	if err != nil {
		return fmt.Errorf("failed to send ticket updated notification via Maileroo: %w", err)
	}

	return nil
}
