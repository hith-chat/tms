package service

import (
	"fmt"
	"strings"

	"github.com/bareuptime/tms/internal/config"
)

// NewEmailProvider builds the concrete email provider based on configuration.
func NewEmailProvider(emailCfg *config.EmailConfig, resendCfg *config.ResendConfig, mailerooCfg *config.MailerooConfig, environment string) (EmailProvider, error) {
	provider := strings.ToLower(emailCfg.Provider)
	switch provider {
	case "", "resend":
		return NewResendService(resendCfg, environment), nil
	case "maileroo":
		return NewMailerooService(mailerooCfg, environment)
	default:
		return nil, fmt.Errorf("unsupported email provider %q", provider)
	}
}
