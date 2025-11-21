package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/bareuptime/tms/internal/logger"
	"github.com/bareuptime/tms/internal/models"
	redisService "github.com/bareuptime/tms/internal/redis"
	"github.com/bareuptime/tms/internal/repo"
)

// SlackService handles Slack API interactions
type SlackService struct {
	integrationRepo *repo.ProjectIntegrationRepository
	sessionRepo     *repo.ChatSessionRepo
	redisService    *redisService.Service
	httpClient      *http.Client
}

// NewSlackService creates a new Slack service
func NewSlackService(
	integrationRepo *repo.ProjectIntegrationRepository,
	sessionRepo *repo.ChatSessionRepo,
	redisService *redisService.Service,
) *SlackService {
	return &SlackService{
		integrationRepo: integrationRepo,
		sessionRepo:     sessionRepo,
		redisService:    redisService,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SlackMessageRequest represents a Slack API message request
type SlackMessageRequest struct {
	Channel   string `json:"channel"`
	Text      string `json:"text"`
	ThreadTS  string `json:"thread_ts,omitempty"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

// SlackMessageResponse represents a Slack API message response
type SlackMessageResponse struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	TS      string `json:"ts"` // Message timestamp
	Channel string `json:"channel"`
	Message struct {
		ThreadTS string `json:"thread_ts,omitempty"`
	} `json:"message"`
}

// SlackUserInfo represents Slack user information
type SlackUserInfo struct {
	OK   bool `json:"ok"`
	User struct {
		ID       string `json:"id"`
		RealName string `json:"real_name"`
		Profile  struct {
			DisplayName string `json:"display_name"`
			RealName    string `json:"real_name"`
		} `json:"profile"`
	} `json:"user"`
}

// GetSlackIntegration retrieves the Slack integration for a widget's project
func (s *SlackService) GetSlackIntegration(ctx context.Context, tenantID, projectID uuid.UUID) (*models.SlackIntegrationMeta, error) {
	integration, err := s.integrationRepo.GetByProjectAndType(ctx, tenantID, projectID, models.ProjectIntegrationTypeSlack)
	if err != nil {
		return nil, fmt.Errorf("failed to get Slack integration: %w", err)
	}
	if integration == nil {
		return nil, nil // No Slack integration configured
	}

	if integration.Status != models.ProjectIntegrationStatusActive {
		return nil, fmt.Errorf("Slack integration is not active")
	}

	slackMeta, err := models.SlackMetaFromIntegration(integration.Meta)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Slack metadata: %w", err)
	}

	return slackMeta, nil
}

// PostMessageToSlack posts a message to Slack using chat.postMessage API
// Note: We use the API instead of incoming webhooks because we need threading support
func (s *SlackService) PostMessageToSlack(
	ctx context.Context,
	tenantID, projectID uuid.UUID,
	session *models.ChatSession,
	messageContent string,
	senderName string,
) error {
	// Get Slack integration for this widget's project
	fmt.Println("fetch slack")
	slackMeta, err := s.GetSlackIntegration(ctx, tenantID, projectID)
	fmt.Println(slackMeta)
	if err != nil {
		logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to get Slack integration")
		return err
	}
	if slackMeta == nil {
		fmt.Println("slack meta nil")
		// No Slack integration configured, skip
		return nil
	}

	// Get the channel from webhooks (use first webhook's channel)
	if len(slackMeta.Webhooks) == 0 {
		return fmt.Errorf("no Slack webhook configured")
	}
	channelID := slackMeta.Webhooks[0].ChannelID

	// Format message with sender name
	text := messageContent
	if senderName != "" {
		text = fmt.Sprintf("*%s:* %s", senderName, messageContent)
	}

	// Check if thread already exists
	fmt.Println("session --> ", session)
	slackThreadMeta := session.GetSlackMeta()
	fmt.Println("slack thread meta", slackThreadMeta)
	var threadTS string
	if slackThreadMeta != nil {
		threadTS = slackThreadMeta.ThreadTS
	}

	// Prepare Slack API request
	reqBody := SlackMessageRequest{
		Channel:  channelID,
		Text:     text,
		ThreadTS: threadTS,
	}

	// Call Slack API using access token
	resp, err := s.postMessageToSlackAPI(ctx, slackMeta.AccessToken, &reqBody)
	if err != nil {
		return err
	}

	fmt.Println("slack post response", resp)

	// If this is the first message (no thread yet), store the thread_ts
	if slackThreadMeta == nil {
		// Use thread_ts if available, otherwise use the message ts
		newThreadTS := resp.Message.ThreadTS
		if newThreadTS == "" {
			newThreadTS = resp.TS
		}

		fmt.Println("slack post response bikanuuc --> ", resp.TS, resp.Message.ThreadTS)

		// Update session with Slack thread info (both in-memory and database)
		session.SetSlackMeta(newThreadTS, channelID)

		// Update session in database using dedicated columns method
		if err := s.sessionRepo.UpdateSlackThreadInfo(ctx, session.ID, newThreadTS, channelID); err != nil {
			logger.GetTxLogger(ctx).Error().Err(err).Msg("Failed to update session Slack thread info")
			return err
		}

		logger.GetTxLogger(ctx).Info().
			Str("session_id", session.ID.String()).
			Str("thread_ts", newThreadTS).
			Str("channel_id", channelID).
			Msg("Created Slack thread for chat session")
	}

	return nil
}

// postMessageToSlackAPI calls the Slack chat.postMessage API
func (s *SlackService) postMessageToSlackAPI(ctx context.Context, token string, req *SlackMessageRequest) (*SlackMessageResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	fmt.Print("slakckck", string(jsonData))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Slack API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var slackResp SlackMessageResponse
	if err := json.Unmarshal(body, &slackResp); err != nil {
		return nil, err
	}

	if !slackResp.OK {
		return nil, fmt.Errorf("Slack API error: %s", slackResp.Error)
	}

	return &slackResp, nil
}

// GetUserDisplayName fetches a Slack user's display name with Redis caching
func (s *SlackService) GetUserDisplayName(ctx context.Context, token, userID string) (string, error) {
	// Try to get from Redis cache first
	cacheKey := slackUserCacheKey(userID)
	cachedName, err := s.redisService.GetClient().Get(ctx, cacheKey).Result()
	if err == nil && cachedName != "" {
		logger.GetTxLogger(ctx).Debug().
			Str("user_id", userID).
			Str("display_name", cachedName).
			Msg("Slack user display name retrieved from cache")
		return cachedName, nil
	}

	// If not in cache or error, fetch from Slack API
	if err != nil && err != redis.Nil {
		logger.GetTxLogger(ctx).Warn().Err(err).Msg("Failed to get Slack user from cache, fetching from API")
	}

	url := fmt.Sprintf("https://slack.com/api/users.info?user=%s", userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var userInfo SlackUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return "", err
	}

	if !userInfo.OK {
		return "", fmt.Errorf("failed to get Slack user info")
	}

	// Prefer display name, fall back to real name
	displayName := userInfo.User.Profile.DisplayName
	if displayName == "" {
		displayName = userInfo.User.Profile.RealName
	}
	if displayName == "" {
		displayName = userInfo.User.RealName
	}
	if displayName == "" {
		displayName = "Slack User"
	}

	// Cache the display name for 24 hours
	if err := s.redisService.GetClient().Set(ctx, cacheKey, displayName, 24*time.Hour).Err(); err != nil {
		logger.GetTxLogger(ctx).Warn().Err(err).Msg("Failed to cache Slack user display name")
		// Don't fail the request if caching fails
	} else {
		logger.GetTxLogger(ctx).Debug().
			Str("user_id", userID).
			Str("display_name", displayName).
			Msg("Slack user display name cached")
	}

	return displayName, nil
}

// slackUserCacheKey generates a Redis cache key for Slack user display names
func slackUserCacheKey(userID string) string {
	return fmt.Sprintf("slack:user:%s:display_name", userID)
}
