package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bareuptime/tms/internal/models"
)

// MockIntegrationRepo is a mock for ProjectIntegrationRepository
type MockIntegrationRepo struct {
	mock.Mock
}

func (m *MockIntegrationRepo) GetByProjectAndType(ctx context.Context, tenantID, projectID uuid.UUID, integrationType models.IntegrationType) (*models.Integration, error) {
	args := m.Called(ctx, tenantID, projectID, integrationType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Integration), args.Error(1)
}

// Implement other methods of the interface if needed, or embed the real repo struct if it's not an interface
// Since ProjectIntegrationRepository is a struct in the original code, we might need to interface it or just mock the behavior if we can't easily mock the struct.
// However, for this unit test, we can just create a test server and test the PostMessageToTeams logic if we can inject the repo.
// The service takes *repo.ProjectIntegrationRepository. If it's a struct, we can't mock it easily without an interface.
// Let's check if we can just test the HTTP request part by mocking the integration response if we can't mock the repo.
// Wait, the service uses the repo to get the integration. If I can't mock the repo, I can't test GetTeamsIntegration easily without a DB.
// But I can verify the HTTP request if I manually construct the service or if I can mock the repo.
// Let's assume for now I can't easily mock the repo struct without refactoring.
// I will create a test that focuses on the HTTP request part if I can, but the method calls GetTeamsIntegration first.
//
// Alternative: I will create a new test file that defines an interface for the repo if one doesn't exist, or I'll just skip the repo part and test the logic if I extract it.
// But the method is coupled.
//
// Let's look at the service code again.
// type MicrosoftTeamsService struct {
// 	integrationRepo *repo.ProjectIntegrationRepository
// 	httpClient      *http.Client
// }
//
// I can't mock integrationRepo if it's a concrete struct.
// I will create a test that assumes the repo returns a specific integration.
// Since I can't mock the struct method easily in Go without an interface, I might need to refactor to use an interface or just write a test that doesn't rely on the repo (e.g. by testing a helper method if I extracted one).
//
// Actually, I can just create a test that verifies the payload structure by creating a separate function that generates the payload, or I can try to run the code if I can set up a real DB.
// But I don't have a running DB for tests here.
//
// I will create a test that verifies the `TeamsMessageCard` struct and JSON marshaling, which is the core logic I added.

func TestTeamsMessageCard_Marshal(t *testing.T) {
	card := TeamsMessageCard{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: "0076D7",
		Summary:    "New Chat Message",
		Sections: []TeamsSection{
			{
				ActivityTitle:    "John Doe",
				ActivitySubtitle: "Session: 123",
				Text:             "Hello world",
			},
		},
	}

	data, err := json.Marshal(card)
	assert.NoError(t, err)

	expected := `{"@type":"MessageCard","@context":"http://schema.org/extensions","themeColor":"0076D7","summary":"New Chat Message","sections":[{"activityTitle":"John Doe","activitySubtitle":"Session: 123","text":"Hello world"}]}`
	assert.JSONEq(t, expected, string(data))
}
