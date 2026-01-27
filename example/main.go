package main

import (
	"github.com/andinianst93/logger"
	"go.uber.org/zap"
)

func main() {
	// Create production logger with metadata
	log, err := logger.New(logger.Config{
		Level:       "info",
		Format:      "json",
		OutputPath:  "stdout",
		ServiceName: "user-service",
		Environment: "production",
		Version:     "v1.0.0",
		// Host and PID will be auto-detected
	})
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	// Basic logging
	log.Info("Application started")

	// Request tracing example
	traceID := logger.NewTraceID()
	reqLog := logger.WithTraceID(log, traceID)

	reqLog.Info("Processing user request",
		zap.String("user_id", "12345"),
		zap.String("action", "login"),
	)

	// Add more context
	userLog := logger.WithFields(reqLog,
		zap.String("user_id", "12345"),
		zap.String("tenant_id", "acme-corp"),
	)

	userLog.Info("User action completed", zap.Int("status", 200))
}
