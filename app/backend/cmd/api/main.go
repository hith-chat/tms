package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bareuptime/tms/internal/auth"
	"github.com/bareuptime/tms/internal/config" // Global middleware
	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/handlers"
	"github.com/bareuptime/tms/internal/mail"
	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/rbac"
	"github.com/bareuptime/tms/internal/redis"
	"github.com/bareuptime/tms/internal/repo"
	"github.com/bareuptime/tms/internal/service"
	"github.com/bareuptime/tms/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize JWT auth
	jwtAuth := auth.NewService(
		cfg.JWT.Secret,
		int(cfg.JWT.AccessTokenExpiry.Seconds()),
		int(cfg.JWT.RefreshTokenExpiry.Seconds()),
		int(cfg.JWT.MagicLinkExpiry.Seconds()),
		int(cfg.JWT.UnauthTokenExpiry.Seconds()),
	)

	// Initialize RBAC service
	rbacService := rbac.NewService(database.DB.DB)

	// Initialize repositories
	ticketRepo := repo.NewTicketRepository(database.DB.DB)
	agentRepo := repo.NewAgentRepository(database.DB.DB)
	customerRepo := repo.NewCustomerRepository(database.DB.DB)
	messageRepo := repo.NewTicketMessageRepository(database.DB.DB)
	projectRepo := repo.NewProjectRepository(database.DB)
	integrationRepo := repo.NewIntegrationRepository(database.DB)
	emailRepo := repo.NewEmailRepo(database.DB)
	apiKeyRepo := repo.NewApiKeyRepository(database.DB)
	settingsRepo := repo.NewSettingsRepository(database.DB.DB)
	tenantRepo := repo.NewTenantRepository(database.DB.DB)
	emailInboxRepo := repo.NewEmailInboxRepository(database.DB.DB)
	domainValidationRepo := repo.NewDomainValidationRepo(database.DB)
	notificationRepo := repo.NewNotificationRepo(database.DB)

	// Chat repositories
	chatWidgetRepo := repo.NewChatWidgetRepo(database.DB)
	chatSessionRepo := repo.NewChatSessionRepo(database.DB)
	chatMessageRepo := repo.NewChatMessageRepo(database.DB)

	// Knowledge management repositories
	knowledgeRepo := repo.NewKnowledgeRepository(database.DB)

	// Alarm repository
	alarmRepo := repo.NewAlarmRepository(database.DB)

	// Initialize mail service
	mailLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	mailService := mail.NewService(mailLogger)

	// Initialize Redis service
	redisService := redis.NewService(redis.RedisConfig{
		Sentinels:        cfg.Redis.Sentinels,
		URL:              cfg.Redis.URL,
		Password:         cfg.Redis.Password,
		SentinelPassword: cfg.Redis.SentinelPassword,
		MasterName:       cfg.Redis.MasterName,
		Environment:      cfg.Server.Environment,
	})

	// Initialize services
	resendService := service.NewResendService(&cfg.Resend, cfg.Server.Environment)

	// Create feature flags for auth service
	authFeatureFlags := &service.FeatureFlags{
		RequireCorporateEmail: cfg.Features.RequireCorporateEmail,
	}

	authService := service.NewAuthService(agentRepo, rbacService, jwtAuth, redisService, resendService, authFeatureFlags, tenantRepo, domainValidationRepo, projectRepo)
	projectService := service.NewProjectService(projectRepo)
	agentService := service.NewAgentService(agentRepo, projectRepo, rbacService)
	tenantService := service.NewTenantService(tenantRepo, agentRepo, rbacService)
	customerService := service.NewCustomerService(customerRepo, rbacService)
	messageService := service.NewMessageService(messageRepo, ticketRepo, customerRepo, agentRepo, rbacService)
	publicService := service.NewPublicService(ticketRepo, messageRepo, jwtAuth, messageService)

	ticketService := service.NewTicketService(ticketRepo, customerRepo, agentRepo, messageRepo, rbacService, mailService, publicService, resendService, cfg.Server.PublicTicketUrl)
	emailInboxService := service.NewEmailInboxService(emailInboxRepo, ticketRepo, messageRepo, customerRepo, emailRepo, mailService, mailLogger)
	domainValidationService := service.NewDomainValidationService(domainValidationRepo, mailService)

	// Chat services
	chatWidgetService := service.NewChatWidgetService(chatWidgetRepo, domainValidationRepo)

	// Initialize enterprise connection manager (needed for chat session service)
	connectionManager := websocket.NewConnectionManager(redisService.GetClient())
	defer connectionManager.Shutdown()

	// Alarm services (Phase 4 implementation) - needed by chat session service
	howlingAlarmService := service.NewHowlingAlarmService(cfg, connectionManager, alarmRepo)

	chatSessionService := service.NewChatSessionService(chatSessionRepo, chatMessageRepo, chatWidgetRepo, customerRepo, ticketService, agentService, connectionManager, redisService, howlingAlarmService)

	// Knowledge management services
	embeddingService := service.NewEmbeddingService(&cfg.Knowledge)
	documentProcessorService := service.NewDocumentProcessorService(knowledgeRepo, embeddingService, "./uploads", cfg.Knowledge.MaxFileSize)
	webScrapingService := service.NewWebScrapingService(knowledgeRepo, embeddingService, &cfg.Knowledge)
	knowledgeService := service.NewKnowledgeService(knowledgeRepo, embeddingService)

	// Greeting services for agentic behavior
	greetingDetectionService := service.NewGreetingDetectionService(&cfg.Agentic)
	brandGreetingService := service.NewBrandGreetingService(settingsRepo)

	// Integration services
	integrationService := service.NewIntegrationService(integrationRepo)

	// Notification service (needs connection manager for WebSocket delivery)
	notificationService := service.NewNotificationService(notificationRepo, connectionManager)
	// Enhanced notification service for agentic behavior
	// enhancedNotificationService := service.NewEnhancedNotificationService(notificationRepo, connectionManager, howlingAlarmService, cfg)

	// AI service (needs knowledge service for RAG, greeting services for agentic behavior, connection manager for handoff notifications, and auto assignment service)
	aiService := service.NewAIService(&cfg.AI, &cfg.Agentic, chatSessionService, knowledgeService, greetingDetectionService, brandGreetingService, connectionManager, howlingAlarmService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, publicService, cfg.Server.AiAgentLoginAccessKey)
	projectHandler := handlers.NewProjectHandler(projectService)
	ticketHandler := handlers.NewTicketHandler(ticketService, messageService)
	publicHandler := handlers.NewPublicHandler(publicService)
	integrationHandler := handlers.NewIntegrationHandler(integrationService)
	emailHandler := handlers.NewEmailHandler(emailRepo, redisService, mailService)
	emailInboxHandler := handlers.NewEmailInboxHandler(emailInboxService)
	agentHandler := handlers.NewAgentHandler(agentService)
	customerHandler := handlers.NewCustomerHandler(customerService)
	apiKeyHandler := handlers.NewApiKeyHandler(apiKeyRepo)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	domainValidationHandler := handlers.NewDomainValidationHandler(domainValidationService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Chat handlers
	chatWidgetHandler := handlers.NewChatWidgetHandler(chatWidgetService)
	chatSessionHandler := handlers.NewChatSessionHandler(chatSessionService, chatWidgetService, redisService)

	// Knowledge management handlers
	knowledgeHandler := handlers.NewKnowledgeHandler(documentProcessorService, webScrapingService, knowledgeService)

	alarmHandler := handlers.NewAlarmHandler(howlingAlarmService)

	// Initialize agent client for Python agent service communication
	agentClient := service.NewAgentClient(cfg.Knowledge.AiAgentServiceUrl)

	chatWebSocketHandler := handlers.NewChatWebSocketHandler(chatSessionService, connectionManager, notificationService, aiService, agentClient, jwtAuth)
	agentWebSocketHandler := handlers.NewAgentWebSocketHandler(chatSessionService, connectionManager, agentService)

	// Set up combined message handling - ChatWebSocketHandler handles all Redis pub/sub messages
	// since it manages both visitor and agent connections
	agentWebSocketHandler.SetChatWSHandler(chatWebSocketHandler)

	// Setup router
	router := setupRouter(database.DB.DB, jwtAuth, apiKeyRepo, &cfg.CORS, authHandler, projectHandler, ticketHandler, publicHandler, integrationHandler, emailHandler, emailInboxHandler, agentHandler, customerHandler, apiKeyHandler, settingsHandler, tenantHandler, domainValidationHandler, notificationHandler, chatWidgetHandler, chatSessionHandler, chatWebSocketHandler, agentWebSocketHandler, knowledgeHandler, alarmHandler)

	// Create HTTP server
	serverAddr := cfg.Server.Port
	// Ensure address has proper format (add colon if just port number)
	if serverAddr[0] != ':' && !strings.Contains(serverAddr, ":") {
		serverAddr = ":" + serverAddr
	}
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on the port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter(database *sql.DB, jwtAuth *auth.Service, apiKeyRepo repo.ApiKeyRepository, corsConfig *config.CORSConfig, authHandler *handlers.AuthHandler, projectHandler *handlers.ProjectHandler, ticketHandler *handlers.TicketHandler, publicHandler *handlers.PublicHandler, integrationHandler *handlers.IntegrationHandler, emailHandler *handlers.EmailHandler, emailInboxHandler *handlers.EmailInboxHandler, agentHandler *handlers.AgentHandler, customerHandler *handlers.CustomerHandler, apiKeyHandler *handlers.ApiKeyHandler, settingsHandler *handlers.SettingsHandler, tenantHandler *handlers.TenantHandler, domainNameHandler *handlers.DomainNameHandler, notificationHandler *handlers.NotificationHandler, chatWidgetHandler *handlers.ChatWidgetHandler, chatSessionHandler *handlers.ChatSessionHandler, chatWebSocketHandler *handlers.ChatWebSocketHandler, agentWebSocketHandler *handlers.AgentWebSocketHandler, knowledgeHandler *handlers.KnowledgeHandler, alarmHandler *handlers.AlarmHandler) *gin.Engine {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.CORSMiddleware(corsConfig))
	router.Use(middleware.TenantMiddleware(database))

	// Health check endpoints - support both GET and HEAD for load balancers
	router.GET("/health", publicHandler.Health)
	router.HEAD("/health", publicHandler.Health)

	// Public routes
	publicRoutes := router.Group("/api/public")
	{
		publicRoutes.GET("/health", publicHandler.Health)
		publicRoutes.GET("/tickets/tokens/:token", publicHandler.GetTicketByMagicLink)
		publicRoutes.GET("/tickets/tokens/:token/messages", publicHandler.GetTicketMessagesByMagicLink)
		publicRoutes.POST("/tickets/tokens/:token/messages", publicHandler.AddMessageByMagicLink)
		// Testing endpoint - remove in production
		publicRoutes.POST("/generate-magic-link", publicHandler.GenerateMagicLink)

		// Ticket ID based routes (without magic link)
		publicRoutes.GET("/tickets/:ticketId", publicHandler.GetTicketByID)
		publicRoutes.GET("/tickets/:ticketId/messages", publicHandler.GetTicketMessagesByID)
		publicRoutes.POST("/tickets/:ticketId/messages", publicHandler.AddMessageByID)
	}

	// Auth routes (not protected by auth middleware)
	authRoutes := router.Group("/v1/auth")
	{

		authRoutes.POST("/refresh", authHandler.Refresh)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/ai-agent/tenant/:tenant_id/project/:project_id/login", authHandler.AiAgentLogin)
		authRoutes.POST("/signup", authHandler.SignUp)
		authRoutes.POST("/verify-signup-otp", authHandler.VerifySignupOTP)
		authRoutes.POST("/resend-signup-otp", authHandler.ResendSignupOTP)
	}

	// Enterprise admin routes (protected by auth middleware but cross-tenant)
	enterprise := router.Group("/v1/enterprise")
	enterprise.Use(middleware.AuthMiddleware(jwtAuth))
	{
		enterprise.GET("/tenants", tenantHandler.ListTenants)
	}

	// API routes (protected by auth middleware)
	api := router.Group("/v1/tenants/:tenant_id")
	api.Use(middleware.AuthMiddleware(jwtAuth))
	{
		// Global agent WebSocket endpoint (not session-specific)
		api.GET("/chat/agent/ws", agentWebSocketHandler.HandleAgentWebSocket)

		// Authentication endpoints that require auth
		auth := api.Group("/auth")
		{
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", authHandler.Me)
		}

		// Project management endpoints
		{
			api.GET("/projects", projectHandler.ListProjects)
			api.POST("/projects", projectHandler.CreateProject)
			api.GET("/projects/:project_id", projectHandler.GetProject)
			api.PUT("/projects/:project_id", projectHandler.UpdateProject)
			api.DELETE("/projects/:project_id", projectHandler.DeleteProject)
		}

		// Global integration endpoints (not project-scoped)
		// {
		// 	// Integration categories and templates
		// 	api.GET("/integrations/categories", integrationHandler.ListIntegrationCategories)
		// 	api.GET("/integrations/templates", integrationHandler.ListIntegrationTemplates)
		// 	api.GET("/integrations/templates/:type", integrationHandler.GetIntegrationTemplate)

		// 	// OAuth endpoints
		// 	api.POST("/integrations/oauth/start", integrationHandler.StartOAuth)
		// 	api.POST("/integrations/:type/oauth/callback", integrationHandler.HandleOAuthCallback)
		// }

		// Agent management endpoints
		{
			api.GET("/agents", agentHandler.ListAgents)
			api.POST("/agents", agentHandler.CreateAgent)
			api.GET("/agents/:agent_id", agentHandler.GetAgent)
			api.PATCH("/agents/:agent_id", agentHandler.UpdateAgent)
			api.DELETE("/agents/:agent_id", agentHandler.DeleteAgent)
			api.POST("/agents/:agent_id/roles", agentHandler.AssignRole)
			api.DELETE("/agents/:agent_id/roles", agentHandler.RemoveRole)
			api.GET("/agents/:agent_id/roles", agentHandler.GetAgentRoles)
			// Project assignment endpoints - restricted to tenant admins only
			api.POST("/agents/:agent_id/projects/:project_id", middleware.TenantAdminMiddleware(), agentHandler.AssignToProject)
			api.DELETE("/agents/:agent_id/projects/:project_id", middleware.TenantAdminMiddleware(), agentHandler.RemoveFromProject)
			api.GET("/agents/:agent_id/projects", agentHandler.GetAgentProjects)
			// Agent notification preferences (Phase 4)
			api.GET("/agents/:agent_id/notification-preferences", alarmHandler.GetNotificationPreferences)
			api.PUT("/agents/:agent_id/notification-preferences", alarmHandler.UpdateNotificationPreferences)
		}

		// Customer management (tenant-level)
		{
			api.POST("/customers", middleware.TenantAdminMiddleware(), customerHandler.CreateCustomer)
			api.PUT("/customers/:customer_id", middleware.TenantAdminMiddleware(), customerHandler.UpdateCustomer)
			// Deletion should be protected by appropriate middleware/permissions; service will also enforce RBAC
			api.DELETE("/customers/:customer_id", middleware.TenantAdminMiddleware(), customerHandler.DeleteCustomer)
		}

		// API Key management endpoints

		// Project-scoped endpoints
		projects := api.Group("/projects/:project_id")
		{
			apiKeys := projects.Group("/api-keys")
			{
				apiKeys.GET("", apiKeyHandler.ListApiKeys)
				apiKeys.POST("", apiKeyHandler.CreateApiKey)
				apiKeys.GET("/:key_id", apiKeyHandler.GetApiKey)
				apiKeys.PATCH("/:key_id", apiKeyHandler.UpdateApiKey)
				apiKeys.DELETE("/:key_id", apiKeyHandler.DeleteApiKey)
			}

			// Settings endpoints
			settings := projects.Group("/settings")
			{
				settings.GET("/branding", middleware.ProjectAdminMiddleware(), settingsHandler.GetBrandingSettings)
				settings.PUT("/branding", middleware.ProjectAdminMiddleware(), settingsHandler.UpdateBrandingSettings)
				settings.GET("/automation", middleware.ProjectAdminMiddleware(), settingsHandler.GetAutomationSettings)
				settings.PUT("/automation", middleware.ProjectAdminMiddleware(), settingsHandler.UpdateAutomationSettings)
				settings.GET("/about-me", middleware.ProjectAdminMiddleware(), settingsHandler.GetAboutMeSettings)
				settings.PUT("/about-me", middleware.ProjectAdminMiddleware(), settingsHandler.UpdateAboutMeSettings)
			}

			// Notifications endpoints
			notifications := projects.Group("/notifications")
			{
				notifications.GET("", notificationHandler.GetNotifications)
				notifications.GET("/count", notificationHandler.GetNotificationCount)
				notifications.PUT("/:notification_id/read", notificationHandler.MarkNotificationAsRead)
				notifications.PUT("/mark-all-read", notificationHandler.MarkAllNotificationsAsRead)
			}

			// Alarms endpoints (Phase 4 implementation)
			alarms := projects.Group("/alarms")
			{
				alarms.GET("/active", alarmHandler.GetActiveAlarms)
				alarms.GET("/stats", alarmHandler.GetAlarmStats)
				alarms.POST("/:alarmId/acknowledge", alarmHandler.AcknowledgeAlarm)
			}

			// Integrations - using the available methods
			integrations := projects.Group("/integrations")
			{
				// Integration categories and templates
				integrations.GET("/categories", integrationHandler.ListIntegrationCategories)
				integrations.GET("/templates", integrationHandler.ListIntegrationTemplates)
				integrations.GET("/templates/:type", integrationHandler.GetIntegrationTemplate)

				// OAuth endpoints
				integrations.POST("/oauth/start", integrationHandler.StartOAuth)
				integrations.POST("/:type/oauth/callback", integrationHandler.HandleOAuthCallback)
				integrations.GET("", integrationHandler.ListIntegrations)
				integrations.POST("", integrationHandler.CreateIntegration)
				integrations.GET("/with-templates", integrationHandler.ListIntegrationsWithTemplates)
				integrations.GET("/:integration_id", integrationHandler.GetIntegration)
				integrations.PATCH("/:integration_id", integrationHandler.UpdateIntegration)
				integrations.DELETE("/:integration_id", integrationHandler.DeleteIntegration)
				// integrations.POST("/:integration_id/test", integrationHandler.TestIntegrationConnection)
				// integrations.GET("/:integration_id/metrics", integrationHandler.GetIntegrationMetrics)

				// Integration configurations
				// integrations.POST("/:integration_id/slack", integrationHandler.CreateSlackConfiguration)
				// integrations.POST("/:integration_id/jira", integrationHandler.CreateJiraConfiguration)
				// integrations.POST("/:integration_id/calendly", integrationHandler.CreateCalendlyConfiguration)
				// integrations.POST("/:integration_id/zapier", integrationHandler.CreateZapierConfiguration)

				// Webhook subscriptions
				// webhooks := integrations.Group("/:integration_id/webhooks")
				// {
				// 	webhooks.GET("", integrationHandler.ListWebhookSubscriptions)
				// 	webhooks.POST("", integrationHandler.CreateWebhookSubscription)
				// }
			}

			// Email connectors and mailboxes
			email := projects.Group("/email")
			{
				// Email connectors
				email.GET("/connectors", emailHandler.ListConnectors)
				email.POST("/connectors", emailHandler.CreateConnector)
				email.GET("/connectors/:connector_id", emailHandler.GetConnector)
				email.PATCH("/connectors/:connector_id", emailHandler.UpdateConnector)
				email.DELETE("/connectors/:connector_id", emailHandler.DeleteConnector)
				email.POST("/connectors/:connector_id/test", emailHandler.TestConnector)
				email.POST("/connectors/:connector_id/validate", emailHandler.ValidateConnector)
				email.POST("/connectors/:connector_id/verify-otp", emailHandler.VerifyConnectorOTP)

				// Email mailboxes
				email.GET("/mailboxes", emailHandler.ListMailboxes)
				email.POST("/mailboxes", emailHandler.CreateMailbox)
				email.GET("/mailboxes/:mailbox_id", emailHandler.GetMailbox)
				email.PUT("/mailboxes/:mailbox_id", emailHandler.UpdateMailbox)
				email.DELETE("/mailboxes/:mailbox_id", emailHandler.DeleteMailbox)

				// Email inbox
				inbox := email.Group("/inbox")
				{
					inbox.GET("", emailInboxHandler.ListEmails)
					inbox.GET("/:id", emailInboxHandler.GetEmail)
					inbox.POST("/:id/convert-to-ticket", emailInboxHandler.ConvertToTicket)
					inbox.POST("/:id/reply", emailInboxHandler.ReplyToEmail)
					inbox.POST("/:id/mark-read", emailInboxHandler.MarkAsRead)
					inbox.POST("/sync", emailInboxHandler.SyncEmails)
					inbox.GET("/sync-status", emailInboxHandler.GetSyncStatus)
				}

				// Domain validation
				domains := email.Group("/domains")
				{
					domains.GET("", domainNameHandler.ListDomainNames)
					domains.POST("", domainNameHandler.CreateDomainName)
					domains.POST("/:domain_id/verify", domainNameHandler.VerifyDomain)
					domains.DELETE("/:domain_id", domainNameHandler.DeleteDomainName)
				}
			}

			// Chat system endpoints
			chat := projects.Group("/chat")
			{
				// Chat widgets
				chat.GET("/widgets", chatWidgetHandler.ListChatWidgets)
				chat.POST("/widgets", chatWidgetHandler.CreateChatWidget)
				chat.GET("/widgets/:widget_id", chatWidgetHandler.GetChatWidget)
				chat.PATCH("/widgets/:widget_id", chatWidgetHandler.UpdateChatWidget)
				chat.DELETE("/widgets/:widget_id", chatWidgetHandler.DeleteChatWidget)

				// Chat sessions (agent endpoints)
				chat.GET("/sessions", chatSessionHandler.ListChatSessions)
				chat.GET("/sessions/:session_id", chatSessionHandler.GetChatSession)
				chat.POST("/sessions/:session_id/assign", chatSessionHandler.AssignAgent)
				chat.POST("/sessions/:session_id/escalate", middleware.TenantAdminMiddleware(), chatSessionHandler.EscalateSession)
				chat.GET("/sessions/:session_id/messages", chatSessionHandler.GetChatMessages)
				chat.POST("/sessions/:session_id/messages/:message_id/read", chatSessionHandler.MarkAgentMessagesAsRead)
				chat.GET("/sessions/:session_id/client/status", chatSessionHandler.IsCustomerOnline)

			}

			// Knowledge management endpoints
			knowledge := projects.Group("/knowledge")
			{
				// Document management
				knowledge.POST("/documents", knowledgeHandler.UploadDocument)
				knowledge.GET("/documents", knowledgeHandler.ListDocuments)
				knowledge.GET("/documents/:document_id", knowledgeHandler.GetDocument)
				knowledge.DELETE("/documents/:document_id", knowledgeHandler.DeleteDocument)

				// Web scraping
				knowledge.POST("/scrape", knowledgeHandler.CreateScrapingJob)
				knowledge.GET("/scraping-jobs", knowledgeHandler.ListScrapingJobs)
				knowledge.GET("/scraping-jobs/:job_id", knowledgeHandler.GetScrapingJob)
				knowledge.GET("/scraping-jobs/:job_id/pages", knowledgeHandler.GetJobPages)

				// Knowledge search
				knowledge.POST("/search", knowledgeHandler.SearchKnowledgeBase)
				knowledge.GET("/search", knowledgeHandler.SearchKnowledgeBaseGET)

				// Settings
				knowledge.GET("/settings", knowledgeHandler.GetKnowledgeSettings)
				knowledge.PUT("/settings", knowledgeHandler.UpdateKnowledgeSettings)

				// Statistics
				knowledge.GET("/stats", knowledgeHandler.GetKnowledgeStats)
			}
		}

		// Public chat endpoints (no authentication required)
		publicChat := router.Group("/api/public/chat")
		{
			// Widget endpoints
			publicChat.GET("/widgets/domain/:domain", chatWidgetHandler.GetChatWidgetByDomain)
			publicChat.GET("/widgets/:widget_id", chatWidgetHandler.GetChatWidgetByPublicId)

			// Public chat session endpoints (token-based auth)
			publicChat.POST("/sessions/:session_id/messages/:message_id/read", chatSessionHandler.MarkVisitorMessagesAsRead)

			// WebSocket endpoint for visitors
			publicChat.GET("/ws/widgets/:widget_id/chat/:session_token", chatWebSocketHandler.HandleWebSocketPublic)
		}
	}

	// Tickets with flexible authentication (JWT or API key) - separate from api group to avoid inheriting AuthMiddleware
	flexibleTickets := router.Group("/v1/tenants/:tenant_id/projects/:project_id/tickets")
	flexibleTickets.Use(middleware.ApiKeyOrJWTAuthMiddleware(apiKeyRepo, jwtAuth))
	flexibleTickets.Use(middleware.TicketAccessMiddleware())
	{
		flexibleTickets.GET("", ticketHandler.ListTickets)
		flexibleTickets.POST("", ticketHandler.CreateTicket)
		flexibleTickets.GET("/:ticket_id", ticketHandler.GetTicket)

		// Apply reassignment middleware for update operations
		flexibleTickets.PATCH("/:ticket_id", middleware.TicketReassignmentMiddleware(), ticketHandler.UpdateTicket)

		// Dedicated reassignment endpoint (requires admin permissions)
		flexibleTickets.POST("/:ticket_id/reassign", middleware.ProjectAdminMiddleware(), ticketHandler.ReassignTicket)

		// Delete ticket (requires admin permissions)
		flexibleTickets.DELETE("/:ticket_id", middleware.ProjectAdminMiddleware(), ticketHandler.DeleteTicket)

		// Customer validation and magic links
		flexibleTickets.POST("/:ticket_id/validate-customer", ticketHandler.ValidateCustomer)
		flexibleTickets.POST("/:ticket_id/send-magic-link", ticketHandler.SendMagicLink)

		// Ticket messages
		flexibleTickets.GET("/:ticket_id/messages", ticketHandler.GetTicketMessages)
		flexibleTickets.POST("/:ticket_id/messages", ticketHandler.AddMessage)
		flexibleTickets.PATCH("/:ticket_id/messages/:message_id", ticketHandler.UpdateMessage)
		flexibleTickets.DELETE("/:ticket_id/messages/:message_id", ticketHandler.DeleteMessage)
	}

	simpleTicketUrls := router.Group("/v1/tickets")
	simpleTicketUrls.Use(middleware.ApiKeyOrJWTAuthMiddleware(apiKeyRepo, jwtAuth))
	{
		simpleTicketUrls.GET("", ticketHandler.ListTickets)
		simpleTicketUrls.POST("", ticketHandler.CreateTicket)
		simpleTicketUrls.GET("/:ticket_id", ticketHandler.GetTicket)

		// Apply reassignment middleware for update operations
		simpleTicketUrls.PATCH("/:ticket_id", middleware.TicketReassignmentMiddleware(), ticketHandler.UpdateTicket)

		// Delete ticket (requires admin permissions)
		simpleTicketUrls.DELETE("/:ticket_id", middleware.ProjectAdminMiddleware(), ticketHandler.DeleteTicket)
	}

	return router
}
