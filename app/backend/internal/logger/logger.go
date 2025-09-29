package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

type contextKey string

const (
	loggerKey        contextKey = "logger"
	transactionIDKey contextKey = "txid"
	appNameKey       contextKey = "appname"
)

var (
	globalLogger zerolog.Logger
	appName      string = "tms-backend"
	counter      int64
	setupOnce    sync.Once
)

// Setup initializes the global logger - called automatically
func Setup() error {
	var err error
	setupOnce.Do(func() {
		err = setupLogger()
	})
	return err
}

func setupLogger() error {
	// Simple console output
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	// Set to info level by default
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Create global logger
	globalLogger = zerolog.New(output).
		With().
		Timestamp().
		Str("app", appName).
		Logger()

	return nil
}

// generateTransactionID creates a unique transaction ID
func generateTransactionID() string {
	pid := os.Getpid()
	timestamp := time.Now().Unix()
	count := atomic.AddInt64(&counter, 1)

	return fmt.Sprintf("tx_%d_%d_%d", timestamp, count, pid)
}

// WithTransaction creates a new context with a transaction ID and logger
func WithTransaction(ctx context.Context) context.Context {
	txID := generateTransactionID()
	logger := globalLogger.With().Str("tx_id", txID).Logger()

	ctx = context.WithValue(ctx, transactionIDKey, txID)
	ctx = context.WithValue(ctx, loggerKey, logger)
	ctx = context.WithValue(ctx, appNameKey, appName)

	return ctx
}

// getLogger retrieves the logger from context or returns global logger with caller info
func getLogger(ctx context.Context) *zerolog.Logger {
	// Ensure setup is called
	Setup()

	if ctx != nil {
		if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
			return &logger
		}
	}

	// Add caller info for global logger
	_, file, line, _ := runtime.Caller(2)
	filename := file[strings.LastIndex(file, "/")+1:]

	logger := globalLogger.With().
		Str("file", fmt.Sprintf("%s:%d", filename, line)).
		Logger()

	return &logger
}

// Standard logging functions - simple and familiar!

// Traditional log levels
func Info(msg string) {
	getLogger(nil).Info().Msg(msg)
}

func Infof(format string, args ...interface{}) {
	getLogger(nil).Info().Msgf(format, args...)
}

func Debug(msg string) {
	getLogger(nil).Debug().Msg(msg)
}

func Debugf(format string, args ...interface{}) {
	getLogger(nil).Debug().Msgf(format, args...)
}

func Error(err error, msg string) {
	getLogger(nil).Error().Err(err).Msg(msg)
}

func Errorf(format string, args ...interface{}) {
	getLogger(nil).Error().Msgf(format, args...)
}

func Warn(msg string) {
	getLogger(nil).Warn().Msg(msg)
}

func Warnf(format string, args ...interface{}) {
	getLogger(nil).Warn().Msgf(format, args...)
}

// Context-aware versions - automatically include transaction ID if available
func InfoCtx(ctx context.Context, msg string) {
	getLogger(ctx).Info().Msg(msg)
}

func InfofCtx(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Info().Msgf(format, args...)
}

func DebugCtx(ctx context.Context, msg string) {
	getLogger(ctx).Debug().Msg(msg)
}

func DebugfCtx(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Debug().Msgf(format, args...)
}

func ErrorCtx(ctx context.Context, err error, msg string) {
	getLogger(ctx).Error().Err(err).Msg(msg)
}

func ErrorfCtx(ctx context.Context, err error, format string, args ...interface{}) {
	getLogger(ctx).Error().Err(err).Msgf(format, args...)
}

func WarnCtx(ctx context.Context, msg string) {
	getLogger(ctx).Warn().Msg(msg)
}

func WarnfCtx(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Warn().Msgf(format, args...)
}

// Utility functions
func GetTransactionID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if txID, ok := ctx.Value(transactionIDKey).(string); ok {
		return txID
	}

	return ""
}

// TransactionalLogger provides backward compatibility
type TransactionalLogger struct {
	logger zerolog.Logger
}

// GetLoggerInstance returns a logger instance (backward compatibility)
func GetLoggerInstance() *TransactionalLogger {
	Setup()
	return &TransactionalLogger{logger: globalLogger}
}

// GetTxLogger returns a transactional logger from context (backward compatibility)
func GetTxLogger(ctx context.Context) *TransactionalLogger {
	Setup()

	if ctx != nil {
		if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
			return &TransactionalLogger{logger: logger}
		}
	}

	// Add caller info for global logger
	_, file, line, _ := runtime.Caller(1)
	filename := file[strings.LastIndex(file, "/")+1:]

	callerLogger := globalLogger.With().
		Str("file", fmt.Sprintf("%s:%d", filename, line)).
		Logger()

	return &TransactionalLogger{logger: callerLogger}
}

// WithTxLogger creates a context with a transaction logger (backward compatibility)
func WithTxLogger(ctx context.Context, txID string, logger zerolog.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx = context.WithValue(ctx, transactionIDKey, txID)
	ctx = context.WithValue(ctx, loggerKey, logger)

	return ctx
}

// TransactionalLogger methods
func (tl *TransactionalLogger) WithComponent(component string) *ComponentLogger {
	return &ComponentLogger{
		logger:    tl,
		component: component,
	}
}

func (tl *TransactionalLogger) WithTransaction(ctx context.Context) context.Context {
	return WithTransaction(ctx)
}

func (tl *TransactionalLogger) With() zerolog.Context {
	return tl.logger.With()
}

func (tl *TransactionalLogger) Info() *zerolog.Event {
	return tl.logger.Info()
}

func (tl *TransactionalLogger) Debug() *zerolog.Event {
	return tl.logger.Debug()
}

func (tl *TransactionalLogger) Warn() *zerolog.Event {
	return tl.logger.Warn()
}

func (tl *TransactionalLogger) Error() *zerolog.Event {
	return tl.logger.Error()
}

func (tl *TransactionalLogger) Fatal() *zerolog.Event {
	return tl.logger.Fatal()
}

// ComponentLogger provides component-scoped logging
type ComponentLogger struct {
	logger    *TransactionalLogger
	component string
}

func (cl *ComponentLogger) With() zerolog.Context {
	return cl.logger.logger.With().Str("component", cl.component)
}

func (cl *ComponentLogger) Info() *zerolog.Event {
	return cl.logger.logger.Info().Str("component", cl.component)
}

func (cl *ComponentLogger) Debug() *zerolog.Event {
	return cl.logger.logger.Debug().Str("component", cl.component)
}

func (cl *ComponentLogger) Warn() *zerolog.Event {
	return cl.logger.logger.Warn().Str("component", cl.component)
}

func (cl *ComponentLogger) Error() *zerolog.Event {
	return cl.logger.logger.Error().Str("component", cl.component)
}

func (cl *ComponentLogger) Fatal() *zerolog.Event {
	return cl.logger.logger.Fatal().Str("component", cl.component)
}
