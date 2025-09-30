package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBrandSettingsRepo struct {
	mock.Mock
}

func (m *mockBrandSettingsRepo) GetSetting(ctx context.Context, tenantID, projectID uuid.UUID, settingKey string) (map[string]interface{}, int, error) {
	args := m.Called(ctx, tenantID, projectID, settingKey)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).(map[string]interface{}), args.Int(1), args.Error(2)
}

func TestBrandGreetingService_GenerateGreetingResponse_WithBrandInfo(t *testing.T) {
	ctx := context.Background()
	repoMock := &mockBrandSettingsRepo{}

	tenantID := uuid.New()
	projectID := uuid.New()

	settings := map[string]interface{}{
		"company_name": "Acme Rockets",
		"about":        "We make space travel delightful.",
		"support_url":  "https://support.acme.test",
	}

	repoMock.On("GetSetting", mock.Anything, tenantID, projectID, "branding_settings").
		Return(settings, 200, nil).Once()

	service := NewBrandGreetingService(repoMock)

	response, err := service.GenerateGreetingResponse(ctx, tenantID, projectID, "Hello")
	assert.NoError(t, err)
	assert.Equal(t, "branded_greeting", response.Template)
	assert.Equal(t, settings["company_name"], response.BrandInfo.CompanyName)
	assert.Equal(t, settings["about"], response.BrandInfo.About)
	assert.Contains(t, response.Message, "Acme Rockets")
}

func TestBrandGreetingService_GenerateGreetingResponse_DefaultOnError(t *testing.T) {
	ctx := context.Background()
	repoMock := &mockBrandSettingsRepo{}

	tenantID := uuid.New()
	projectID := uuid.New()

	repoMock.On("GetSetting", mock.Anything, tenantID, projectID, "branding_settings").
		Return(nil, 500, errors.New("db error")).Once()

	service := NewBrandGreetingService(repoMock)

	response, err := service.GenerateGreetingResponse(ctx, tenantID, projectID, "Hello")
	assert.NoError(t, err)
	assert.Equal(t, "default_greeting", response.Template)
	assert.Empty(t, response.BrandInfo.CompanyName)
	assert.Contains(t, response.Message, "Thanks for reaching out")
}

func TestBrandGreetingService_IsValidBrandInfo(t *testing.T) {
	repoMock := &mockBrandSettingsRepo{}
	service := NewBrandGreetingService(repoMock)

	assert.False(t, service.IsValidBrandInfo(nil))
	assert.False(t, service.IsValidBrandInfo(&BrandInfo{}))
	assert.True(t, service.IsValidBrandInfo(&BrandInfo{CompanyName: "Acme"}))
	assert.True(t, service.IsValidBrandInfo(&BrandInfo{About: "We help"}))
}

func TestBrandGreetingService_GetTimeOfDay(t *testing.T) {
	repoMock := &mockBrandSettingsRepo{}
	service := NewBrandGreetingService(repoMock)

	result := service.getTimeOfDay()
	valid := map[string]bool{
		"Good morning":   true,
		"Good afternoon": true,
		"Good evening":   true,
		"Hello":          true,
	}

	assert.True(t, valid[result], "unexpected time of day: %s", result)
}

func TestBrandGreetingService_GenerateGenericGreeting(t *testing.T) {
	repoMock := &mockBrandSettingsRepo{}
	service := NewBrandGreetingService(repoMock)

	// force deterministic selection by stubbing time via repeated calls
	greeting1 := service.generateGenericGreeting("Hello")
	time.Sleep(time.Millisecond)
	greeting2 := service.generateGenericGreeting("Hello")

	assert.NotEmpty(t, greeting1)
	assert.NotEmpty(t, greeting2)
}
