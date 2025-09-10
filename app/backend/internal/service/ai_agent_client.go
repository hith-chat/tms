package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// AiAgentClient handles communication with the Python agent service
type AiAgentClient struct {
	baseURL    string
	httpClient *http.Client
}

// ChatRequest represents a request to the Python agent service
type ChatRequest struct {
	Message   string            `json:"message"`
	TenantID  string            `json:"tenant_id"`
	ProjectID string            `json:"project_id"`
	SessionID string            `json:"session_id"`
	UserID    string            `json:"user_id,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// AgentResponse represents a response from the agent service
type AgentResponse struct {
	Type     string            `json:"type"`
	Content  string            `json:"content,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewAgentClient creates a new agent client
func NewAgentClient(agentUrl string) *AiAgentClient {
	fmt.Println("Initializing AiAgentClient with URL:", agentUrl)
	return &AiAgentClient{
		baseURL: agentUrl,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessMessageStream sends a message to the agent service and returns a channel for streaming responses
func (ac *AiAgentClient) ProcessMessageStream(ctx context.Context, req ChatRequest) (<-chan AgentResponse, <-chan error) {
	responseChan := make(chan AgentResponse, 100)
	errorChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errorChan)

		// Prepare request
		reqBody, err := json.Marshal(req)
		if err != nil {
			errorChan <- fmt.Errorf("error marshaling request: %w", err)
			return
		}

		// Create HTTP request
		httpReq, err := http.NewRequestWithContext(ctx, "POST", ac.baseURL+"/chat/process", bytes.NewReader(reqBody))
		if err != nil {
			errorChan <- fmt.Errorf("error creating request: %w", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "text/event-stream")
		httpReq.Header.Set("Cache-Control", "no-cache")

		// Send request
		resp, err := ac.httpClient.Do(httpReq)
		if err != nil {
			errorChan <- fmt.Errorf("error sending request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errorChan <- fmt.Errorf("agent service returned status %d: %s", resp.StatusCode, string(body))
			return
		}

		// Read SSE stream
		ac.readSSEStream(ctx, resp.Body, responseChan, errorChan)
	}()

	return responseChan, errorChan
}

// readSSEStream reads Server-Sent Events from the response body
func (ac *AiAgentClient) readSSEStream(ctx context.Context, body io.Reader, responseChan chan<- AgentResponse, errorChan chan<- error) {
	reader := bufio.NewReader(body)
	var dataLines []string

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Read a line (including the trailing '\n')
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Process any remaining line and buffered data
				line = strings.TrimRight(line, "\r\n")
				if line != "" && hasPrefix(line, "data: ") {
					dataLines = append(dataLines, line[6:])
				}
				if len(dataLines) > 0 {
					ac.processSSEData(strings.Join(dataLines, "\n"), responseChan)
				}
				return
			}
			log.Printf("Error reading stream: %v", err)
			errorChan <- fmt.Errorf("error reading stream: %w", err)
			return
		}

		// Trim trailing newline characters
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			// blank line indicates end of an event, process accumulated data lines
			if len(dataLines) > 0 {
				ac.processSSEData(strings.Join(dataLines, "\n"), responseChan)
				dataLines = dataLines[:0]
			}
			continue
		}

		if hasPrefix(trimmed, "data: ") {
			dataLines = append(dataLines, trimmed[6:])
		}
		// ignore other SSE fields (event:, id:, retry:, etc.)
	}
}

// processSSEData unmarshals JSON data from an SSE 'data' payload and forwards it
func (ac *AiAgentClient) processSSEData(jsonData string, responseChan chan<- AgentResponse) {
	if jsonData == "" {
		return
	}
	var response AgentResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		log.Printf("Error parsing JSON data: %v, data: %s", err, jsonData)
		return
	}

	// Best-effort send; block if receiver isn't ready to avoid dropping responses silently
	responseChan <- response
}

// Helper functions for string processing
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// HealthCheck checks if the agent service is healthy
func (ac *AiAgentClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", ac.baseURL+"/chat/health", nil)
	if err != nil {
		return fmt.Errorf("error creating health check request: %w", err)
	}

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending health check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent service health check failed with status %d", resp.StatusCode)
	}

	return nil
}
