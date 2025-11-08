package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/models"
)

type AIBuilderService struct {
	chatWidgetService  *ChatWidgetService
	webScrapingService *WebScrapingService
	knowledgeService   *KnowledgeService
	aiService          *AIService
}

type AIBuilderEvent struct {
	Type      string         `json:"type"`
	Stage     string         `json:"stage,omitempty"`
	Message   string         `json:"message,omitempty"`
	Detail    string         `json:"detail,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

func NewAIBuilderService(chatWidget *ChatWidgetService, webScraper *WebScrapingService, knowledge *KnowledgeService, ai *AIService) *AIBuilderService {
	return &AIBuilderService{
		chatWidgetService:  chatWidget,
		webScrapingService: webScraper,
		knowledgeService:   knowledge,
		aiService:          ai,
	}
}

func (s *AIBuilderService) Run(ctx context.Context, tenantID, projectID uuid.UUID, rootURL string, depth int, generateFaq bool, events chan<- AIBuilderEvent) error {
	if depth <= 0 {
		depth = 3
	}

	parsedRoot, err := url.Parse(rootURL)
	if err != nil || parsedRoot.Scheme == "" || parsedRoot.Host == "" {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "initialization",
			Message: "Invalid URL provided",
			Detail:  errMessage(err),
		})
		return fmt.Errorf("invalid root url: %w", err)
	}

	baseDomain := parsedRoot.Host
	rootURL = parsedRoot.String()

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "builder_started",
		Stage:   "initialization",
		Message: fmt.Sprintf("Starting AI workspace setup for %s", rootURL),
	})

	widget, err := s.buildWidget(ctx, tenantID, projectID, rootURL, events)
	if err != nil {
		return err
	}

	jobID, selections, err := s.buildKnowledge(ctx, tenantID, projectID, rootURL, baseDomain, depth, events)
	if err != nil {
		return err
	}

	faqItems := []*models.KnowledgeFAQItem{}
	if generateFaq {

		faqItems, err = s.generateFAQ(ctx, tenantID, projectID, rootURL, jobID, selections, events)
		if err != nil {
			return err
		}
	}

	embedCode := ""
	if widget != nil && widget.EmbedCode != nil {
		embedCode = *widget.EmbedCode
	}

	data := map[string]any{
		"embed_code":     embedCode,
		"widget_id":      widget.ID.String(),
		"project_id":     projectID.String(),
		"selected_links": selections,
		"faq_count":      len(faqItems),
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "completed",
		Stage:   "summary",
		Message: "AI builder completed successfully",
		Data:    data,
	})

	return nil
}

func (s *AIBuilderService) buildWidget(ctx context.Context, tenantID, projectID uuid.UUID, rootURL string, events chan<- AIBuilderEvent) (*models.ChatWidget, error) {
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "widget_stage_started",
		Stage:   "widget",
		Message: "Analyzing website brand and building chat widget",
	})

	themeData, err := s.webScrapingService.ScrapeWebsiteTheme(ctx, rootURL)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "widget",
			Message: "Failed to analyze website theme",
			Detail:  err.Error(),
		})
		return nil, fmt.Errorf("scrape theme: %w", err)
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "widget_theme_ready",
		Stage:   "widget",
		Message: "Brand palette extracted",
		Data: map[string]any{
			"brand_name":  themeData.BrandName,
			"page_title":  themeData.PageTitle,
			"color_count": len(themeData.Colors),
		},
	})

	cfg, err := s.aiService.GenerateWidgetTheme(ctx, themeData)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "widget",
			Message: "Failed to generate widget theme",
			Detail:  err.Error(),
		})
		return nil, fmt.Errorf("generate widget theme: %w", err)
	}

	s.applyWidgetDefaults(cfg)

	widget, err := s.createOrUpdateWidget(ctx, tenantID, projectID, cfg)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "widget",
			Message: "Failed to persist chat widget",
			Detail:  err.Error(),
		})
		return nil, fmt.Errorf("create/update widget: %w", err)
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "widget_ready",
		Stage:   "widget",
		Message: "Chat widget is ready",
		Data: map[string]any{
			"widget_id": widget.ID.String(),
		},
	})

	return widget, nil
}

func (s *AIBuilderService) buildKnowledge(ctx context.Context, tenantID, projectID uuid.UUID, rootURL, baseDomain string, depth int, events chan<- AIBuilderEvent) (uuid.UUID, []KnowledgeLinkSelection, error) {
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "knowledge_stage_started",
		Stage:   "scraping",
		Message: "Discovering high-signal knowledge pages",
	})

	scrapingEvents := make(chan ScrapingEvent)
	req := &models.CreateScrapingJobRequest{URL: rootURL, MaxDepth: depth}
	job, err := s.webScrapingService.CreateScrapingJobWithStream(ctx, tenantID, projectID, req, scrapingEvents)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "scraping",
			Message: "Failed to start website scraping",
			Detail:  err.Error(),
		})
		return uuid.Nil, nil, fmt.Errorf("create scraping job: %w", err)
	}

	if err := s.consumeScrapingEvents(ctx, scrapingEvents, events); err != nil {
		return uuid.Nil, nil, err
	}

	previews, err := s.loadDiscoveredLinks(ctx, job.ID, tenantID, projectID, events)
	if err != nil {
		return uuid.Nil, nil, err
	}

	candidates := make([]KnowledgeLinkCandidate, 0, len(previews))
	candidateSet := make(map[string]struct{}, len(previews))
	for _, preview := range previews {
		if preview == nil {
			continue
		}
		parsed, parseErr := url.Parse(preview.URL)
		if parseErr != nil || parsed.Host != baseDomain {
			continue
		}
		key := normalizeURL(preview.URL)
		if _, exists := candidateSet[key]; exists {
			continue
		}
		candidateSet[key] = struct{}{}
		candidates = append(candidates, KnowledgeLinkCandidate{
			URL:        preview.URL,
			Title:      preview.Title,
			Depth:      preview.Depth,
			TokenCount: preview.TokenCount,
		})
	}

	if len(candidates) == 0 {
		err := fmt.Errorf("no suitable pages discovered for knowledge ingestion")
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "scraping",
			Message: err.Error(),
		})
		return uuid.Nil, nil, err
	}

	selections, err := s.aiService.SelectTopKnowledgeLinks(ctx, rootURL, candidates, 5)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "scraping",
			Message: "Failed to prioritise knowledge pages",
			Detail:  err.Error(),
		})
		return uuid.Nil, nil, fmt.Errorf("select knowledge links: %w", err)
	}

	ordered := make([]string, 0, len(selections))
	filtered := make([]KnowledgeLinkSelection, 0, len(selections))
	unique := make(map[string]struct{}, len(selections))
	for _, sel := range selections {
		key := normalizeURL(sel.URL)
		if _, exists := unique[key]; exists {
			continue
		}
		if _, exists := candidateSet[key]; !exists {
			continue
		}
		unique[key] = struct{}{}
		ordered = append(ordered, sel.URL)
		filtered = append(filtered, sel)
	}

	if len(ordered) == 0 {
		err := fmt.Errorf("AI selection did not return valid links")
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "scraping",
			Message: err.Error(),
		})
		return uuid.Nil, nil, err
	}

	if err := s.webScrapingService.StoreLinkSelection(ctx, job.ID, tenantID, projectID, ordered); err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "scraping",
			Message: "Failed to store selected links",
			Detail:  err.Error(),
		})
		return uuid.Nil, nil, fmt.Errorf("store link selection: %w", err)
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "knowledge_links_chosen",
		Stage:   "scraping",
		Message: "Key pages selected for indexing",
		Data: map[string]any{
			"count": len(ordered),
		},
	})

	if err := s.streamIndexing(ctx, job.ID, tenantID, projectID, events); err != nil {
		return uuid.Nil, nil, err
	}

	return job.ID, filtered, nil
}

func (s *AIBuilderService) generateFAQ(ctx context.Context, tenantID, projectID uuid.UUID, rootURL string, jobID uuid.UUID, selections []KnowledgeLinkSelection, events chan<- AIBuilderEvent) ([]*models.KnowledgeFAQItem, error) {
	pages, err := s.webScrapingService.GetJobPages(ctx, jobID, tenantID, projectID)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "faq",
			Message: "Failed to load scraped pages",
			Detail:  err.Error(),
		})
		return nil, fmt.Errorf("get job pages: %w", err)
	}

	pageMap := make(map[string]*models.KnowledgeScrapedPage, len(pages))
	for _, page := range pages {
		if page == nil {
			continue
		}
		pageMap[normalizeURL(page.URL)] = page
	}

	orderedSections := make([]KnowledgeSectionSummary, 0, len(selections))
	for _, sel := range selections {
		page := pageMap[normalizeURL(sel.URL)]
		if page == nil {
			continue
		}
		title := ""
		if page.Title != nil {
			title = *page.Title
		}
		orderedSections = append(orderedSections, KnowledgeSectionSummary{
			URL:     page.URL,
			Title:   title,
			Content: page.Content,
		})
	}

	// Fallback: if no selections matched, use all available scraped pages
	if len(orderedSections) == 0 {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "faq_fallback",
			Stage:   "faq",
			Message: "Using all scraped pages for FAQ generation",
			Data: map[string]any{
				"selections_count": len(selections),
				"pages_count":      len(pages),
			},
		})

		for _, page := range pages {
			if page == nil || page.Content == "" {
				continue
			}
			title := ""
			if page.Title != nil {
				title = *page.Title
			}
			orderedSections = append(orderedSections, KnowledgeSectionSummary{
				URL:     page.URL,
				Title:   title,
				Content: page.Content,
			})
		}
	}

	if len(orderedSections) == 0 {
		err := fmt.Errorf("no content available for FAQ generation")
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "faq",
			Message: err.Error(),
		})
		return nil, err
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "faq_generation_started",
		Stage:   "faq",
		Message: "Generating knowledge Q&A",
	})

	generated, err := s.aiService.GenerateKnowledgeFAQ(ctx, rootURL, orderedSections, 10)
	if err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "faq",
			Message: "Failed to generate FAQs",
			Detail:  err.Error(),
		})
		return nil, fmt.Errorf("generate faq: %w", err)
	}

	selectionLookup := make(map[string]KnowledgeLinkSelection)
	for _, sel := range selections {
		selectionLookup[normalizeURL(sel.URL)] = sel
	}

	faqItems := make([]*models.KnowledgeFAQItem, 0, len(generated))
	now := time.Now()
	for _, item := range generated {
		question := strings.TrimSpace(item.Question)
		answer := strings.TrimSpace(item.Answer)
		if question == "" || answer == "" {
			continue
		}

		faq := &models.KnowledgeFAQItem{
			ID:        uuid.New(),
			Question:  question,
			Answer:    answer,
			Metadata:  models.JSONMap{"source": "auto_ai_builder"},
			CreatedAt: now,
			UpdatedAt: now,
		}

		if trimmed := strings.TrimSpace(item.SourceURL); trimmed != "" {
			sel := selectionLookup[normalizeURL(trimmed)]
			if sel.Category != "" {
				faq.Metadata["category"] = sel.Category
			}
			faq.Metadata["rationale"] = sel.Rationale
			urlCopy := trimmed
			faq.SourceURL = &urlCopy
		}

		if section := strings.TrimSpace(item.SectionRef); section != "" {
			sectionCopy := section
			faq.SourceSection = &sectionCopy
		}

		faqItems = append(faqItems, faq)
	}

	if len(faqItems) == 0 {
		err := fmt.Errorf("FAQ generation returned empty results")
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "faq",
			Message: err.Error(),
		})
		return nil, err
	}

	if err := s.knowledgeService.ReplaceFAQItems(ctx, tenantID, projectID, faqItems); err != nil {
		s.emit(ctx, events, AIBuilderEvent{
			Type:    "error",
			Stage:   "faq",
			Message: "Failed to persist FAQs",
			Detail:  err.Error(),
		})
		return nil, fmt.Errorf("persist faq items: %w", err)
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "faq_ready",
		Stage:   "faq",
		Message: "Knowledge Q&A compiled",
		Data: map[string]any{
			"count": len(faqItems),
		},
	})

	return faqItems, nil
}

func (s *AIBuilderService) consumeScrapingEvents(ctx context.Context, scraping <-chan ScrapingEvent, events chan<- AIBuilderEvent) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-scraping:
			if !ok {
				return nil
			}
			payload := map[string]any{
				"job_id":        event.JobID,
				"message":       event.Message,
				"url":           event.URL,
				"current_depth": event.CurrentDepth,
				"max_depth":     event.MaxDepth,
				"links_found":   event.LinksFound,
			}
			s.emit(ctx, events, AIBuilderEvent{
				Type:    "scraping_" + event.Type,
				Stage:   "scraping",
				Message: event.Message,
				Data:    payload,
			})

			if event.Type == "error" {
				return fmt.Errorf("scraping error: %s", event.Message)
			}
		}
	}
}

func (s *AIBuilderService) streamIndexing(ctx context.Context, jobID, tenantID, projectID uuid.UUID, events chan<- AIBuilderEvent) error {
	s.emit(ctx, events, AIBuilderEvent{
		Type:    "indexing_started",
		Stage:   "indexing",
		Message: "Indexing selected pages",
	})

	idxEvents := make(chan IndexingEvent)
	go func() {
		defer close(idxEvents)
		if err := s.webScrapingService.StreamIndexing(ctx, jobID, tenantID, projectID, idxEvents); err != nil {
			s.emit(ctx, events, AIBuilderEvent{
				Type:    "error",
				Stage:   "indexing",
				Message: "Indexing failed",
				Detail:  err.Error(),
			})
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-idxEvents:
			if !ok {
				s.emit(ctx, events, AIBuilderEvent{
					Type:    "indexing_completed",
					Stage:   "indexing",
					Message: "Indexing finished",
				})
				return nil
			}
			payload := map[string]any{
				"message":      event.Message,
				"url":          event.URL,
				"completed":    event.Completed,
				"pending":      event.Pending,
				"token_count":  event.TokenCount,
				"total_tokens": event.TotalTokens,
			}
			s.emit(ctx, events, AIBuilderEvent{
				Type:    "indexing_" + event.Type,
				Stage:   "indexing",
				Message: event.Message,
				Data:    payload,
			})

			if event.Type == "error" {
				return fmt.Errorf("indexing error: %s", event.Message)
			}
		}
	}
}

func (s *AIBuilderService) loadDiscoveredLinks(ctx context.Context, jobID, tenantID, projectID uuid.UUID, events chan<- AIBuilderEvent) ([]*models.ScrapedLinkPreview, error) {
	var previews []*models.ScrapedLinkPreview
	var err error

	for i := 0; i < 5; i++ {
		previews, err = s.webScrapingService.GetStagedLinks(ctx, jobID, tenantID, projectID)
		if err == nil {
			return previews, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(400 * time.Millisecond):
		}
	}

	s.emit(ctx, events, AIBuilderEvent{
		Type:    "error",
		Stage:   "scraping",
		Message: "Unable to retrieve discovered links",
		Detail:  err.Error(),
	})
	return nil, fmt.Errorf("load discovered links: %w", err)
}

func (s *AIBuilderService) createOrUpdateWidget(ctx context.Context, tenantID, projectID uuid.UUID, cfg *models.CreateChatWidgetRequest) (*models.ChatWidget, error) {
	widgets, err := s.chatWidgetService.ListChatWidgets(ctx, tenantID, projectID)
	if err != nil {
		return nil, err
	}

	if len(widgets) == 0 {
		return s.chatWidgetService.CreateChatWidget(ctx, tenantID, projectID, cfg)
	}

	widget := widgets[0]

	update := &models.UpdateChatWidgetRequest{}
	update.PrimaryColor = stringPtr(cfg.PrimaryColor)
	update.SecondaryColor = stringPtr(cfg.SecondaryColor)
	update.BackgroundColor = stringPtr(cfg.BackgroundColor)
	update.Position = stringPtr(cfg.Position)
	update.WelcomeMessage = stringPtr(cfg.WelcomeMessage)
	update.OfflineMessage = stringPtr(cfg.OfflineMessage)
	update.ChatBubbleStyle = stringPtr(cfg.ChatBubbleStyle)
	update.WidgetShape = stringPtr(cfg.WidgetShape)
	update.AgentName = stringPtr(cfg.AgentName)
	update.RequireEmail = boolPtr(cfg.RequireEmail)
	update.RequireName = boolPtr(cfg.RequireName)
	update.ShowAgentAvatars = boolPtr(cfg.ShowAgentAvatars)
	update.AllowFileUploads = boolPtr(cfg.AllowFileUploads)
	update.UseAI = boolPtr(true)
	if cfg.CustomGreeting != nil {
		update.CustomGreeting = cfg.CustomGreeting
	}
	if cfg.AgentAvatarURL != nil {
		update.AgentAvatarURL = cfg.AgentAvatarURL
	}

	return s.chatWidgetService.UpdateChatWidget(ctx, tenantID, projectID, widget.ID, update)
}

func (s *AIBuilderService) applyWidgetDefaults(cfg *models.CreateChatWidgetRequest) {
	if cfg.PrimaryColor == "" {
		cfg.PrimaryColor = "#2563eb"
	}
	if cfg.SecondaryColor == "" {
		cfg.SecondaryColor = "#f3f4f6"
	}
	if cfg.BackgroundColor == "" {
		cfg.BackgroundColor = "#ffffff"
	}
	if cfg.Position == "" {
		cfg.Position = "bottom-right"
	}
	if cfg.WelcomeMessage == "" {
		cfg.WelcomeMessage = "Hello! How can we help you today?"
	}
	if cfg.OfflineMessage == "" {
		cfg.OfflineMessage = "We are currently offline. Leave us a message and we will respond soon."
	}
	if cfg.AgentName == "" {
		cfg.AgentName = "Support Agent"
	}
	cfg.UseAI = true
}

func (s *AIBuilderService) emit(ctx context.Context, events chan<- AIBuilderEvent, event AIBuilderEvent) {
	event.Timestamp = time.Now()
	select {
	case <-ctx.Done():
	case events <- event:
	}
}

func normalizeURL(u string) string {
	return strings.TrimRight(strings.TrimSpace(u), "/")
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}

func boolPtr(value bool) *bool {
	v := value
	return &v
}

func errMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
