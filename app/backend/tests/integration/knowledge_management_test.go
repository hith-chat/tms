package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/bareuptime/tms/internal/config"
	"github.com/bareuptime/tms/internal/db"
	"github.com/bareuptime/tms/internal/models"
)

// Test database setup
func setupTestDB(t *testing.T) *db.DB {
	cfg := &config.DatabaseConfig{
		URL: "postgres://tms:tms123@localhost:5432/tms?sslmode=disable",
	}
	
	database, err := db.Connect(cfg)
	require.NoError(t, err)
	
	return database
}

func teardownTestDB(database *db.DB) {
	if database != nil {
		database.Close()
	}
}

func stringPtr(s string) *string {
	return &s
}

// Test knowledge document creation
func TestKnowledgeDocumentCreation(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(database)
	
	// Clean up first
	database.Exec("DELETE FROM knowledge_documents WHERE filename = 'integration_test.txt'")
	
	// Test data
	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	projectID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	
	// Create document
	doc := &models.KnowledgeDocument{
		ID:               uuid.New(),
		TenantID:         tenantID,
		ProjectID:        projectID,
		Filename:         "integration_test.txt",
		ContentType:      "text/plain",
		FileSize:         65,
		FilePath:         "/tmp/integration_test.txt",
		OriginalContent:  stringPtr("This is content for integration testing of knowledge management."),
		ProcessedContent: stringPtr("This is content for integration testing of knowledge management."),
		Status:           "processing",
		Metadata:         models.JSONMap{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	
	query := `
		INSERT INTO knowledge_documents (id, tenant_id, project_id, filename, content_type, file_size, file_path, original_content, processed_content, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	
	metadataJSON, err := json.Marshal(doc.Metadata)
	require.NoError(t, err)
	
	_, err = database.Exec(query, doc.ID, doc.TenantID, doc.ProjectID, doc.Filename, doc.ContentType, 
		doc.FileSize, doc.FilePath, doc.OriginalContent, doc.ProcessedContent, doc.Status, string(metadataJSON), doc.CreatedAt, doc.UpdatedAt)
	assert.NoError(t, err)
	
	// Verify document was created
	var retrievedDoc models.KnowledgeDocument
	var metadataStr string
	selectQuery := `
		SELECT id, tenant_id, project_id, filename, content_type, file_size, status, metadata
		FROM knowledge_documents WHERE id = $1
	`
	
	err = database.QueryRow(selectQuery, doc.ID).Scan(
		&retrievedDoc.ID, &retrievedDoc.TenantID, &retrievedDoc.ProjectID,
		&retrievedDoc.Filename, &retrievedDoc.ContentType, &retrievedDoc.FileSize,
		&retrievedDoc.Status, &metadataStr,
	)
	assert.NoError(t, err)
	assert.Equal(t, doc.ID, retrievedDoc.ID)
	assert.Equal(t, doc.Filename, retrievedDoc.Filename)
	assert.Equal(t, doc.Status, retrievedDoc.Status)
	
	// Cleanup
	database.Exec("DELETE FROM knowledge_documents WHERE id = $1", doc.ID)
}

// Test knowledge chunks creation
func TestKnowledgeChunksCreation(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(database)
	
	// Clean up first
	database.Exec("DELETE FROM knowledge_chunks WHERE content LIKE '%chunk of content%'")
	database.Exec("DELETE FROM knowledge_documents WHERE filename = 'test_chunks.txt'")
	
	// First create a parent document
	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	projectID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	
	doc := &models.KnowledgeDocument{
		ID:               uuid.New(),
		TenantID:         tenantID,
		ProjectID:        projectID,
		Filename:         "test_chunks.txt",
		ContentType:      "text/plain",
		FileSize:         45,
		FilePath:         "/tmp/test_chunks.txt",
		OriginalContent:  stringPtr("This is test content for knowledge management testing."),
		ProcessedContent: stringPtr("This is test content for knowledge management testing."),
		Status:           "completed",
		Metadata:         models.JSONMap{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	
	docQuery := `
		INSERT INTO knowledge_documents (id, tenant_id, project_id, filename, content_type, file_size, file_path, original_content, processed_content, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	
	metadataJSON, err := json.Marshal(doc.Metadata)
	require.NoError(t, err)
	
	_, err = database.Exec(docQuery, doc.ID, doc.TenantID, doc.ProjectID, doc.Filename, doc.ContentType, 
		doc.FileSize, doc.FilePath, doc.OriginalContent, doc.ProcessedContent, doc.Status, string(metadataJSON), doc.CreatedAt, doc.UpdatedAt)
	require.NoError(t, err)
	
	// Create chunks
	chunks := []models.KnowledgeChunk{
		{
			ID:         uuid.New(),
			DocumentID: doc.ID,
			ChunkIndex: 0,
			Content:    "This is the first chunk of content.",
			TokenCount: 8,
			Metadata:   models.JSONMap{},
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			DocumentID: doc.ID,
			ChunkIndex: 1,
			Content:    "This is the second chunk of content.",
			TokenCount: 8,
			Metadata:   models.JSONMap{},
			CreatedAt:  time.Now(),
		},
	}
	
	// Insert chunks
	for _, chunk := range chunks {
		query := `
			INSERT INTO knowledge_chunks (id, document_id, chunk_index, content, token_count, metadata, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		
		chunkMetadataJSON, err := json.Marshal(chunk.Metadata)
		require.NoError(t, err)
		
		_, err = database.Exec(query, chunk.ID, chunk.DocumentID, chunk.ChunkIndex,
			chunk.Content, chunk.TokenCount, string(chunkMetadataJSON), chunk.CreatedAt)
		assert.NoError(t, err)
	}
	
	// Verify chunks were created
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM knowledge_chunks WHERE document_id = $1", doc.ID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	
	// Cleanup
	database.Exec("DELETE FROM knowledge_chunks WHERE document_id = $1", doc.ID)
	database.Exec("DELETE FROM knowledge_documents WHERE id = $1", doc.ID)
}

// Test web scraping job lifecycle
func TestWebScrapingJobLifecycle(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(database)
	
	// Clean up first
	database.Exec("DELETE FROM knowledge_scraping_jobs WHERE url = 'https://example.com'")
	
	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	projectID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	
	// Create scraping job
	job := &models.KnowledgeScrapingJob{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ProjectID: projectID,
		URL:       "https://example.com",
		MaxDepth:  2,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	query := `
		INSERT INTO knowledge_scraping_jobs (id, tenant_id, project_id, url, max_depth, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	_, err := database.Exec(query, job.ID, job.TenantID, job.ProjectID, job.URL, 
		job.MaxDepth, job.Status, job.CreatedAt, job.UpdatedAt)
	assert.NoError(t, err)
	
	// Update job status to completed
	updateQuery := `
		UPDATE knowledge_scraping_jobs 
		SET status = $1, completed_at = $2, pages_scraped = $3, updated_at = $4 
		WHERE id = $5
	`
	completedAt := time.Now()
	_, err = database.Exec(updateQuery, "completed", completedAt, 2, time.Now(), job.ID)
	assert.NoError(t, err)
	
	// Verify job completion
	var retrievedJob models.KnowledgeScrapingJob
	selectQuery := `
		SELECT id, status, pages_scraped, url, max_depth
		FROM knowledge_scraping_jobs WHERE id = $1
	`
	
	err = database.QueryRow(selectQuery, job.ID).Scan(
		&retrievedJob.ID, &retrievedJob.Status, &retrievedJob.PagesScraped,
		&retrievedJob.URL, &retrievedJob.MaxDepth,
	)
	assert.NoError(t, err)
	assert.Equal(t, "completed", retrievedJob.Status)
	assert.Equal(t, 2, retrievedJob.PagesScraped)
	
	// Cleanup
	database.Exec("DELETE FROM knowledge_scraping_jobs WHERE id = $1", job.ID)
}

// Test pgvector extension is working
func TestPgVectorExtension(t *testing.T) {
	database := setupTestDB(t)
	defer teardownTestDB(database)
	
	// Test that pgvector extension is available
	var hasVectorExtension bool
	err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&hasVectorExtension)
	assert.NoError(t, err)
	assert.True(t, hasVectorExtension, "pgvector extension should be installed")
}
