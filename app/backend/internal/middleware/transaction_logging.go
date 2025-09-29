package middleware

import (
	"time"

	"github.com/bareuptime/tms/internal/logger"
	"github.com/gin-gonic/gin"
)

// TransactionLoggingMiddleware creates a middleware that adds transactional logging to each request
func TransactionLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Create transaction context
		ctx := logger.WithTransaction(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		// Log request start
		logger.InfofCtx(ctx, "HTTP request started - %s %s from %s",
			c.Request.Method, c.Request.URL.Path, c.ClientIP())

		// Process request
		c.Next()

		// Log request completion
		duration := time.Since(start)
		logger.InfofCtx(ctx, "HTTP request completed - %s %s %d (%v)",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}

// GetTxLoggerFromGin retrieves the transaction logger from gin context
func GetTxLoggerFromGin(c *gin.Context) *logger.TransactionalLogger {
	return logger.GetTxLogger(c.Request.Context())
}
