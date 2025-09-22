package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/flashmob/go-guerrilla"
	"github.com/flashmob/go-guerrilla/backends"
	guerrillamail "github.com/flashmob/go-guerrilla/mail"
)

// Ticket data structure
type TicketData struct {
	Tenant    string    `json:"tenant"`
	From      string    `json:"from"`
	Subject   string    `json:"subject"`
	Content   string    `json:"content"`
	MessageID string    `json:"message_id"`
	InReplyTo string    `json:"in_reply_to"`
	CreatedAt time.Time `json:"created_at"`
}

// Custom backend processor that implements Processor interface
type TicketProcessor struct{}

func (p *TicketProcessor) Process(e *guerrillamail.Envelope, task backends.SelectTask) (backends.Result, error) {
	var rcptStr string
	if len(e.RcptTo) > 0 {
		rcptStr = e.RcptTo[0].String()
	}
	log.Printf("Processing email from %s to %s", e.MailFrom.String(), rcptStr)

	// Parse email message
	msg, err := mail.ReadMessage(strings.NewReader(e.Data.String()))
	if err != nil {
		log.Printf("Error parsing message: %v", err)
		return backends.NewResult("550 Error parsing message"), err
	}

	// Extract ticket data
	ticket := TicketData{
		Tenant:    extractTenant(rcptStr),
		From:      e.MailFrom.String(),
		Subject:   msg.Header.Get("Subject"),
		Content:   extractContent(msg),
		MessageID: msg.Header.Get("Message-ID"),
		InReplyTo: msg.Header.Get("In-Reply-To"),
		CreatedAt: time.Now(),
	}

	// Send to your ticket API
	if err := sendToTicketAPI(ticket); err != nil {
		log.Printf("Failed to create a ticket: %v", err)
		return backends.NewResult("451 Temporary failure - please retry"), err
	}

	log.Printf("Ticket created successfully for tenant: %s", ticket.Tenant)
	return backends.NewResult("250 Message accepted for delivery"), nil
}

// Extract clean content from email
func extractContent(msg *mail.Message) string {
	body, err := io.ReadAll(msg.Body)
	if err != nil {
		return ""
	}

	content := string(body)

	// Handle multipart messages (get text/plain part)
	if strings.Contains(msg.Header.Get("Content-Type"), "multipart") {
		content = extractTextFromMultipart(content)
	}

	// Clean up content
	content = cleanEmailContent(content)

	// Limit content size
	if len(content) > 5000 {
		content = content[:5000] + "\n\n[Content truncated]"
	}

	return content
}

func extractTextFromMultipart(content string) string {
	// Look for text/plain part in multipart message
	lines := strings.Split(content, "\n")
	var textContent []string
	inTextPart := false

	for _, line := range lines {
		if strings.Contains(line, "Content-Type: text/plain") {
			inTextPart = true
			continue
		}
		if strings.HasPrefix(line, "--") && inTextPart {
			break
		}
		if inTextPart && !strings.HasPrefix(line, "Content-") {
			textContent = append(textContent, line)
		}
	}

	if len(textContent) > 0 {
		return strings.Join(textContent, "\n")
	}

	return content
}

func cleanEmailContent(content string) string {
	lines := strings.Split(content, "\n")
	var cleanLines []string

	// Regex patterns for common email artifacts
	quotedReplyPattern := regexp.MustCompile(`^>.*`)
	forwardPattern := regexp.MustCompile(`(?i)^(from:|sent:|to:|subject:)`)
	signaturePattern := regexp.MustCompile(`^--\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines at the start
		if len(cleanLines) == 0 && line == "" {
			continue
		}

		// Stop at quoted replies
		if quotedReplyPattern.MatchString(line) {
			break
		}

		// Stop at forwarded message headers
		if forwardPattern.MatchString(line) {
			break
		}

		// Stop at email signatures
		if signaturePattern.MatchString(line) {
			break
		}

		// Stop at common reply indicators
		if strings.Contains(strings.ToLower(line), "wrote:") ||
			strings.Contains(strings.ToLower(line), "on ") && strings.Contains(strings.ToLower(line), "at ") {
			break
		}

		cleanLines = append(cleanLines, line)
	}

	return strings.TrimSpace(strings.Join(cleanLines, "\n"))
}

func extractTenant(email string) string {
	// Extract tenant from email like tenant-penify@yourmailserver.com
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		localPart := parts[0]
		if strings.HasPrefix(localPart, "tenant-") {
			return strings.TrimPrefix(localPart, "tenant-")
		}
		// Fallback: use the whole local part
		return localPart
	}
	return "unknown"
}

func sendToTicketAPI(ticket TicketData) error {
	apiURL := os.Getenv("TICKET_API_URL")
	if apiURL == "" {
		apiURL = "http://backend:8080/v1/public/email-to-ticket"
	}

	jsonData, err := json.Marshal(ticket)
	if err != nil {
		return fmt.Errorf("error marshaling ticket data: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending to API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("Ticket created successfully - API responded with %d", resp.StatusCode)
	return nil
}

func main() {
	// Configuration
	domain := os.Getenv("MAIL_DOMAIN")
	if domain == "" {
		domain = "hith.chat"
	}

	listenInterface := os.Getenv("LISTEN_INTERFACE")
	if listenInterface == "" {
		listenInterface = "0.0.0.0:25"
	}

	maxSize := int64(1024000) // 1MB default
	if sizeEnv := os.Getenv("MAX_MESSAGE_SIZE"); sizeEnv != "" {
		if size, err := fmt.Sscanf(sizeEnv, "%d", &maxSize); err == nil && size == 1 {
			log.Printf("Using max message size: %d bytes", maxSize)
		}
	}

	// Guerrilla configuration
	cfg := &guerrilla.AppConfig{
		LogFile:      "stdout",
		PidFile:      "/tmp/guerrilla.pid",
		AllowedHosts: []string{domain},
	}

	// Server configuration
	sc := guerrilla.ServerConfig{
		ListenInterface: listenInterface,
		IsEnabled:       true,
		MaxSize:         maxSize,
		Timeout:         180, // 3 minutes timeout for production
		Hostname:        domain,
		MaxClients:      500, // Higher limit for production
		TLS: guerrilla.ServerTLSConfig{
			StartTLSOn: false, // Set to true if you have SSL certs
		},
	}
	cfg.Servers = append(cfg.Servers, sc)

	// Backend configuration
	cfg.BackendConfig = backends.BackendConfig{
		"save_workers_size": 8,
		"save_process":      "HeadersParser|Header|Hasher|TicketProcessor",
	}

	log.Printf("Starting Guerrilla Mail Server on %s for domain %s", listenInterface, domain)
	log.Printf("Ticket API URL: %s", os.Getenv("TICKET_API_URL"))
	log.Printf("Max message size: %d bytes", maxSize)

	// Register our custom processor
	processor := &TicketProcessor{}
	backends.Svc.AddProcessor("TicketProcessor", func() backends.Decorator {
		return func(p backends.Processor) backends.Processor {
			return processor
		}
	})

	// Create and start the daemon
	daemon := guerrilla.Daemon{Config: cfg}
	err := daemon.Start()
	if err != nil {
		log.Fatalf("Failed to start mail server: %v", err)
	}

	// Keep the server running
	select {}
}
