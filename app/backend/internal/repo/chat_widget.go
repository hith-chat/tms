package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/bareuptime/tms/internal/models"
)

type ChatWidgetRepo struct {
	db *sqlx.DB
}

func NewChatWidgetRepo(db *sqlx.DB) *ChatWidgetRepo {
	return &ChatWidgetRepo{db: db}
}

// CreateChatWidget creates a new chat widget
func (r *ChatWidgetRepo) CreateChatWidget(ctx context.Context, widget *models.ChatWidget) error {
	query := `
		INSERT INTO chat_widgets (
			id, tenant_id, project_id, name, is_active,
			primary_color, secondary_color, background_color, position, widget_shape, chat_bubble_style,
			widget_size, animation_style, custom_css,
			welcome_message, offline_message, custom_greeting, away_message,
			agent_name, agent_avatar_url,
			auto_open_delay, show_agent_avatars, allow_file_uploads, require_email, require_name,
			sound_enabled, show_powered_by, use_ai,
			business_hours, embed_code, created_at, updated_at
		) VALUES (
			:id, :tenant_id, :project_id, :name, :is_active,
			:primary_color, :secondary_color, :background_color, :position, :widget_shape, :chat_bubble_style,
			:widget_size, :animation_style, :custom_css,
			:welcome_message, :offline_message, :custom_greeting, :away_message,
			:agent_name, :agent_avatar_url,
			:auto_open_delay, :show_agent_avatars, :allow_file_uploads, :require_email, :require_name,
			:sound_enabled, :show_powered_by, :use_ai,
			:business_hours, :embed_code, :created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, widget)
	return err
}

// GetChatWidget gets a chat widget by ID
func (r *ChatWidgetRepo) GetChatWidget(ctx context.Context, tenantID, projectID, widgetID uuid.UUID) (*models.ChatWidget, error) {
	query := `
		SELECT cw.id, cw.tenant_id, cw.project_id, cw.name, cw.is_active,
			   cw.primary_color, cw.secondary_color, cw.background_color, cw.position, cw.widget_shape, cw.chat_bubble_style,
			   cw.widget_size, cw.animation_style, cw.custom_css,
			   cw.welcome_message, cw.offline_message, cw.custom_greeting, cw.away_message,
			   cw.agent_name, cw.agent_avatar_url,
			   cw.auto_open_delay, cw.show_agent_avatars, cw.allow_file_uploads, cw.require_email, cw.require_name,
			   cw.sound_enabled, cw.show_powered_by, cw.use_ai,
			   cw.business_hours, cw.embed_code, cw.created_at, cw.updated_at
		FROM chat_widgets cw
		WHERE cw.tenant_id = $1 AND cw.project_id = $2 AND cw.id = $3
	`

	var widget models.ChatWidget
	err := r.db.GetContext(ctx, &widget, query, tenantID, projectID, widgetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &widget, nil
}

func (r *ChatWidgetRepo) GetChatWidgetById(ctx context.Context, widgetID uuid.UUID) (*models.ChatWidget, error) {
	query := `
		SELECT cw.id, cw.tenant_id, cw.project_id, cw.name, cw.is_active,
			   cw.primary_color, cw.secondary_color, cw.background_color, cw.position, cw.widget_shape, cw.chat_bubble_style,
			   cw.widget_size, cw.animation_style, cw.custom_css,
			   cw.welcome_message, cw.offline_message, cw.custom_greeting, cw.away_message,
			   cw.agent_name, cw.agent_avatar_url,
			   cw.auto_open_delay, cw.show_agent_avatars, cw.allow_file_uploads, cw.require_email,
			   cw.require_name, cw.sound_enabled, cw.show_powered_by, cw.use_ai,
			   cw.business_hours, cw.embed_code, cw.created_at, cw.updated_at
		FROM chat_widgets cw
		WHERE cw.id = $1
	`

	var widget models.ChatWidget
	err := r.db.GetContext(ctx, &widget, query, widgetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &widget, nil
}

// GetChatWidgetByDomain gets a chat widget by domain for public access
func (r *ChatWidgetRepo) GetChatWidgetByDomain(ctx context.Context, domain string) (*models.ChatWidget, error) {
	query := `
		SELECT cw.id, cw.tenant_id, cw.project_id, cw.domain_id, cw.name, cw.is_active,
			   cw.primary_color, cw.secondary_color, cw.background_color, cw.position, cw.widget_shape, cw.chat_bubble_style,
			   cw.widget_size, cw.animation_style, cw.custom_css,
			   cw.welcome_message, cw.offline_message, cw.custom_greeting, cw.away_message,
			   cw.agent_name, cw.agent_avatar_url,
			   cw.auto_open_delay, cw.show_agent_avatars, cw.allow_file_uploads, cw.require_email,
			   cw.require_name, cw.sound_enabled, cw.show_powered_by, cw.use_ai,
			   cw.business_hours, cw.embed_code, cw.created_at, cw.updated_at
		FROM chat_widgets cw
		WHERE edv.domain = $1 AND cw.is_active = true AND edv.status = 'verified'
		LIMIT 1
	`

	var widget models.ChatWidget
	err := r.db.GetContext(ctx, &widget, query, domain)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &widget, nil
}

// ListChatWidgets lists all chat widgets for a project
func (r *ChatWidgetRepo) ListChatWidgets(ctx context.Context, tenantID, projectID uuid.UUID) ([]*models.ChatWidget, error) {
	query := `
		SELECT cw.id, cw.tenant_id, cw.project_id, cw.name, cw.is_active,
			   cw.primary_color, cw.secondary_color, cw.position, cw.widget_shape, cw.chat_bubble_style,
			   cw.widget_size, cw.animation_style, cw.custom_css,
			   cw.welcome_message, cw.offline_message, cw.custom_greeting, cw.away_message,
			   cw.agent_name, cw.agent_avatar_url,
			   cw.auto_open_delay, cw.show_agent_avatars, cw.allow_file_uploads, cw.require_email,
			   cw.sound_enabled, cw.show_powered_by, cw.use_ai,
			   cw.require_name, cw.business_hours, cw.embed_code, cw.created_at, cw.updated_at
		FROM chat_widgets cw
		WHERE cw.tenant_id = $1 AND cw.project_id = $2
		ORDER BY cw.created_at DESC
	`

	var widgets []*models.ChatWidget
	err := r.db.SelectContext(ctx, &widgets, query, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	return widgets, nil
}

// UpdateChatWidget updates a chat widget
func (r *ChatWidgetRepo) UpdateChatWidget(ctx context.Context, widget *models.ChatWidget) error {
	widget.UpdatedAt = time.Now()

	query := `
		UPDATE chat_widgets SET
			name = :name,
			is_active = :is_active,
			primary_color = :primary_color,
			secondary_color = :secondary_color,
			background_color = :background_color,
			position = :position,
			widget_shape = :widget_shape,
			chat_bubble_style = :chat_bubble_style,
			widget_size = :widget_size,
			animation_style = :animation_style,
			custom_css = :custom_css,
			welcome_message = :welcome_message,
			offline_message = :offline_message,
			custom_greeting = :custom_greeting,
			away_message = :away_message,
			agent_name = :agent_name,
			agent_avatar_url = :agent_avatar_url,
			auto_open_delay = :auto_open_delay,
			show_agent_avatars = :show_agent_avatars,
			allow_file_uploads = :allow_file_uploads,
			require_email = :require_email,
			require_name = :require_name,
			sound_enabled = :sound_enabled,
			show_powered_by = :show_powered_by,
			use_ai = :use_ai,
			business_hours = :business_hours,
			embed_code = :embed_code,
			updated_at = :updated_at
		WHERE tenant_id = :tenant_id AND project_id = :project_id AND id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, widget)
	return err
}

// DeleteChatWidget deletes a chat widget
func (r *ChatWidgetRepo) DeleteChatWidget(ctx context.Context, tenantID, projectID, widgetID uuid.UUID) error {
	query := `DELETE FROM chat_widgets WHERE tenant_id = $1 AND project_id = $2 AND id = $3`
	_, err := r.db.ExecContext(ctx, query, tenantID, projectID, widgetID)
	return err
}
