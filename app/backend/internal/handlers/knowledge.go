package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
)

type KnowledgeHandler struct {
	documentProcessor *service.DocumentProcessorService
	webScraper       *service.WebScrapingService
	knowledgeService *service.KnowledgeService
}

func NewKnowledgeHandler(
	documentProcessor *service.DocumentProcessorService,
	webScraper *service.WebScrapingService,
	knowledgeService *service.KnowledgeService,
) *KnowledgeHandler {
	return &KnowledgeHandler{
		documentProcessor: documentProcessor,
		webScraper:       webScraper,
		knowledgeService: knowledgeService,
	}
}

// Document endpoints

// UploadDocument handles document upload
func (h *KnowledgeHandler) UploadDocument(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from request"})
		return
	}
	defer file.Close()

	// Process document
	doc, err := h.documentProcessor.ProcessDocument(c.Request.Context(), tenantID, projectID, file, header)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// ListDocuments returns a list of documents
func (h *KnowledgeHandler) ListDocuments(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	// Parse query parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get documents
	documents, err := h.documentProcessor.ListDocuments(c.Request.Context(), tenantID, projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list documents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetDocument returns a specific document
func (h *KnowledgeHandler) GetDocument(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// Get document processing status
	status, err := h.documentProcessor.GetDocumentProcessingStatus(c.Request.Context(), documentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// DeleteDocument deletes a document
func (h *KnowledgeHandler) DeleteDocument(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// Delete document
	if err := h.documentProcessor.DeleteDocument(c.Request.Context(), documentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

// Web scraping endpoints

// CreateScrapingJob creates a new web scraping job
func (h *KnowledgeHandler) CreateScrapingJob(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	var req models.CreateScrapingJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default max depth if not specified
	if req.MaxDepth == 0 {
		req.MaxDepth = 3
	}

	// Create scraping job
	job, err := h.webScraper.CreateScrapingJob(c.Request.Context(), tenantID, projectID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// ListScrapingJobs returns a list of scraping jobs
func (h *KnowledgeHandler) ListScrapingJobs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	// Parse query parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get scraping jobs
	jobs, err := h.webScraper.ListScrapingJobs(c.Request.Context(), tenantID, projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list scraping jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":   jobs,
		"limit":  limit,
		"offset": offset,
	})
}

// GetScrapingJob returns a specific scraping job
func (h *KnowledgeHandler) GetScrapingJob(c *gin.Context) {
	jobIDStr := c.Param("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// Get scraping job
	job, err := h.webScraper.GetScrapingJob(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scraping job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// GetJobPages returns pages scraped by a job
func (h *KnowledgeHandler) GetJobPages(c *gin.Context) {
	jobIDStr := c.Param("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// Get pages
	pages, err := h.webScraper.GetJobPages(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scraped pages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pages": pages})
}

// Knowledge search endpoints

// SearchKnowledgeBase searches the knowledge base
func (h *KnowledgeHandler) SearchKnowledgeBase(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	var req models.KnowledgeSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.MaxResults == 0 {
		req.MaxResults = 10
	}
	if req.SimilarityScore == 0 {
		req.SimilarityScore = 0.7
	}
	if !req.IncludeDocuments && !req.IncludePages {
		req.IncludeDocuments = true
		req.IncludePages = true
	}

	// Search knowledge base
	response, err := h.knowledgeService.SearchKnowledgeBase(c.Request.Context(), tenantID, projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search knowledge base"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET version of search for easier testing
func (h *KnowledgeHandler) SearchKnowledgeBaseGET(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// Parse optional parameters
	maxResults := 10
	if maxStr := c.Query("max_results"); maxStr != "" {
		if m, err := strconv.Atoi(maxStr); err == nil && m > 0 && m <= 50 {
			maxResults = m
		}
	}

	similarityScore := 0.7
	if scoreStr := c.Query("similarity_score"); scoreStr != "" {
		if s, err := strconv.ParseFloat(scoreStr, 64); err == nil && s >= 0 && s <= 1 {
			similarityScore = s
		}
	}

	includeDocuments := c.Query("include_documents") != "false"
	includePages := c.Query("include_pages") != "false"

	req := &models.KnowledgeSearchRequest{
		Query:            query,
		MaxResults:       maxResults,
		SimilarityScore:  similarityScore,
		IncludeDocuments: includeDocuments,
		IncludePages:     includePages,
	}

	// Search knowledge base
	response, err := h.knowledgeService.SearchKnowledgeBase(c.Request.Context(), tenantID, projectID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search knowledge base"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Settings endpoints

// GetKnowledgeSettings returns knowledge settings for a project
func (h *KnowledgeHandler) GetKnowledgeSettings(c *gin.Context) {
	projectID := middleware.GetProjectID(c)

	settings, err := h.knowledgeService.GetKnowledgeSettings(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Knowledge settings not found"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateKnowledgeSettings updates knowledge settings for a project
func (h *KnowledgeHandler) UpdateKnowledgeSettings(c *gin.Context) {
	projectID := middleware.GetProjectID(c)

	var req models.UpdateKnowledgeSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.knowledgeService.UpdateKnowledgeSettings(c.Request.Context(), projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update knowledge settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// Stats endpoints

// GetKnowledgeStats returns knowledge base statistics
func (h *KnowledgeHandler) GetKnowledgeStats(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	stats, err := h.knowledgeService.GetKnowledgeStats(c.Request.Context(), tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get knowledge stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
