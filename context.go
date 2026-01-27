package logger

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// NewTraceID generates a new random trace ID
// Format: 16 bytes hex string (32 characters)
func NewTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%d", os.Getpid())
	}
	return hex.EncodeToString(b)
}

// WithTraceID adds trace_id to logger for request tracing
func WithTraceID(logger *zap.Logger, traceID string) *zap.Logger {
	if traceID == "" {
		traceID = NewTraceID()
	}
	return logger.With(zap.String("trace_id", traceID))
}

// WithRequestID adds request_id to logger (alias for WithTraceID)
func WithRequestID(logger *zap.Logger, requestID string) *zap.Logger {
	if requestID == "" {
		requestID = NewTraceID()
	}
	return logger.With(zap.String("request_id", requestID))
}

// WithFields adds custom fields to logger
func WithFields(logger *zap.Logger, fields ...zap.Field) *zap.Logger {
	return logger.With(fields...)
}
