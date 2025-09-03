package fixtures

import (
	"time"

	"github.com/bareuptime/tms/internal/models"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// MockTenantID for testing
var MockTenantID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
var MockProjectID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
var MockUserID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174002")
var MockDocumentID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174003")
var MockJobID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174004")

// MockKnowledgeDocument creates a mock knowledge document for testing
func MockKnowledgeDocument() *models.KnowledgeDocument {
	return &models.KnowledgeDocument{
		ID:          MockDocumentID,
		TenantID:    MockTenantID,
		ProjectID:   MockProjectID,
		Filename:    "test-document.pdf",
		ContentType: "application/pdf",
		FileSize:    1024000, // 1MB
		FilePath:    "/uploads/test-document.pdf",
		Status:      "completed",
		Metadata:    models.JSONMap{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// MockKnowledgeChunk creates a mock knowledge chunk for testing
func MockKnowledgeChunk() *models.KnowledgeChunk {
	// Create a mock embedding vector (1536 dimensions for OpenAI)
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = 0.1 // Simple mock values
	}

	// Create the vector and get its address
	vector := pgvector.NewVector(embedding)

	return &models.KnowledgeChunk{
		ID:         uuid.New(),
		DocumentID: MockDocumentID,
		ChunkIndex: 0,
		Content:    "This is a test chunk of content from a PDF document.",
		TokenCount: 12,
		Embedding:  &vector,
		Metadata:   models.JSONMap{},
		CreatedAt:  time.Now(),
	}
}

// MockScrapingJob creates a mock scraping job for testing
func MockScrapingJob() *models.KnowledgeScrapingJob {
	return &models.KnowledgeScrapingJob{
		ID:           MockJobID,
		TenantID:     MockTenantID,
		ProjectID:    MockProjectID,
		URL:          "https://example.com",
		MaxDepth:     3,
		Status:       "completed",
		PagesScraped: 5,
		TotalPages:   5,
		StartedAt:    &time.Time{},
		CompletedAt:  &time.Time{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// MockScrapedPage creates a mock scraped page for testing
func MockScrapedPage() *models.KnowledgeScrapedPage {
	// Create a mock embedding vector
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = 0.2 // Simple mock values
	}

	// Create the vector and get its address
	vector := pgvector.NewVector(embedding)

	title := "Test Page Title"
	return &models.KnowledgeScrapedPage{
		ID:        uuid.New(),
		JobID:     MockJobID,
		URL:       "https://example.com/page1",
		Title:     &title,
		Content:   "This is test content from a scraped web page.",
		ScrapedAt: time.Now(),
		Embedding: &vector,
		Metadata:  models.JSONMap{},
	}
}

// MockUploadRequest creates a mock upload request
func MockUploadRequest() *models.UploadDocumentRequest {
	return &models.UploadDocumentRequest{
		Filename:    "test-upload.pdf",
		ContentType: "application/pdf",
	}
}

// MockScrapingRequest creates a mock scraping request
func MockScrapingRequest() *models.CreateScrapingJobRequest {
	return &models.CreateScrapingJobRequest{
		URL:      "https://example.com",
		MaxDepth: 3,
	}
}

// MockSearchRequest creates a mock search request
func MockSearchRequest() *models.KnowledgeSearchRequest {
	return &models.KnowledgeSearchRequest{
		Query:            "test search query",
		MaxResults:       10,
		SimilarityScore:  0.7,
		IncludeDocuments: true,
		IncludePages:     true,
	}
}

// MockSearchResult creates a mock search result
func MockSearchResult() *models.KnowledgeSearchResponse {
	return &models.KnowledgeSearchResponse{
		Results: []models.KnowledgeSearchResult{
			{
				ID:         uuid.New(),
				Type:       "document",
				Content:    "This is a test search result",
				Score:      0.95,
				Source:     "test-document.pdf",
				Title:      nil,
				DocumentID: &MockDocumentID,
				Metadata:   models.JSONMap{},
			},
		},
		TotalCount:  1,
		Query:       "test search query",
		ProcessedIn: "10ms",
	}
}

// TestPDFContent provides sample PDF content for testing
var TestPDFContent = `This is a sample PDF document content for testing purposes.
It contains multiple paragraphs and sections to test text extraction and chunking.

Section 1: Introduction
This is the introduction section of the test document.

Section 2: Main Content
This section contains the main content of the document.
It has multiple sentences to test the chunking algorithm.

Section 3: Conclusion
This is the conclusion of the test document.`

// TestWebContent provides sample web content for testing
var TestWebContent = `<!DOCTYPE html>
<html>
<head>
    <title>Test Web Page</title>
</head>
<body>
    <h1>Test Page Title</h1>
    <p>This is a test web page content for scraping tests.</p>
    <div>
        <p>More content in a div element.</p>
        <a href="/page2">Link to page 2</a>
        <a href="/page3">Link to page 3</a>
    </div>
</body>
</html>`

// MockEmbedding creates a mock embedding vector
func MockEmbedding() []float32 {
	embedding := make([]float32, 1536)
	for i := range embedding {
		embedding[i] = float32(i) * 0.001 // Create varied values
	}
	return embedding
}

// TestChunks provides sample text chunks for testing
var TestChunks = []string{
	"This is the first chunk of test content.",
	"This is the second chunk with different content.",
	"The third chunk contains more test data for validation.",
	"Final chunk to complete the test dataset.",
}

// MockDocumentList creates a list of mock documents
func MockDocumentList() []models.KnowledgeDocument {
	return []models.KnowledgeDocument{
		*MockKnowledgeDocument(),
		{
			ID:          uuid.New(),
			TenantID:    MockTenantID,
			ProjectID:   MockProjectID,
			Filename:    "second-document.pdf",
			ContentType: "application/pdf",
			FileSize:    2048000,
			FilePath:    "/uploads/second-document.pdf",
			Status:      "processing",
			Metadata:    models.JSONMap{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}

// DatabaseTestConfig provides test database configuration
type DatabaseTestConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func TestDatabaseConfig() *DatabaseTestConfig {
	return &DatabaseTestConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "tms_test",
		SSLMode:  "disable",
	}
}

// Test file paths
const (
	TestPDFFilePath  = "sample_files/test.pdf"
	TestTextFilePath = "sample_files/test.txt"
	TestLargePDFPath = "sample_files/large_test.pdf"
)

// Error test cases
var (
	ErrInvalidFile    = "invalid file format"
	ErrFileTooLarge   = "file size exceeds limit"
	ErrInvalidURL     = "invalid URL format"
	ErrNetworkTimeout = "network timeout"
	ErrUnauthorized   = "unauthorized access"
)
