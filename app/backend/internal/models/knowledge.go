package models

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// KnowledgeDocument represents a document uploaded to the knowledge base
type KnowledgeDocument struct {
	ID               uuid.UUID `db:"id" json:"id"`
	TenantID         uuid.UUID `db:"tenant_id" json:"tenant_id"`
	ProjectID        uuid.UUID `db:"project_id" json:"project_id"`
	Filename         string    `db:"filename" json:"filename"`
	ContentType      string    `db:"content_type" json:"content_type"`
	FileSize         int64     `db:"file_size" json:"file_size"`
	FilePath         string    `db:"file_path" json:"file_path"`
	OriginalContent  *string   `db:"original_content" json:"original_content,omitempty"`
	ProcessedContent *string   `db:"processed_content" json:"processed_content,omitempty"`
	Status           string    `db:"status" json:"status"`
	ErrorMessage     *string   `db:"error_message" json:"error_message,omitempty"`
	Metadata         JSONMap   `db:"metadata" json:"metadata"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// KnowledgeChunk represents a chunk of a document with its embedding
type KnowledgeChunk struct {
	ID         uuid.UUID        `db:"id" json:"id"`
	DocumentID uuid.UUID        `db:"document_id" json:"document_id"`
	ChunkIndex int              `db:"chunk_index" json:"chunk_index"`
	Content    string           `db:"content" json:"content"`
	TokenCount int              `db:"token_count" json:"token_count"`
	Embedding  *pgvector.Vector `db:"embedding" json:"embedding,omitempty"`
	Metadata   JSONMap          `db:"metadata" json:"metadata"`
	CreatedAt  time.Time        `db:"created_at" json:"created_at"`
}

// KnowledgeScrapingJob represents a web scraping job
type KnowledgeScrapingJob struct {
	ID                  uuid.UUID      `db:"id" json:"id"`
	TenantID            uuid.UUID      `db:"tenant_id" json:"tenant_id"`
	ProjectID           uuid.UUID      `db:"project_id" json:"project_id"`
	URL                 string         `db:"url" json:"url"`
	MaxDepth            int            `db:"max_depth" json:"max_depth"`
	Status              string         `db:"status" json:"status"`
	PagesScraped        int            `db:"pages_scraped" json:"pages_scraped"`
	TotalPages          int            `db:"total_pages" json:"total_pages"`
	StagingFilePath     *string        `db:"staging_file_path" json:"staging_file_path,omitempty"`
	SelectedLinks       pq.StringArray `db:"selected_links" json:"selected_links"`
	IndexingStartedAt   *time.Time     `db:"indexing_started_at" json:"indexing_started_at,omitempty"`
	IndexingCompletedAt *time.Time     `db:"indexing_completed_at" json:"indexing_completed_at,omitempty"`
	ErrorMessage        *string        `db:"error_message" json:"error_message,omitempty"`
	StartedAt           *time.Time     `db:"started_at" json:"started_at,omitempty"`
	CompletedAt         *time.Time     `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt           time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at" json:"updated_at"`
}

// ScrapedLinkPreview represents a staged scraped page before indexing
type ScrapedLinkPreview struct {
	URL            string `json:"url"`
	Title          string `json:"title,omitempty"`
	Depth          int    `json:"depth"`
	TokenCount     int    `json:"token_count"`
	ContentPreview string `json:"content_preview,omitempty"`
	Selected       bool   `json:"selected"`
}

// KnowledgeScrapedPage represents a scraped web page
type KnowledgeScrapedPage struct {
	ID          uuid.UUID        `db:"id" json:"id"`
	JobID       uuid.NullUUID    `db:"job_id" json:"job_id,omitempty"` // Nullable - tracks FIRST job that discovered this page
	PageID      *uuid.UUID       `db:"page_id" json:"page_id,omitempty"`
	URL         string           `db:"url" json:"url"`
	Title       *string          `db:"title" json:"title,omitempty"`
	Content     string           `db:"content" json:"content"`
	ContentHash *string          `db:"content_hash" json:"content_hash,omitempty"`
	TokenCount  int              `db:"token_count" json:"token_count"`
	ScrapedAt   time.Time        `db:"scraped_at" json:"scraped_at"`
	Embedding   *pgvector.Vector `db:"embedding" json:"embedding,omitempty"`
	Metadata    JSONMap          `db:"metadata" json:"metadata"`
}

// WidgetKnowledgePage represents the association between a widget and a knowledge page
type WidgetKnowledgePage struct {
	ID        uuid.UUID `db:"id" json:"id"`
	WidgetID  uuid.UUID `db:"widget_id" json:"widget_id"`
	PageID    uuid.UUID `db:"page_id" json:"page_id"`
	TenantID  uuid.UUID `db:"tenant_id" json:"tenant_id"`
	ProjectID uuid.UUID `db:"project_id" json:"project_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// WidgetKnowledgePageWithDetails represents a widget knowledge page with page details
type WidgetKnowledgePageWithDetails struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	WidgetID  uuid.UUID  `db:"widget_id" json:"widget_id"`
	PageID    uuid.UUID  `db:"page_id" json:"page_id"`
	TenantID  uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	ProjectID uuid.UUID  `db:"project_id" json:"project_id"`
	URL       string     `db:"url" json:"url"`
	Title     *string    `db:"title" json:"title,omitempty"`
	TokenCount int       `db:"token_count" json:"token_count"`
	ScrapedAt time.Time  `db:"scraped_at" json:"scraped_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	JobID     *uuid.UUID `db:"job_id" json:"job_id,omitempty"`
}

// KnowledgeSettings represents knowledge base settings for a project
type KnowledgeSettings struct {
	ID                  uuid.UUID `db:"id" json:"id"`
	TenantID            uuid.UUID `db:"tenant_id" json:"tenant_id"`
	ProjectID           uuid.UUID `db:"project_id" json:"project_id"`
	Enabled             bool      `db:"enabled" json:"enabled"`
	EmbeddingModel      string    `db:"embedding_model" json:"embedding_model"`
	ChunkSize           int       `db:"chunk_size" json:"chunk_size"`
	ChunkOverlap        int       `db:"chunk_overlap" json:"chunk_overlap"`
	MaxContextChunks    int       `db:"max_context_chunks" json:"max_context_chunks"`
	SimilarityThreshold float64   `db:"similarity_threshold" json:"similarity_threshold"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

// Request/Response models

// UploadDocumentRequest represents a document upload request
type UploadDocumentRequest struct {
	Filename    string `form:"filename" binding:"required"`
	ContentType string `form:"content_type"`
}

// CreateScrapingJobRequest represents a web scraping job creation request
type CreateScrapingJobRequest struct {
	URL      string `json:"url" binding:"required,url"`
	MaxDepth int    `json:"max_depth" binding:"min=1,max=5"`
}

// SelectScrapingLinksRequest captures user-selected URLs for indexing
type SelectScrapingLinksRequest struct {
	URLs []string `json:"urls" binding:"required,min=1"`
}

// KnowledgeSearchRequest represents a knowledge search request
type KnowledgeSearchRequest struct {
	Query            string  `json:"query" binding:"required"`
	MaxResults       int     `json:"max_results" binding:"min=1,max=20"`
	SimilarityScore  float64 `json:"similarity_score" binding:"min=0,max=1"`
	IncludeDocuments bool    `json:"include_documents"`
	IncludePages     bool    `json:"include_pages"`
}

// KnowledgeSearchResult represents a single search result
type KnowledgeSearchResult struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	Type       string     `json:"type" db:"type"` // "document" or "webpage"
	Content    string     `json:"content" db:"content"`
	Score      float64    `json:"score" db:"score"`
	Source     string     `json:"source" db:"source"` // filename or URL
	Title      *string    `json:"title,omitempty" db:"title"`
	DocumentID *uuid.UUID `json:"document_id,omitempty" db:"document_id"`
	JobID      *uuid.UUID `json:"job_id,omitempty" db:"job_id"`
	ChunkIndex *int       `json:"chunk_index,omitempty" db:"chunk_index"`
	Metadata   JSONMap    `json:"metadata" db:"metadata"`
}

// KnowledgeSearchResponse represents a search response
type KnowledgeSearchResponse struct {
	Results     []KnowledgeSearchResult `json:"results"`
	TotalCount  int                     `json:"total_count"`
	Query       string                  `json:"query"`
	ProcessedIn string                  `json:"processed_in"`
}

// UpdateKnowledgeSettingsRequest represents a settings update request
type UpdateKnowledgeSettingsRequest struct {
	Enabled             *bool    `json:"enabled,omitempty"`
	EmbeddingModel      *string  `json:"embedding_model,omitempty"`
	ChunkSize           *int     `json:"chunk_size,omitempty" binding:"omitempty,min=100,max=2000"`
	ChunkOverlap        *int     `json:"chunk_overlap,omitempty" binding:"omitempty,min=0,max=500"`
	MaxContextChunks    *int     `json:"max_context_chunks,omitempty" binding:"omitempty,min=1,max=10"`
	SimilarityThreshold *float64 `json:"similarity_threshold,omitempty" binding:"omitempty,min=0,max=1"`
}

// KnowledgeStats represents statistics about the knowledge base
type KnowledgeStats struct {
    TotalDocuments    int   `json:"total_documents"`
    TotalChunks       int   `json:"total_chunks"`
    TotalScrapingJobs int   `json:"total_scraping_jobs"`
    TotalScrapedPages int   `json:"total_scraped_pages"`
    TotalStorageBytes int64 `json:"total_storage_bytes"`
}

// KnowledgeFAQItem represents an auto-generated FAQ entry for the knowledge base
type KnowledgeFAQItem struct {
    ID            uuid.UUID `db:"id" json:"id"`
    TenantID      uuid.UUID `db:"tenant_id" json:"tenant_id"`
    ProjectID     uuid.UUID `db:"project_id" json:"project_id"`
    Question      string    `db:"question" json:"question"`
    Answer        string    `db:"answer" json:"answer"`
    SourceURL     *string   `db:"source_url" json:"source_url,omitempty"`
    SourceSection *string   `db:"source_section" json:"source_section,omitempty"`
    Metadata      JSONMap   `db:"metadata" json:"metadata"`
    CreatedAt     time.Time `db:"created_at" json:"created_at"`
    UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// DocumentProcessingStatus represents the processing status of a document
type DocumentProcessingStatus struct {
	DocumentID      uuid.UUID `json:"document_id"`
	Status          string    `json:"status"`
	Progress        float64   `json:"progress"` // 0.0 to 1.0
	ProcessedChunks int       `json:"processed_chunks"`
	TotalChunks     int       `json:"total_chunks"`
	ErrorMessage    *string   `json:"error_message,omitempty"`
}

// Embedding request/response for external services
type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// Helper functions for pgvector operations

// NewVectorFromFloat32Slice creates a pgvector.Vector from []float32
func NewVectorFromFloat32Slice(slice []float32) pgvector.Vector {
	return pgvector.NewVector(slice)
}

// VectorToFloat32Slice converts pgvector.Vector to []float32
func VectorToFloat32Slice(v pgvector.Vector) []float32 {
	return v.Slice()
}

// CosineSimilarity calculates cosine similarity between two vectors
// Note: This is for client-side calculations. Database queries should use vector_cosine_ops
func CosineSimilarity(a, b pgvector.Vector) float64 {
	sliceA := a.Slice()
	sliceB := b.Slice()

	if len(sliceA) != len(sliceB) {
		return 0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(sliceA); i++ {
		dotProduct += float64(sliceA[i]) * float64(sliceB[i])
		normA += float64(sliceA[i]) * float64(sliceA[i])
		normB += float64(sliceB[i]) * float64(sliceB[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
