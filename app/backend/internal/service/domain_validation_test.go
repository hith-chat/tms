package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bareuptime/tms/internal/models"
)

type mockDomainValidationRepo struct {
	mock.Mock
}

func (m *mockDomainValidationRepo) CreateDomainValidation(ctx context.Context, validation *models.EmailDomain) error {
	args := m.Called(ctx, validation)
	return args.Error(0)
}

func (m *mockDomainValidationRepo) GetDomainByID(ctx context.Context, tenantID, validationID uuid.UUID) (*models.EmailDomain, error) {
	args := m.Called(ctx, tenantID, validationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EmailDomain), args.Error(1)
}

func (m *mockDomainValidationRepo) UpdateDomainValidation(ctx context.Context, validation *models.EmailDomain) error {
	args := m.Called(ctx, validation)
	return args.Error(0)
}

func (m *mockDomainValidationRepo) ListDomainNames(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.EmailDomain, error) {
	args := m.Called(ctx, tenantID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.EmailDomain), args.Error(1)
}

func (m *mockDomainValidationRepo) DeleteDomainName(ctx context.Context, tenantID, projectID, domainID uuid.UUID) error {
	args := m.Called(ctx, tenantID, projectID, domainID)
	return args.Error(0)
}

func TestDomainNameService_CreateDomainValidation(t *testing.T) {
	repoMock := &mockDomainValidationRepo{}
	service := NewDomainValidationService(repoMock, nil)

	tenantID := uuid.New()
	projectID := uuid.New()

	repoMock.On("CreateDomainValidation", mock.Anything, mock.MatchedBy(func(validation *models.EmailDomain) bool {
		return validation.Domain == "example.com" && len(validation.ValidationToken) == 32 && validation.Metadata["dns_record"] != ""
	})).Return(nil).Once()

	result, err := service.CreateDomainValidation(context.Background(), tenantID, projectID, "example.com")

	assert.NoError(t, err)
	assert.Equal(t, "example.com", result.Domain)
	assert.Len(t, result.ValidationToken, 32)
	repoMock.AssertExpectations(t)
}

func TestDomainNameService_CreateDomainValidation_InvalidDomain(t *testing.T) {
	repoMock := &mockDomainValidationRepo{}
	service := NewDomainValidationService(repoMock, nil)

	_, err := service.CreateDomainValidation(context.Background(), uuid.New(), uuid.New(), "invalid")

	assert.Error(t, err)
	repoMock.AssertNotCalled(t, "CreateDomainValidation", mock.Anything, mock.Anything)
}

func TestDomainNameService_VerifyDomainName_Success(t *testing.T) {
	repoMock := &mockDomainValidationRepo{}
	service := NewDomainValidationService(repoMock, nil)

	tenantID := uuid.New()
	validationID := uuid.New()

	validation := &models.EmailDomain{
		ID:              validationID,
		TenantID:        tenantID,
		ProjectID:       uuid.New(),
		Domain:          "example.com",
		ValidationToken: "token123",
		Status:          models.DomainValidationStatusPending,
		ExpiresAt:       time.Now().Add(time.Hour),
		Metadata:        models.JSONMap{},
	}

	repoMock.On("GetDomainByID", mock.Anything, tenantID, validationID).Return(validation, nil).Once()
	repoMock.On("UpdateDomainValidation", mock.Anything, mock.MatchedBy(func(updated *models.EmailDomain) bool {
		return updated.Status == models.DomainValidationStatusVerified && updated.VerifiedAt != nil
	})).Return(nil).Once()

	service.dnsLookup = func(name string) ([]string, error) {
		assert.Equal(t, "_tms-validation.example.com", name)
		return []string{"token123"}, nil
	}

	err := service.VerifyDomainName(context.Background(), tenantID, validationID, "ignored")
	assert.NoError(t, err)
	repoMock.AssertExpectations(t)
}

func TestDomainNameService_VerifyDomainName_Failure(t *testing.T) {
	repoMock := &mockDomainValidationRepo{}
	service := NewDomainValidationService(repoMock, nil)

	tenantID := uuid.New()
	validationID := uuid.New()

	validation := &models.EmailDomain{
		ID:              validationID,
		TenantID:        tenantID,
		Domain:          "example.com",
		ValidationToken: "token123",
		Status:          models.DomainValidationStatusPending,
		ExpiresAt:       time.Now().Add(time.Hour),
		Metadata:        models.JSONMap{},
	}

	repoMock.On("GetDomainByID", mock.Anything, tenantID, validationID).Return(validation, nil).Once()
	repoMock.On("UpdateDomainValidation", mock.Anything, mock.MatchedBy(func(updated *models.EmailDomain) bool {
		return updated.Status == models.DomainValidationStatusFailed
	})).Return(nil).Once()

	service.dnsLookup = func(name string) ([]string, error) {
		return nil, errors.New("dns mismatch")
	}

	err := service.VerifyDomainName(context.Background(), tenantID, validationID, "wrong")
	assert.Error(t, err)
	repoMock.AssertExpectations(t)
}

func TestDomainNameService_VerifyDomainName_Expired(t *testing.T) {
	repoMock := &mockDomainValidationRepo{}
	service := NewDomainValidationService(repoMock, nil)

	tenantID := uuid.New()
	validationID := uuid.New()

	validation := &models.EmailDomain{
		ID:        validationID,
		TenantID:  tenantID,
		Domain:    "example.com",
		Status:    models.DomainValidationStatusPending,
		ExpiresAt: time.Now().Add(-time.Hour),
		Metadata:  models.JSONMap{},
	}

	repoMock.On("GetDomainByID", mock.Anything, tenantID, validationID).Return(validation, nil).Once()
	repoMock.On("UpdateDomainValidation", mock.Anything, mock.MatchedBy(func(updated *models.EmailDomain) bool {
		return updated.Status == models.DomainValidationStatusExpired
	})).Return(nil).Once()

	err := service.VerifyDomainName(context.Background(), tenantID, validationID, "ignored")
	assert.Error(t, err)
	repoMock.AssertExpectations(t)
}

func TestDomainNameService_GetAndDelete(t *testing.T) {
	repoMock := &mockDomainValidationRepo{}
	service := NewDomainValidationService(repoMock, nil)

	tenantID := uuid.New()
	projectID := uuid.New()
	validationID := uuid.New()

	repoMock.On("ListDomainNames", mock.Anything, tenantID, projectID).Return([]*models.EmailDomain{}, nil).Once()
	repoMock.On("DeleteDomainName", mock.Anything, tenantID, projectID, validationID).Return(nil).Once()

	_, err := service.GetDomainNames(context.Background(), tenantID, projectID)
	assert.NoError(t, err)

	err = service.DeleteDomainName(context.Background(), tenantID, projectID, validationID)
	assert.NoError(t, err)

	repoMock.AssertExpectations(t)
}
