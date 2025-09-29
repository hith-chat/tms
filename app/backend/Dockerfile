# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o tms-backend ./cmd/api/main.go

# Final stage - use official Playwright image which has browsers pre-installed
FROM mcr.microsoft.com/playwright:v1.52.0-jammy

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/tms-backend .

# Copy templates directory
COPY --from=builder /app/templates ./templates


# Expose port
EXPOSE 8080

# Health check
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD wget --no-verbose --tries=1 --spider http://localhost:8080/v1/public/health || exit 1

# Run the binary
CMD ["./tms-backend"]
