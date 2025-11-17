package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	MinIO         MinIOConfig         `mapstructure:"minio"`
	SMTP          SMTPConfig          `mapstructure:"smtp"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	CORS          CORSConfig          `mapstructure:"cors"`
	Features      FeatureFlags        `mapstructure:"features"`
	Agentic       AgenticConfig       `mapstructure:"agentic"`
	Email         EmailConfig         `mapstructure:"email"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	AI            AIConfig            `mapstructure:"ai"`
	Knowledge     KnowledgeConfig     `mapstructure:"knowledge"`
	Resend        ResendConfig        `mapstructure:"resend"`
	Maileroo      MailerooConfig      `mapstructure:"maileroo"`
	Payment       PaymentConfig       `mapstructure:"payment"`
	OAuth         OAuthConfig         `mapstructure:"oauth"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port                  string        `mapstructure:"port"`
	ReadTimeout           time.Duration `mapstructure:"read_timeout"`
	WriteTimeout          time.Duration `mapstructure:"write_timeout"`
	IdleTimeout           time.Duration `mapstructure:"idle_timeout"`
	Environment           string        `mapstructure:"environment"` // "development", "production", etc.
	AiAgentLoginAccessKey string        `mapstructure:"ai_agent_login_access_key"`
	PublicTicketUrl       string        `mapstructure:"public_ticket_url"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	// URL is an optional full database connection URL (e.g. postgres://user:pass@host:port/dbname?sslmode=disable)
	// If set, this will be used instead of composing host/port/user/password/dbname/sslmode.
	URL string `mapstructure:"url"`

	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Sentinels        []string `mapstructure:"sentinels"`         // Redis Sentinel URLs (comma-separated)
	URL              string   `mapstructure:"url"`               // Redis URL for local development
	Password         string   `mapstructure:"password"`          // Password for Redis master
	SentinelPassword string   `mapstructure:"sentinel_password"` // Password for Sentinel authentication
	MasterName       string   `mapstructure:"master_name"`       // Redis master name
}

// MinIOConfig represents MinIO configuration
type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
}

// SMTPConfig represents SMTP configuration
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret             string        `mapstructure:"secret"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
	MagicLinkExpiry    time.Duration `mapstructure:"magic_link_expiry"`
	UnauthTokenExpiry  time.Duration `mapstructure:"unauth_token_expiry"`
}

// FeatureFlags represents feature toggles
type FeatureFlags struct {
	EnableRegistration    bool `mapstructure:"enable_registration"`
	EnableEmailLogin      bool `mapstructure:"enable_email_login"`
	EnableMagicLinks      bool `mapstructure:"enable_magic_links"`
	RequireCorporateEmail bool `mapstructure:"require_corporate_email"`
}

// AgenticConfig represents agentic behavior configuration
type AgenticConfig struct {
	Enabled                   bool     `mapstructure:"enabled"`
	GreetingDetection         bool     `mapstructure:"greeting_detection"`
	KnowledgeResponses        bool     `mapstructure:"knowledge_responses"`
	AgentAssignment           bool     `mapstructure:"agent_assignment"`
	AgentRequestDetection     bool     `mapstructure:"agent_request_detection"`
	NotificationAlerts        bool     `mapstructure:"notification_alerts"`
	GreetingConfidence        float64  `mapstructure:"greeting_confidence"`
	KnowledgeConfidence       float64  `mapstructure:"knowledge_confidence"`
	DomainRelevanceConfidence float64  `mapstructure:"domain_relevance_confidence"`
	AgentRequestConfidence    float64  `mapstructure:"agent_request_confidence"`
	AgentRequestThreshold     float64  `mapstructure:"agent_request_threshold"`
	GreetingKeywords          []string `mapstructure:"greeting_keywords"`
	AgentRequestKeywords      []string `mapstructure:"agent_request_keywords"`
	NegativeKeywords          []string `mapstructure:"negative_keywords"`
	ResponseTimeoutMs         int      `mapstructure:"response_timeout_ms"`
	MaxConcurrentSessions     int      `mapstructure:"max_concurrent_sessions"`
}

// ResendConfig represents Resend email service configuration
type ResendConfig struct {
	APIKey    string `mapstructure:"api_key"`
	FromEmail string `mapstructure:"from_email"`
	FromName  string `mapstructure:"from_name"`
}

// MailerooConfig represents Maileroo email service configuration
type MailerooConfig struct {
	APIKey         string `mapstructure:"api_key"`
	FromEmail      string `mapstructure:"from_email"`
	FromName       string `mapstructure:"from_name"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
}

// EmailConfig represents email subsystem configuration
type EmailConfig struct {
	Provider                   string        `mapstructure:"provider"`
	DefaultIMAPPollingInterval time.Duration `mapstructure:"default_imap_polling_interval"`
	MaxAttachmentSize          int64         `mapstructure:"max_attachment_size"`
	EnableEmailToTicket        bool          `mapstructure:"enable_email_to_ticket"`
	DefaultReturnPathDomain    string        `mapstructure:"default_return_path_domain"`
}

// ObservabilityConfig represents observability configuration
type ObservabilityConfig struct {
	EnableTracing   bool   `mapstructure:"enable_tracing"`
	EnableMetrics   bool   `mapstructure:"enable_metrics"`
	TracingEndpoint string `mapstructure:"tracing_endpoint"`
	MetricsAddr     string `mapstructure:"metrics_addr"`
}

// AIConfig represents AI/LLM configuration
type AIConfig struct {
	Enabled              bool          `mapstructure:"enabled"`
	Provider             string        `mapstructure:"provider"` // "openai", "anthropic", "azure"
	APIKey               string        `mapstructure:"api_key"`
	ThemeExtractionModel string        `mapstructure:"theme_extraction_model"`
	UrlRankingModel      string        `mapstructure:"url_ranking_model"`
	Model                string        `mapstructure:"model"`
	BaseURL              string        `mapstructure:"base_url"`
	MaxTokens            int           `mapstructure:"max_tokens"`
	Temperature          float64       `mapstructure:"temperature"`
	SystemPrompt         string        `mapstructure:"system_prompt"`
	HandoffKeywords      []string      `mapstructure:"handoff_keywords"`
	AutoHandoffTime      time.Duration `mapstructure:"auto_handoff_time"`
}

// KnowledgeConfig represents knowledge management configuration
type KnowledgeConfig struct {
	Enabled                  bool          `mapstructure:"enabled"`
	AiAgentServiceUrl        string        `mapstructure:"ai_agent_service_url"`
	MaxFileSize              int64         `mapstructure:"max_file_size"`
	MaxFilesPerProject       int           `mapstructure:"max_files_per_project"`
	EmbeddingService         string        `mapstructure:"embedding_service"`
	OpenAIEmbeddingModel     string        `mapstructure:"openai_embedding_model"`
	OpenAIAPIKey             string        `mapstructure:"openai_api_key"`
	ChunkSize                int           `mapstructure:"chunk_size"`
	ChunkOverlap             int           `mapstructure:"chunk_overlap"`
	ScrapeMaxDepth           int           `mapstructure:"scrape_max_depth"`
	ScrapeRateLimit          time.Duration `mapstructure:"scrape_rate_limit"`
	ScrapeUserAgent          string        `mapstructure:"scrape_user_agent"`
	ScrapeTimeout            time.Duration `mapstructure:"scrape_timeout"`
	EmbeddingTimeout         time.Duration `mapstructure:"embedding_timeout"`
	PlaywrightWorkerCount    int           `mapstructure:"playwright_worker_count"`    // Workers for depth 0-1 (Playwright)
	CollyWorkerCount         int           `mapstructure:"colly_worker_count"`         // Workers for depth >= 2 (Colly)
	EnablePerformanceMetrics bool          `mapstructure:"enable_performance_metrics"` // Enable detailed performance tracking
	AllowedFileExtensions    []string      `mapstructure:"allowed_file_extensions"`    // File extensions to crawl (.html, .md, etc.)
}

// PaymentConfig represents payment gateway configuration
type PaymentConfig struct {
	Stripe   StripeConfig   `mapstructure:"stripe"`
	Cashfree CashfreeConfig `mapstructure:"cashfree"`
}

// StripeConfig represents Stripe configuration
type StripeConfig struct {
	WebhookSecret string `mapstructure:"webhook_secret"`
}

// CashfreeConfig represents Cashfree configuration
type CashfreeConfig struct {
	WebhookSecret string `mapstructure:"webhook_secret"`
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	Google GoogleOAuthConfig `mapstructure:"google"`
}

// GoogleOAuthConfig represents Google OAuth configuration
type GoogleOAuthConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/tms")

	// Set defaults
	setDefaults()

	// Enable environment variable binding with prefix
	viper.AutomaticEnv()
	viper.SetEnvPrefix("") // No prefix to allow direct env var names

	// Bind specific environment variables to config keys
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.environment", "APP_ENV")
	viper.BindEnv("server.ai_agent_login_access_key", "TMS_API_S2S_KEY")
	viper.BindEnv("server.public_ticket_url", "PUBLIC_TICKET_URL")
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	viper.BindEnv("database.url", "DATABASE_URL")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")

	// Connection pool settings
	viper.BindEnv("redis.sentinels", "REDIS_SENTINELS")
	viper.BindEnv("redis.url", "REDIS_URL")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.sentinel_password", "REDIS_SENTINEL_PASSWORD")
	viper.BindEnv("redis.master_name", "REDIS_MASTER_NAME")
	viper.BindEnv("redis.environment", "APP_ENV")

	// JWT
	viper.BindEnv("jwt.secret", "JWT_SECRET_KEY")
	viper.BindEnv("jwt.access_token_expiry", "JWT_TOKEN_EXPIRY")
	viper.BindEnv("jwt.refresh_token_expiry", "JWT_REFRESH_TOKEN_EXPIRY")

	// AI configuration bindings
	viper.BindEnv("ai.enabled", "AI_ENABLED")
	viper.BindEnv("ai.provider", "AI_PROVIDER")
	viper.BindEnv("ai.theme_extraction_model", "AI_THEME_EXTRACTION_MODEL")
	viper.BindEnv("ai.url_ranking_model", "AI_URL_RANKING_MODEL")
	viper.BindEnv("ai.api_key", "AI_API_KEY")
	viper.BindEnv("ai.model", "AI_MODEL")
	viper.BindEnv("ai.base_url", "AI_BASE_URL")
	viper.BindEnv("ai.max_tokens", "AI_MAX_TOKENS")
	viper.BindEnv("ai.temperature", "AI_TEMPERATURE")
	viper.BindEnv("ai.system_prompt", "AI_SYSTEM_PROMPT")
	viper.BindEnv("ai.auto_handoff_time", "AI_AUTO_HANDOFF_TIME")

	// Knowledge management configuration bindings
	viper.BindEnv("knowledge.enabled", "KNOWLEDGE_ENABLED")
	viper.BindEnv("knowledge.ai_agent_service_url", "AI_AGENT_SERVICE_URL")
	viper.BindEnv("knowledge.max_file_size", "KNOWLEDGE_MAX_FILE_SIZE")
	viper.BindEnv("knowledge.max_files_per_project", "KNOWLEDGE_MAX_FILES_PER_PROJECT")
	viper.BindEnv("knowledge.embedding_service", "KNOWLEDGE_EMBEDDING_SERVICE")
	viper.BindEnv("knowledge.openai_embedding_model", "KNOWLEDGE_OPENAI_EMBEDDING_MODEL")
	viper.BindEnv("knowledge.openai_api_key", "OPENAI_API_KEY")
	viper.BindEnv("knowledge.chunk_size", "KNOWLEDGE_CHUNK_SIZE")
	viper.BindEnv("knowledge.chunk_overlap", "KNOWLEDGE_CHUNK_OVERLAP")
	viper.BindEnv("knowledge.scrape_max_depth", "KNOWLEDGE_SCRAPE_MAX_DEPTH")
	viper.BindEnv("knowledge.scrape_rate_limit", "KNOWLEDGE_SCRAPE_RATE_LIMIT")
	viper.BindEnv("knowledge.scrape_user_agent", "KNOWLEDGE_SCRAPE_USER_AGENT")
	viper.BindEnv("knowledge.scrape_timeout", "KNOWLEDGE_SCRAPE_TIMEOUT")
	viper.BindEnv("knowledge.embedding_timeout", "KNOWLEDGE_EMBEDDING_TIMEOUT")
	viper.BindEnv("knowledge.playwright_worker_count", "KNOWLEDGE_PLAYWRIGHT_WORKER_COUNT")
	viper.BindEnv("knowledge.colly_worker_count", "KNOWLEDGE_COLLY_WORKER_COUNT")
	viper.BindEnv("knowledge.enable_performance_metrics", "KNOWLEDGE_ENABLE_PERFORMANCE_METRICS")
	viper.BindEnv("knowledge.allowed_file_extensions", "KNOWLEDGE_ALLOWED_FILE_EXTENSIONS")

	// Email subsystem bindings
	viper.BindEnv("email.provider", "EMAIL_PROVIDER")

	// Resend configuration bindings
	viper.BindEnv("resend.api_key", "RESEND_API_KEY")
	viper.BindEnv("resend.from_email", "EMAIL_FROM_ADDRESS")
	viper.BindEnv("resend.from_name", "EMAIL_FROM_NAME")

	// Maileroo configuration bindings
	viper.BindEnv("maileroo.api_key", "MAILEROO_API_KEY")
	viper.BindEnv("maileroo.from_email", "EMAIL_FROM_ADDRESS")
	viper.BindEnv("maileroo.from_name", "EMAIL_FROM_NAME")
	viper.BindEnv("maileroo.timeout_seconds", "MAILEROO_TIMEOUT_SECONDS")

	// CORS configuration bindings
	viper.BindEnv("cors.allowed_origins", "CORS_ORIGINS")
	viper.BindEnv("cors.allow_credentials", "CORS_ALLOW_CREDENTIALS")

	// OAuth configuration bindings
	viper.BindEnv("oauth.google.client_id", "GOOGLE_OAUTH_CLIENT_ID")
	viper.BindEnv("oauth.google.client_secret", "GOOGLE_OAUTH_CLIENT_SECRET")
	viper.BindEnv("oauth.google.redirect_url", "GOOGLE_OAUTH_REDIRECT_URL")

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Handle comma-separated REDIS_SENTINELS environment variable
	if sentinelsStr := viper.GetString("redis.sentinels"); sentinelsStr != "" {
		sentinels := strings.Split(sentinelsStr, ",")
		for i, sentinel := range sentinels {
			sentinels[i] = strings.TrimSpace(sentinel)
		}
		config.Redis.Sentinels = sentinels
	}

	// Handle comma-separated CORS_ORIGINS environment variable
	if corsOriginsStr := viper.GetString("cors.allowed_origins"); corsOriginsStr != "" {
		origins := strings.Split(corsOriginsStr, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
			if origins[i] == `""` { // literal two quotes
				origins[i] = ""
			}
		}
		config.CORS.AllowedOrigins = origins
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.public_ticket_url", "http://localhost:3001")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "tms")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.url", "")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 25)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	viper.SetDefault("redis.sentinels", []string{})
	viper.SetDefault("redis.url", "")

	// MinIO defaults
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.access_key_id", "minioadmin")
	viper.SetDefault("minio.secret_access_key", "minioadmin")
	viper.SetDefault("minio.use_ssl", false)
	viper.SetDefault("minio.bucket_name", "tms-attachments")

	// SMTP defaults
	viper.SetDefault("smtp.host", "localhost")
	viper.SetDefault("smtp.port", 1025)
	viper.SetDefault("smtp.username", "")
	viper.SetDefault("smtp.password", "")
	viper.SetDefault("smtp.from", "noreply@example.com")

	// Email provider defaults
	viper.SetDefault("email.provider", "resend")
	viper.SetDefault("maileroo.timeout_seconds", 30)

	// JWT defaults
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.access_token_expiry", "24h")
	viper.SetDefault("jwt.refresh_token_expiry", "168h")

	// Feature flags defaults
	viper.SetDefault("features.enable_registration", true)
	viper.SetDefault("features.enable_email_login", true)
	viper.SetDefault("features.enable_magic_links", true)

	// AI defaults
	viper.SetDefault("ai.enabled", true)
	viper.SetDefault("ai.provider", "bb")
	viper.BindEnv("ai.theme_extraction_model", "gpt-4o-mini")
	viper.SetDefault("ai.model", "gpt-4o-mini")
	viper.SetDefault("ai.max_tokens", 1000)
	viper.SetDefault("ai.temperature", 0.7)
	viper.SetDefault("ai.system_prompt", "You are a helpful customer support assistant. Be concise, professional, and friendly. If you cannot help with a request, suggest that a human agent will take over.")
	viper.SetDefault("ai.auto_handoff_time", "10m")

	// Knowledge management defaults
	viper.SetDefault("knowledge.enabled", true)
	viper.SetDefault("knowledge.ai_agent_service_url", "http://localhost:8090")
	viper.SetDefault("knowledge.max_file_size", 10485760) // 10MB
	viper.SetDefault("knowledge.max_files_per_project", 100)
	viper.SetDefault("knowledge.embedding_service", "openai")
	viper.SetDefault("knowledge.openai_embedding_model", "text-embedding-ada-002")
	viper.SetDefault("knowledge.chunk_size", 1000)
	viper.SetDefault("knowledge.chunk_overlap", 200)
	viper.SetDefault("knowledge.scrape_max_depth", 5)
	viper.SetDefault("knowledge.scrape_rate_limit", "1s")
	viper.SetDefault("knowledge.scrape_user_agent", "Hith Knowledge Bot 1.0")
	viper.SetDefault("knowledge.scrape_timeout", "30s")
	viper.SetDefault("knowledge.embedding_timeout", "120s")
	viper.SetDefault("knowledge.playwright_worker_count", 3) // Conservative for browser overhead
	viper.SetDefault("knowledge.colly_worker_count", 15)     // Higher for lightweight HTTP
	viper.SetDefault("knowledge.enable_performance_metrics", true)
	viper.SetDefault("knowledge.allowed_file_extensions", []string{".html", ".htm", ".md", ".markdown", ".txt", ".pdf"})

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"*"})
	viper.SetDefault("cors.allow_credentials", false)

	// OAuth defaults
	viper.SetDefault("oauth.google.client_id", "")
	viper.SetDefault("oauth.google.client_secret", "")
	viper.SetDefault("oauth.google.redirect_url", "http://localhost:3000/auth/google/callback")
}
