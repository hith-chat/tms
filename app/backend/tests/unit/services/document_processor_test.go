package services

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/bareuptime/tms/tests/fixtures"
	"github.com/stretchr/testify/assert"
)

// TestFileValidation tests file validation logic
func TestFileValidation(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		size        int64
		contentType string
		shouldFail  bool
		errorCheck  string
	}{
		{
			name:        "Valid PDF file",
			filename:    "test.pdf",
			size:        1024,
			contentType: "application/pdf",
			shouldFail:  false,
		},
		{
			name:        "File too large",
			filename:    "large.pdf",
			size:        15 * 1024 * 1024, // 15MB
			contentType: "application/pdf",
			shouldFail:  true,
			errorCheck:  "exceeds maximum",
		},
		{
			name:        "Invalid file extension",
			filename:    "test.txt",
			size:        1024,
			contentType: "text/plain",
			shouldFail:  true,
			errorCheck:  "unsupported file type",
		},
		{
			name:        "Invalid content type",
			filename:    "test.pdf",
			size:        1024,
			contentType: "text/plain",
			shouldFail:  true,
			errorCheck:  "invalid content type",
		},
		{
			name:        "Empty filename",
			filename:    "",
			size:        1024,
			contentType: "application/pdf",
			shouldFail:  true,
		},
		{
			name:        "Zero file size",
			filename:    "test.pdf",
			size:        0,
			contentType: "application/pdf",
			shouldFail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic directly
			err := validateTestFile(tt.filename, tt.size, tt.contentType, 10*1024*1024)

			if tt.shouldFail {
				assert.Error(t, err, "Expected validation to fail for %s", tt.name)
				if tt.errorCheck != "" {
					assert.Contains(t, err.Error(), tt.errorCheck, "Error should contain expected message")
				}
			} else {
				assert.NoError(t, err, "Expected validation to pass for %s", tt.name)
			}
		})
	}
}

// validateTestFile simulates the file validation logic for testing
func validateTestFile(filename string, size int64, contentType string, maxSize int64) error {
	// Check file size
	if size > maxSize {
		return errors.New("file size exceeds maximum allowed size")
	}

	if size <= 0 {
		return errors.New("file size must be greater than 0")
	}

	// Check filename
	if filename == "" {
		return errors.New("filename cannot be empty")
	}

	// Check file extension
	if len(filename) < 5 || filename[len(filename)-4:] != ".pdf" {
		return errors.New("unsupported file type. Only PDF files are supported")
	}

	// Validate content type
	if contentType != "" && contentType != "application/pdf" {
		return errors.New("invalid content type. Expected PDF")
	}

	return nil
}

// TestCreateChunks tests the text chunking functionality
func TestCreateChunks(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		chunkSize   int
		overlap     int
		expectedMin int
		expectedMax int
	}{
		{
			name:        "Short text single chunk",
			text:        "This is a short text.",
			chunkSize:   100,
			overlap:     10,
			expectedMin: 1,
			expectedMax: 1,
		},
		{
			name:        "Long text multiple chunks",
			text:        fixtures.TestPDFContent,
			chunkSize:   200,
			overlap:     20,
			expectedMin: 2,
			expectedMax: 10,
		},
		{
			name:        "Empty text",
			text:        "",
			chunkSize:   100,
			overlap:     10,
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name:        "Text with exact chunk size",
			text:        "This text is exactly 50 characters long for test.",
			chunkSize:   50,
			overlap:     0,
			expectedMin: 1,
			expectedMax: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := createTestChunks(tt.text, tt.chunkSize, tt.overlap)

			assert.GreaterOrEqual(t, len(chunks), tt.expectedMin, "Should have at least %d chunks", tt.expectedMin)
			assert.LessOrEqual(t, len(chunks), tt.expectedMax, "Should have at most %d chunks", tt.expectedMax)

			// Verify chunk properties
			for i, chunk := range chunks {
				assert.LessOrEqual(t, len(chunk), tt.chunkSize+tt.overlap, "Chunk should not exceed size limit")
				if len(chunk) > 0 {
					assert.NotEmpty(t, chunk, "Chunk should not be empty")
				}

				// Verify overlap for consecutive chunks (if applicable)
				if i > 0 && tt.overlap > 0 && len(chunks[i-1]) > tt.overlap {
					// Check that there might be some overlap
					assert.NotEqual(t, chunks[i-1], chunk, "Consecutive chunks should be different")
				}
			}
		})
	}
}

// Helper function to simulate chunking for testing
func createTestChunks(text string, chunkSize, overlap int) []string {
	if len(text) == 0 || chunkSize <= 0 {
		return []string{}
	}

	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(text) {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}

		chunk := text[start:end]
		chunks = append(chunks, chunk)

		if end == len(text) {
			break
		}

		// Move start position with overlap consideration
		nextStart := end - overlap
		if nextStart <= start {
			// Prevent infinite loop if overlap is too large
			nextStart = start + 1
		}
		start = nextStart
	}

	return chunks
}

// TestFileOperations tests file saving and cleanup
func TestFileOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Test file saving
	content := "Test PDF content for file operations"
	file := &mockFile{reader: bytes.NewReader([]byte(content))}

	// Test successful file save
	savedPath, err := saveTestFile(file, "test-doc-id", "test.pdf", tempDir)
	assert.NoError(t, err, "File saving should succeed")
	assert.FileExists(t, savedPath, "File should be saved to disk")

	// Verify file content
	savedContent, err := os.ReadFile(savedPath)
	assert.NoError(t, err, "Should be able to read saved file")
	assert.Equal(t, content, string(savedContent), "Saved content should match original")

	// Test cleanup
	err = os.Remove(savedPath)
	assert.NoError(t, err, "Should be able to clean up file")
	assert.NoFileExists(t, savedPath, "File should be removed after cleanup")
}

// saveTestFile simulates file saving logic for testing
func saveTestFile(file io.Reader, docID, filename, uploadDir string) (string, error) {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Generate safe filename
	ext := ".pdf"
	if len(filename) > 4 {
		ext = filename[len(filename)-4:]
	}
	safeFilename := docID + ext
	filePath := uploadDir + "/" + safeFilename

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		os.Remove(filePath)
		return "", err
	}

	return filePath, nil
}

// TestTokenCounting tests token counting logic
func TestTokenCounting(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectedMin int
		expectedMax int
	}{
		{
			name:        "Simple sentence",
			text:        "This is a simple test sentence.",
			expectedMin: 5,
			expectedMax: 8,
		},
		{
			name:        "Empty string",
			text:        "",
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name:        "Single word",
			text:        "word",
			expectedMin: 1,
			expectedMax: 1,
		},
		{
			name:        "Multiple spaces",
			text:        "word1    word2     word3",
			expectedMin: 3,
			expectedMax: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenCount := countTestTokens(tt.text)

			assert.GreaterOrEqual(t, tokenCount, tt.expectedMin, "Token count should be at least %d", tt.expectedMin)
			assert.LessOrEqual(t, tokenCount, tt.expectedMax, "Token count should be at most %d", tt.expectedMax)
		})
	}
}

// countTestTokens simulates token counting for testing
func countTestTokens(text string) int {
	if text == "" {
		return 0
	}

	// Simple word-based token counting for testing
	words := 0
	inWord := false

	for _, char := range text {
		if char == ' ' || char == '\t' || char == '\n' {
			inWord = false
		} else {
			if !inWord {
				words++
				inWord = true
			}
		}
	}

	return words
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("Invalid directory", func(t *testing.T) {
		// Try to save to an invalid directory
		file := &mockFile{reader: bytes.NewReader([]byte("content"))}
		_, err := saveTestFile(file, "test", "test.pdf", "/invalid/directory/path")
		assert.Error(t, err, "Should fail when directory cannot be created")
	})

	t.Run("Invalid file reader", func(t *testing.T) {
		tempDir := t.TempDir()
		file := &errorReader{}
		_, err := saveTestFile(file, "test", "test.pdf", tempDir)
		assert.Error(t, err, "Should fail when file cannot be read")
	})
}

// mockFile implements io.Reader interface for testing
type mockFile struct {
	reader io.Reader
}

func (m *mockFile) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

// errorReader always returns an error when reading
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

// TestChunkOverlap tests overlap functionality specifically
func TestChunkOverlap(t *testing.T) {
	text := "This is a long text that should be split into multiple chunks with proper overlap between them for testing purposes."

	chunks := createTestChunks(text, 30, 10)

	// Verify multiple chunks are created
	assert.Greater(t, len(chunks), 1, "Should create multiple chunks")

	// Verify each chunk respects size limits
	for i, chunk := range chunks {
		assert.LessOrEqual(t, len(chunk), 30+10, "Chunk %d should not exceed size limit", i)
		assert.NotEmpty(t, chunk, "Chunk %d should not be empty", i)
	}

	// Verify consecutive chunks are different (basic overlap test)
	for i := 1; i < len(chunks); i++ {
		assert.NotEqual(t, chunks[i-1], chunks[i], "Consecutive chunks should be different")
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("Single character text", func(t *testing.T) {
		chunks := createTestChunks("a", 100, 10)
		assert.Len(t, chunks, 1, "Single character should produce one chunk")
		assert.Equal(t, "a", chunks[0], "Single chunk should contain the character")
	})

	t.Run("Whitespace only text", func(t *testing.T) {
		chunks := createTestChunks("   \n\t  ", 100, 10)
		// This depends on implementation - might be empty or contain whitespace
		for _, chunk := range chunks {
			assert.NotEmpty(t, chunk, "If chunks are created, they should not be empty")
		}
	})

	t.Run("Zero chunk size", func(t *testing.T) {
		// This should be handled gracefully
		chunks := createTestChunks("test text", 0, 0)
		assert.Empty(t, chunks, "Zero chunk size should produce no chunks")
	})

	t.Run("Overlap larger than chunk size", func(t *testing.T) {
		chunks := createTestChunks("test text that is longer", 5, 10)
		// Should handle this edge case gracefully
		assert.NotEmpty(t, chunks, "Should still produce chunks")
	})
}
