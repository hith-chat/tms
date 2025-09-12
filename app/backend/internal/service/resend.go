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
	environment string
}

// NewResendService creates a new Resend service
func NewResendService(cfg *config.ResendConfig, environment string) *ResendService {
	apiKey := cfg.APIKey
	if apiKey == "" {
		// Fall back to a test key for development - should be configured properly
		apiKey = "re_123456789"
	}

	fromEmail := cfg.FromEmail
	if fromEmail == "" {
		fromEmail = "noreply@taral.co" // Default fallback
	}
	fmt.Println("Resend api key ->", apiKey)

	client := resend.NewClient(apiKey)
	return &ResendService{
		client:      client,
		fromEmail:   fromEmail,
		environment: environment,
	}
}

// SendSignupVerificationEmail sends a verification email for signup
func (s *ResendService) SendSignupVerificationEmail(ctx context.Context, toEmail, otp string) error {
	if s.environment == "development" {
		fmt.Println("Development mode: skipping email. Here is the OTP:", otp)
		return nil
	}

	fmt.Println("Sending signup verification email to:", toEmail, s.environment, otp)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: "Verify your TMS account",
		Html: fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify your TMS account</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 40px 20px; text-align: center; }
        .content { padding: 40px 20px; }
        .otp-code { background: #f8f9fa; border: 2px solid #e9ecef; border-radius: 8px; padding: 20px; text-align: center; margin: 30px 0; font-size: 32px; font-weight: bold; letter-spacing: 4px; font-family: 'Courier New', monospace; color: #495057; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #6c757d; font-size: 14px; }
        .warning { background: #fff3cd; border: 1px solid #ffeaa7; border-radius: 4px; padding: 15px; margin: 20px 0; color: #856404; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to TMS!</h1>
            <p>Complete your account verification</p>
        </div>
        <div class="content">
            <h2>Hello!</h2>
            <p>Thank you for signing up for TMS. To complete your account setup, please verify your email address using the code below:</p>
            
            <div class="otp-code">%s</div>
            
            <p>Enter this 6-digit code in the verification screen to activate your account.</p>
            
            <div class="warning">
                <strong>Security Note:</strong> This code will expire in 10 minutes for your security. If you didn't create an account, you can safely ignore this email.
            </div>
            
            <p>If you're having trouble, please contact our support team.</p>
            
            <p>Best regards,<br>The TMS Team</p>
        </div>
        <div class="footer">
            <p>This email was sent to %s</p>
            <p>TMS - Ticket Management System</p>
        </div>
    </div>
</body>
</html>
		`, otp, toEmail),
		Text: fmt.Sprintf(`
Welcome to TMS!

Thank you for signing up for TMS. To complete your account setup, please verify your email address using the code below:

Verification Code: %s

Enter this 6-digit code in the verification screen to activate your account.

This code will expire in 10 minutes for your security. If you didn't create an account, you can safely ignore this email.

Best regards,
The TMS Team

This email was sent to %s
TMS - Ticket Management System
		`, otp, toEmail),
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send verification email via Resend: %w", err)
	}

	return nil
}

// SendTicketCreatedNotification sends email notification when a new ticket is created
func (s *ResendService) SendTicketCreatedNotification(ctx context.Context, ticket *db.Ticket, customer *db.Customer, toEmail, recipientName, recipientType string) error {
	if s.environment == "development" {
		fmt.Printf("Development mode: Would send ticket created notification to %s (%s)\n", toEmail, recipientType)
		return nil
	}

	var subject, htmlBody, textBody string

	if recipientType == "tenant_admin" {
		subject = fmt.Sprintf("New Ticket Created: %s", ticket.Subject)
		htmlBody = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Ticket Created</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 40px 20px; text-align: center; }
        .content { padding: 40px 20px; }
        .ticket-info { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 8px; padding: 20px; margin: 20px 0; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #6c757d; font-size: 14px; }
        .btn { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>New Ticket Created</h1>
        </div>
        <div class="content">
            <p>Dear %s,</p>
            <p>A new support ticket has been created in your system.</p>
            
            <div class="ticket-info">
                <h3>Ticket Details</h3>
                <p><strong>Subject:</strong> %s</p>
                <p><strong>Priority:</strong> %s</p>
                <p><strong>Type:</strong> %s</p>
                <p><strong>Customer:</strong> %s (%s)</p>
                <p><strong>Status:</strong> %s</p>
                <p><strong>Source:</strong> %s</p>
            </div>
            
            <p>Please review and assign this ticket to the appropriate agent.</p>
            
            <a href="#" class="btn">View Ticket</a>
        </div>
        <div class="footer">
            <p>This notification was sent to %s</p>
            <p>TMS - Ticket Management System</p>
        </div>
    </div>
</body>
</html>
		`, recipientName, ticket.Subject, ticket.Priority, ticket.Type, customer.Name, customer.Email, ticket.Status, ticket.Source, toEmail)

		textBody = fmt.Sprintf(`
New Ticket Created

Dear %s,

A new support ticket has been created in your system.

Ticket Details:
- Subject: %s
- Priority: %s
- Type: %s
- Customer: %s (%s)
- Status: %s
- Source: %s

Please review and assign this ticket to the appropriate agent.

Best regards,
TMS Team

This notification was sent to %s
TMS - Ticket Management System
		`, recipientName, ticket.Subject, ticket.Priority, ticket.Type, customer.Name, customer.Email, ticket.Status, ticket.Source, toEmail)
	} else {
		// Customer notification
		subject = fmt.Sprintf("Your support ticket has been created: %s", ticket.Subject)
		htmlBody = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Support Ticket Created</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 40px 20px; text-align: center; }
        .content { padding: 40px 20px; }
        .ticket-info { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 8px; padding: 20px; margin: 20px 0; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #6c757d; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Support Ticket Created</h1>
        </div>
        <div class="content">
            <p>Dear %s,</p>
            <p>Thank you for contacting us. Your support ticket has been successfully created and assigned a priority level of <strong>%s</strong>.</p>
            
            <div class="ticket-info">
                <h3>Your Ticket Details</h3>
                <p><strong>Subject:</strong> %s</p>
                <p><strong>Priority:</strong> %s</p>
                <p><strong>Status:</strong> %s</p>
            </div>
            
            <p>Our support team will review your request and respond according to the priority level. You will receive email updates when there are any changes to your ticket.</p>
            
            <p>If you need to add additional information, please reply to this email.</p>
        </div>
        <div class="footer">
            <p>This email was sent to %s</p>
            <p>TMS - Ticket Management System</p>
        </div>
    </div>
</body>
</html>
		`, customer.Name, ticket.Priority, ticket.Subject, ticket.Priority, ticket.Status, customer.Email)

		textBody = fmt.Sprintf(`
Support Ticket Created

Dear %s,

Thank you for contacting us. Your support ticket has been successfully created and assigned a priority level of %s.

Your Ticket Details:
- Subject: %s
- Priority: %s
- Status: %s

Our support team will review your request and respond according to the priority level. You will receive email updates when there are any changes to your ticket.

If you need to add additional information, please reply to this email.

Best regards,
TMS Team

This email was sent to %s
TMS - Ticket Management System
		`, customer.Name, ticket.Priority, ticket.Subject, ticket.Priority, ticket.Status, customer.Email)
	}

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
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

	subject := fmt.Sprintf("Ticket Updated: %s", ticket.Subject)
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ticket Updated</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 40px 20px; text-align: center; }
        .content { padding: 40px 20px; }
        .update-info { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 8px; padding: 20px; margin: 20px 0; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #6c757d; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Ticket Updated</h1>
        </div>
        <div class="content">
            <p>Dear %s,</p>
            <p>Your support ticket has been updated.</p>
            
            <div class="update-info">
                <h3>Update Details</h3>
                <p><strong>Ticket Subject:</strong> %s</p>
                <p><strong>Update Type:</strong> %s</p>
                <p><strong>Current Status:</strong> %s</p>
                <p><strong>Details:</strong> %s</p>
            </div>
            
            <p>If you have any questions or need to provide additional information, please reply to this email.</p>
        </div>
        <div class="footer">
            <p>This email was sent to %s</p>
            <p>TMS - Ticket Management System</p>
        </div>
    </div>
</body>
</html>
	`, recipientName, ticket.Subject, updateType, ticket.Status, updateDetails, toEmail)

	textBody := fmt.Sprintf(`
Ticket Updated

Dear %s,

Your support ticket has been updated.

Update Details:
- Ticket Subject: %s
- Update Type: %s
- Current Status: %s
- Details: %s

If you have any questions or need to provide additional information, please reply to this email.

Best regards,
TMS Team

This email was sent to %s
TMS - Ticket Management System
	`, recipientName, ticket.Subject, updateType, ticket.Status, updateDetails, toEmail)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
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
