package main

import (
	"errors"
	"fmt"

	"github.com/andinianst93/logger"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("=== Logger Examples ===\n")

	// Example 1: Global Logger (Simplest)
	example1GlobalLogger()

	// Example 2: Production Logger with Metadata
	example2Metadata()

	// Example 3: Request Tracing
	example3Tracing()

	// Example 4: Log Rotation
	example4Rotation()

	// Example 5: Multiple Outputs
	example5MultipleOutputs()

	// Example 6: Error Handling Helpers
	example6ErrorHandling()
}

// Example 1: Global Logger - Simplest approach
func example1GlobalLogger() {
	fmt.Println("--- Example 1: Global Logger ---")

	// Initialize global logger once at startup
	log := logger.MustLogger(logger.Config{
		Level:       "info",
		Format:      "console",
		OutputPath:  "stdout",
		ServiceName: "demo-service",
		Environment: "development",
	})
	logger.SetGlobalLogger(log)
	defer logger.Sync()

	// Now use logger anywhere without passing instance
	logger.Info("Application started with global logger")
	logger.Warn("This is a warning")

	// Or get the instance if needed
	log = logger.L()
	log.Info("Got logger instance", zap.String("method", "logger.L()"))

	fmt.Println()
}

// Example 2: Production Logger with Metadata
func example2Metadata() {
	fmt.Println("--- Example 2: Metadata Fields ---")

	log, _ := logger.New(logger.Config{
		Level:       "info",
		Format:      "json",
		OutputPath:  "stdout",
		ServiceName: "user-service",
		Environment: "production",
		Version:     "v1.2.3",
		// Host and PID auto-detected
	})
	defer log.Sync()

	log.Info("User logged in",
		zap.String("user_id", "12345"),
		zap.String("ip", "192.168.1.1"),
	)

	fmt.Println()
}

// Example 3: Request Tracing
func example3Tracing() {
	fmt.Println("--- Example 3: Request Tracing ---")

	log, _ := logger.New(logger.Config{
		Level:       "info",
		Format:      "console",
		OutputPath:  "stdout",
		ServiceName: "api-gateway",
	})
	defer log.Sync()

	// Generate trace ID for this request
	traceID := logger.NewTraceID()
	reqLog := logger.WithTraceID(log, traceID)

	// All logs in this request will have same trace_id
	reqLog.Info("Request started", zap.String("path", "/api/users"))
	reqLog.Info("Database query executed")
	reqLog.Info("Request completed", zap.Int("status", 200))

	// Add more context with WithFields
	userLog := logger.WithFields(reqLog,
		zap.String("user_id", "12345"),
		zap.String("tenant_id", "acme-corp"),
	)
	userLog.Info("User context added")

	fmt.Println()
}

// Example 4: Log Rotation (Production)
func example4Rotation() {
	fmt.Println("--- Example 4: Log Rotation ---")

	log, _ := logger.New(logger.Config{
		Level:       "info",
		Format:      "json",
		OutputPath:  "/tmp/app.log",
		ServiceName: "worker-service",
		Environment: "production",

		// Enable log rotation
		Rotation: &logger.RotationConfig{
			Filename:   "/tmp/app.log",
			MaxSize:    100,  // 100 MB per file
			MaxBackups: 7,    // Keep 7 old files
			MaxAge:     30,   // Keep for 30 days
			Compress:   true, // Compress old files
		},
	})
	defer log.Sync()

	log.Info("Log rotation enabled - files will auto-rotate at 100MB")

	fmt.Println("Log written to /tmp/app.log with rotation\n")
}

// Example 5: Multiple Outputs (Console + File)
func example5MultipleOutputs() {
	fmt.Println("--- Example 5: Multiple Outputs ---")

	log, _ := logger.New(logger.Config{
		Level:          "debug",
		Format:         "console",
		OutputPath:     "/tmp/multi.log",   // Primary output
		AdditionalOuts: []string{"stdout"}, // Also log to console
		ServiceName:    "multi-output-service",
	})
	defer log.Sync()

	log.Info("This appears in BOTH /tmp/multi.log AND console")
	log.Debug("Debug logs also go to both outputs")

	fmt.Println()
}

// Example 6: Error Handling Helpers
func example6ErrorHandling() {
	fmt.Println("--- Example 6: Error Handling ---")

	log, _ := logger.New(logger.Config{
		Level:      "info",
		Format:     "console",
		OutputPath: "stdout",
	})
	defer log.Sync()

	// LogError: logs and returns error
	if err := fetchData(log); err != nil {
		fmt.Printf("Got error: %v\n", err)
	}

	// Must: panic if error (useful for init)
	logger.Must(log, validateConfig())
	log.Info("Config validated successfully")

	// MustLogger: create logger or panic
	_ = logger.MustLogger(logger.Config{
		Level:  "info",
		Format: "json",
	})
	log.Info("Logger created with MustLogger")

	fmt.Println()
}

// Helper functions for examples
func fetchData(log *zap.Logger) error {
	err := errors.New("connection timeout")
	// Log error and return it
	return logger.LogError(log, "Failed to fetch data", err,
		zap.String("source", "database"),
		zap.Int("retry", 3),
	)
}

func validateConfig() error {
	// Return nil = success
	return nil
}
