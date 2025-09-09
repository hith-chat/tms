package service

import (
	"context"
	"fmt"

	"github.com/bareuptime/tms/internal/config"
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
