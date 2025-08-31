package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"

	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/repo"
)

type DocumentProcessorService struct {
	knowledgeRepo   *repo.KnowledgeRepository
	embeddingService *EmbeddingService
	uploadDir       string
	maxFileSize     int64
}

func NewDocumentProcessorService(knowledgeRepo *repo.KnowledgeRepository, embeddingService *EmbeddingService, uploadDir string, maxFileSize int64) *DocumentProcessorService {
	return &DocumentProcessorService{
		knowledgeRepo:   knowledgeRepo,
		embeddingService: embeddingService,
		uploadDir:       uploadDir,
		maxFileSize:     maxFileSize,
	}
}

// ProcessDocument handles the entire document processing pipeline
func (s *DocumentProcessorService) ProcessDocument(ctx context.Context, tenantID, projectID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*models.KnowledgeDocument, error) {
	// Validate file
	if err := s.validateFile(header); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Create document record
	doc := &models.KnowledgeDocument{
		ID:          uuid.New(),
		TenantID:    tenantID,
		ProjectID:   projectID,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		FileSize:    header.Size,
		Status:      "processing",
		Metadata:    models.JSONMap{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save file to disk
	filePath, err := s.saveFile(file, doc.ID.String(), header.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}
	doc.FilePath = filePath

	// Create document record in database
	if err := s.knowledgeRepo.CreateDocument(doc); err != nil {
		// Clean up saved file
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	// Process document asynchronously
	go s.processDocumentAsync(ctx, doc)

	return doc, nil
}

// validateFile validates the uploaded file
func (s *DocumentProcessorService) validateFile(header *multipart.FileHeader) error {
	// Check file size
	if header.Size > s.maxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", header.Size, s.maxFileSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" {
		return fmt.Errorf("unsupported file type: %s. Only PDF files are supported", ext)
	}

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "pdf") {
		return fmt.Errorf("invalid content type: %s. Expected PDF", contentType)
	}

	return nil
}

// saveFile saves the uploaded file to disk
func (s *DocumentProcessorService) saveFile(file multipart.File, docID, filename string) (string, error) {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	safeFilename := fmt.Sprintf("%s%s", docID, ext)
	filePath := filepath.Join(s.uploadDir, safeFilename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	return filePath, nil
}

// processDocumentAsync processes the document content asynchronously
func (s *DocumentProcessorService) processDocumentAsync(ctx context.Context, doc *models.KnowledgeDocument) {
	var err error
	defer func() {
		status := "completed"
		var errorMessage *string
		
		if err != nil {
			status = "failed"
			errStr := err.Error()
			errorMessage = &errStr
		}
		
		s.knowledgeRepo.UpdateDocumentStatus(doc.ID, status, errorMessage)
	}()

	// Extract text content
	content, err := s.extractTextFromPDF(doc.FilePath)
	if err != nil {
		err = fmt.Errorf("failed to extract text from PDF: %w", err)
		return
	}

	// Update document with processed content
	if err = s.knowledgeRepo.UpdateDocumentContent(doc.ID, content); err != nil {
		err = fmt.Errorf("failed to update document content: %w", err)
		return
	}

	// Get knowledge settings for chunking parameters
	settings, settingsErr := s.knowledgeRepo.GetSettings(doc.ProjectID)
	if settingsErr != nil {
		// Use default settings if not found
		settings = &models.KnowledgeSettings{
			ChunkSize:    1000,
			ChunkOverlap: 200,
		}
	}

	// Create chunks
	chunks, err := s.createChunks(content, settings.ChunkSize, settings.ChunkOverlap)
	if err != nil {
		err = fmt.Errorf("failed to create chunks: %w", err)
		return
	}

	// Generate embeddings for chunks
	knowledgeChunks := make([]*models.KnowledgeChunk, 0, len(chunks))
	for i, chunk := range chunks {
		knowledgeChunk := &models.KnowledgeChunk{
			ID:         uuid.New(),
			DocumentID: doc.ID,
			ChunkIndex: i,
			Content:    chunk.Content,
			TokenCount: chunk.TokenCount,
			Metadata:   models.JSONMap{},
			CreatedAt:  time.Now(),
		}

		// Try to generate embedding, but don't fail if it doesn't work
		if s.embeddingService.IsEnabled() {
			embedding, embErr := s.embeddingService.GenerateEmbedding(ctx, chunk.Content)
			if embErr != nil {
				fmt.Printf("Warning: Failed to generate embedding for chunk %d: %v\n", i, embErr)
				// Continue without embedding
			} else {
				knowledgeChunk.Embedding = &embedding
			}
		} else {
			fmt.Printf("Warning: Embedding service is disabled, saving chunk without embedding\n")
		}

		knowledgeChunks = append(knowledgeChunks, knowledgeChunk)
	}

	// Save chunks to database
	if err = s.knowledgeRepo.CreateChunks(knowledgeChunks); err != nil {
		err = fmt.Errorf("failed to save chunks: %w", err)
		return
	}
}

// extractTextFromPDF extracts text content from a PDF file
func (s *DocumentProcessorService) extractTextFromPDF(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	pdfReader, err := pdf.NewReader(file, 0)
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	var content strings.Builder
	
	// Extract text from all pages
	for i := 1; i <= pdfReader.NumPage(); i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			// Log error but continue with other pages
			continue
		}

		content.WriteString(text)
		content.WriteString("\n\n") // Add page separator
	}

	textContent := content.String()
	if len(strings.TrimSpace(textContent)) == 0 {
		return "", fmt.Errorf("no text content found in PDF")
	}

	return textContent, nil
}

// TextChunk represents a chunk of text with metadata
type TextChunk struct {
	Content    string
	TokenCount int
	StartPos   int
	EndPos     int
}

// createChunks splits text into overlapping chunks
func (s *DocumentProcessorService) createChunks(text string, chunkSize, overlap int) ([]*TextChunk, error) {
	if len(text) == 0 {
		return nil, fmt.Errorf("text content is empty")
	}

	var chunks []*TextChunk
	textLen := len(text)
	
	for start := 0; start < textLen; {
		end := start + chunkSize
		if end > textLen {
			end = textLen
		}

		// Try to find a good breaking point (end of sentence or paragraph)
		if end < textLen {
			// Look for sentence endings
			for i := end; i > start+chunkSize/2; i-- {
				if i < textLen && (text[i] == '.' || text[i] == '!' || text[i] == '?' || text[i] == '\n') {
					end = i + 1
					break
				}
			}
		}

		chunkText := strings.TrimSpace(text[start:end])
		if len(chunkText) > 0 {
			chunk := &TextChunk{
				Content:    chunkText,
				TokenCount: s.estimateTokenCount(chunkText),
				StartPos:   start,
				EndPos:     end,
			}
			chunks = append(chunks, chunk)
		}

		// Calculate next start position with overlap
		nextStart := end - overlap
		if nextStart <= start {
			nextStart = start + chunkSize/2 // Ensure we make progress
		}
		start = nextStart

		// Break if we've reached the end
		if end >= textLen {
			break
		}
	}

	return chunks, nil
}

// estimateTokenCount provides a rough estimate of token count
func (s *DocumentProcessorService) estimateTokenCount(text string) int {
	// Rough approximation: 1 token ≈ 4 characters for English text
	// This is a simplified approach; for more accuracy, use tiktoken
	return len(text) / 4
}

// GetDocumentProcessingStatus returns the processing status of a document
func (s *DocumentProcessorService) GetDocumentProcessingStatus(ctx context.Context, documentID uuid.UUID) (*models.DocumentProcessingStatus, error) {
	doc, err := s.knowledgeRepo.GetDocument(documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	chunks, err := s.knowledgeRepo.GetDocumentChunks(documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document chunks: %w", err)
	}

	status := &models.DocumentProcessingStatus{
		DocumentID:      documentID,
		Status:          doc.Status,
		ProcessedChunks: len(chunks),
		ErrorMessage:    doc.ErrorMessage,
	}

	// Calculate progress
	if doc.Status == "completed" {
		status.Progress = 1.0
		status.TotalChunks = len(chunks)
	} else if doc.Status == "processing" && doc.ProcessedContent != nil {
		// Estimate total chunks based on content length
		estimatedChunks := len(*doc.ProcessedContent) / 1000 // Assuming 1000 chars per chunk
		if estimatedChunks > 0 {
			status.Progress = float64(len(chunks)) / float64(estimatedChunks)
			status.TotalChunks = estimatedChunks
		}
	}

	return status, nil
}

// DeleteDocument removes a document and all its associated data
func (s *DocumentProcessorService) DeleteDocument(ctx context.Context, documentID uuid.UUID) error {
	// Get document info first
	doc, err := s.knowledgeRepo.GetDocument(documentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// Delete from database (chunks will be deleted due to CASCADE)
	if err := s.knowledgeRepo.DeleteDocument(documentID); err != nil {
		return fmt.Errorf("failed to delete document from database: %w", err)
	}

	// Delete file from disk
	if doc.FilePath != "" {
		if err := os.Remove(doc.FilePath); err != nil {
			// Log error but don't fail the operation
			// The file might have already been deleted or moved
		}
	}

	return nil
}

// ListDocuments returns a list of documents for a project
func (s *DocumentProcessorService) ListDocuments(ctx context.Context, tenantID, projectID uuid.UUID, limit, offset int) ([]*models.KnowledgeDocument, error) {
	return s.knowledgeRepo.ListDocuments(tenantID, projectID, limit, offset)
}
