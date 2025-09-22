package service

import (
	"context"

	"github.com/bareuptime/tms/internal/db"
)

// EmailProvider defines the contract for sending transactional emails used in the application.
type EmailProvider interface {
	SendSignupVerificationEmail(ctx context.Context, toEmail, otp string) error
	SendTicketCreatedNotification(ctx context.Context, ticket *db.Ticket, customer *db.Customer, toEmail, recipientName, recipientType string) error
	SendTicketUpdatedNotification(ctx context.Context, ticket *db.Ticket, customer *db.Customer, toEmail, recipientName, updateType, updateDetails string) error
}
