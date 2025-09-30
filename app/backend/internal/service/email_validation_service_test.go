package service

import (
	"context"
	"testing"

	"github.com/bareuptime/tms/internal/models"
	"github.com/stretchr/testify/require"
)

func TestExtractDomain(t *testing.T) {
	t.Parallel()

	domain, err := extractDomain("User@Example.com")
	require.NoError(t, err)
	require.Equal(t, "example.com", domain)

	_, err = extractDomain("invalid-email")
	require.Error(t, err)
}

func TestEmailValidationService_ValidateEmailConfiguration_SMTP(t *testing.T) {
	t.Parallel()

	svc := &EmailValidationService{}
	host := "smtp.example.com"
	port := 587
	username := "user@example.com"

	connector := &models.EmailConnector{
		Type:         models.ConnectorTypeOutboundSMTP,
		SMTPHost:     &host,
		SMTPPort:     &port,
		SMTPUsername: &username,
	}

	err := svc.ValidateEmailConfiguration(context.Background(), connector)
	require.NoError(t, err)
}

func TestEmailValidationService_ValidateEmailConfiguration_InvalidIMAPPort(t *testing.T) {
	t.Parallel()

	svc := &EmailValidationService{}
	smtpHost := "smtp.example.com"
	smtpPort := 587
	smtpUsername := "user@example.com"
	imapHost := "imap.example.com"
	imapPort := 999
	imapUsername := "imap@example.com"

	connector := &models.EmailConnector{
		Type:         models.ConnectorTypeInboundIMAP,
		SMTPHost:     &smtpHost,
		SMTPPort:     &smtpPort,
		SMTPUsername: &smtpUsername,
		IMAPHost:     &imapHost,
		IMAPPort:     &imapPort,
		IMAPUsername: &imapUsername,
	}

	err := svc.ValidateEmailConfiguration(context.Background(), connector)
	require.Error(t, err)
	require.Contains(t, err.Error(), "IMAP port")
}
