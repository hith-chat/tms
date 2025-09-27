package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bareuptime/tms/internal/middleware"
	"github.com/bareuptime/tms/internal/models"
	"github.com/bareuptime/tms/internal/service"
)

type KnowledgeHandler struct {
	documentProcessor *service.DocumentProcessorService
	webScraper        *service.WebScrapingService
	knowledgeService  *service.KnowledgeService
	publicURLAnalysis *service.PublicURLAnalysisService
}

func NewKnowledgeHandler(
	documentProcessor *service.DocumentProcessorService,
	webScraper *service.WebScrapingService,
	knowledgeService *service.KnowledgeService,
	publicURLAnalysis *service.PublicURLAnalysisService,
) *KnowledgeHandler {
	return &KnowledgeHandler{
		documentProcessor: documentProcessor,
		webScraper:        webScraper,
		knowledgeService:  knowledgeService,
		publicURLAnalysis: publicURLAnalysis,
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

// CreateScrapingJobWithStream creates a new web scraping job and streams progress
func (h *KnowledgeHandler) CreateScrapingJobWithStream(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	var req models.CreateScrapingJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default max depth if not specified
	if req.MaxDepth == 0 {
		req.MaxDepth = 1
	}

	events := make(chan service.ScrapingEvent)
	ctx := c.Request.Context()

	// Create scraping job with streaming
	job, err := h.webScraper.CreateScrapingJobWithStream(ctx, tenantID, projectID, &req, events)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	// First send the job creation response
	if _, err := fmt.Fprintf(c.Writer, "event: job_created\ndata: %s\n\n", toJSON(job)); err != nil {
		return
	}
	c.Writer.Flush()

	// Stream scraping progress
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case event, ok := <-events:
			if !ok {
				return false
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, toJSON(event)); err != nil {
				return false
			}
			return true
		}
	})
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

/*
// getJobId is not needed, so it is removed.
*/

// GetScrapingJob returns a specific scraping job
func (h *KnowledgeHandler) GetScrapingJob(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	jobIDStr := c.Param("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// Get scraping job
	job, err := h.webScraper.GetScrapingJob(c.Request.Context(), jobID, tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scraping job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// GetJobPages returns pages scraped by a job
func (h *KnowledgeHandler) GetJobPages(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	jobIDStr := c.Param("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// Get pages
	pages, err := h.webScraper.GetJobPages(c.Request.Context(), jobID, tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scraped pages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pages": pages})
}

// GetScrapingJobLinks returns staged links awaiting user confirmation
func (h *KnowledgeHandler) GetScrapingJobLinks(c *gin.Context) {
	projectID := middleware.GetProjectID(c)
	tenantID := middleware.GetTenantID(c)
	jobIDStr := c.Param("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	links, err := h.webScraper.GetStagedLinks(c.Request.Context(), jobID, tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"links":                links,
		"max_selectable_links": service.MaxSelectableLinks,
	})
}

// SelectScrapingJobLinks stores the subset of links that should be indexed
func (h *KnowledgeHandler) SelectScrapingJobLinks(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)
	jobIDStr := c.Param("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := h.webScraper.GetScrapingJob(c.Request.Context(), jobID, tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scraping job not found"})
		return
	}

	if job.ProjectID != projectID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access to this scraping job is not allowed"})
		return
	}

	var req models.SelectScrapingLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.webScraper.StoreLinkSelection(c.Request.Context(), jobID, tenantID, projectID, req.URLs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedJob, jobErr := h.webScraper.GetScrapingJob(c.Request.Context(), jobID, tenantID, projectID)
	selectedCount := len(req.URLs)
	if jobErr == nil {
		selectedCount = len(updatedJob.SelectedLinks)
	}

	c.JSON(http.StatusOK, gin.H{
		"selected_count":       selectedCount,
		"message":              "Link selection saved",
		"max_selectable_links": service.MaxSelectableLinks,
	})
}

// StreamScrapingJobIndex streams indexing progress using server-sent events
func (h *KnowledgeHandler) StreamScrapingJobIndex(c *gin.Context) {
	projectID := middleware.GetProjectID(c)
	tenantID := middleware.GetTenantID(c)
	jobIDStr := c.Param("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := h.webScraper.GetScrapingJob(c.Request.Context(), jobID, tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scraping job not found"})
		return
	}

	if job.ProjectID != projectID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access to this scraping job is not allowed"})
		return
	}

	events := make(chan service.IndexingEvent)
	ctx := c.Request.Context()

	go func() {
		defer close(events)
		if err := h.webScraper.StreamIndexing(ctx, jobID, tenantID, projectID, events); err != nil {
			select {
			case <-ctx.Done():
			case events <- service.IndexingEvent{Type: "error", Message: err.Error(), Timestamp: time.Now()}:
			}
		}
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case event, ok := <-events:
			if !ok {
				return false
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, toJSON(event)); err != nil {
				return false
			}
			return true
		}
	})
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

// ListFAQItems returns auto-generated FAQ pairs for the project
func (h *KnowledgeHandler) ListFAQItems(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	projectID := middleware.GetProjectID(c)

	items, err := h.knowledgeService.ListFAQItems(c.Request.Context(), tenantID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load knowledge FAQs"})
		return
	}

	if items == nil {
		items = []*models.KnowledgeFAQItem{}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
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

// AnalyzePublicURL analyzes a public URL and returns associated URLs with token counts
// @Summary Analyze public URL
// @Description Analyze a public URL to discover associated URLs and their token counts up to specified depth
// @Tags Public
// @Accept json
// @Produce text/event-stream
// @Param request body service.URLAnalysisRequest true "URL analysis request"
// @Success 200 {string} string "Server-sent events stream with analysis progress"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/analyze-url [post]
func (h *KnowledgeHandler) AnalyzePublicURL(c *gin.Context) {
	var req service.URLAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate the request
	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	// Create events channel
	events := make(chan service.URLAnalysisEvent, 100)
	defer close(events)

	// Start analysis in goroutine
	go func() {
		_, err := h.publicURLAnalysis.AnalyzeURLWithStream(ctx, req, events)
		if err != nil {
			// Send error event if analysis fails
			select {
			case events <- service.URLAnalysisEvent{
				Type:      "error",
				Message:   fmt.Sprintf("Analysis failed: %v", err),
				Timestamp: time.Now(),
			}:
			case <-ctx.Done():
				// Context cancelled
			}
		}
	}()

	// Set up Server-Sent Events headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// Stream events
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-events:
			if !ok {
				return false
			}

			eventData := toJSON(event)
			fmt.Fprintf(w, "data: %s\n\n", eventData)

			// Close stream if analysis is completed or errored
			if event.Type == "completed" || event.Type == "error" {
				return false
			}

			return true
		case <-ctx.Done():
			// Context cancelled, close stream
			return false
		case <-time.After(30 * time.Second):
			// Send keepalive ping
			fmt.Fprintf(w, "data: %s\n\n", toJSON(service.URLAnalysisEvent{
				Type:      "ping",
				Message:   "keepalive",
				Timestamp: time.Now(),
			}))
			return true
		}
	})
}

// toJSON helper function for SSE data formatting
func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
