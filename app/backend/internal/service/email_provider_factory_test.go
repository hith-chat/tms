package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bareuptime/tms/internal/config"
)

func TestNewEmailProvider_DefaultsToResend(t *testing.T) {
	emailCfg := &config.EmailConfig{Provider: ""}
	resendCfg := &config.ResendConfig{APIKey: "test", FromEmail: "from@test.com", FromName: "Test"}
	mailerooCfg := &config.MailerooConfig{APIKey: "maileroo-key", FromEmail: "from@test.com", FromName: "Test", TimeoutSeconds: 10}

	provider, err := NewEmailProvider(emailCfg, resendCfg, mailerooCfg, "production")

	assert.NoError(t, err)
	assert.IsType(t, &ResendService{}, provider)
}

func TestNewEmailProvider_ReturnsMaileroo(t *testing.T) {
	emailCfg := &config.EmailConfig{Provider: "maileroo"}
	resendCfg := &config.ResendConfig{}
	mailerooCfg := &config.MailerooConfig{APIKey: "maileroo-key", FromEmail: "from@test.com", FromName: "Test", TimeoutSeconds: 10}

	provider, err := NewEmailProvider(emailCfg, resendCfg, mailerooCfg, "development")

	assert.NoError(t, err)
	assert.IsType(t, &MailerooService{}, provider)
}

func TestNewEmailProvider_Unsupported(t *testing.T) {
	emailCfg := &config.EmailConfig{Provider: "unknown"}
	resendCfg := &config.ResendConfig{}
	mailerooCfg := &config.MailerooConfig{}

	provider, err := NewEmailProvider(emailCfg, resendCfg, mailerooCfg, "development")

	assert.Error(t, err)
	assert.Nil(t, provider)
}
