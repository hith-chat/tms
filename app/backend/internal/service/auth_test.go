package service

import (
	"testing"
	"time"

	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIsValidCorporateEmail(t *testing.T) {
	t.Parallel()

	service := &AuthService{}

	require.NoError(t, service.isValidCorporateEmail("user@company.com"))
	require.Error(t, service.isValidCorporateEmail("user@gmail.com"))
	require.Error(t, service.isValidCorporateEmail("user@temp-mail.org"))
	require.Error(t, service.isValidCorporateEmail("invalid-email"))
}

func TestConvertRoleBindings(t *testing.T) {
	t.Parallel()

	projectID := uuid.New()
	now := time.Now()

	bindings := []*db.RoleBinding{
		{
			AgentID:   uuid.New(),
			TenantID:  uuid.New(),
			ProjectID: &projectID,
			Role:      models.RoleAgent,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			AgentID:   uuid.New(),
			TenantID:  uuid.New(),
			ProjectID: nil,
			Role:      models.RoleTenantAdmin,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	service := &AuthService{}
	result := service.convertRoleBindings(bindings)

	require.Len(t, result, 2)
	require.ElementsMatch(t, []string{models.RoleAgent.String()}, result[projectID.String()])
	require.ElementsMatch(t, []string{models.RoleTenantAdmin.String()}, result[""])
}
