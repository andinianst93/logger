/*
Debug   // Development, technical details, verbose logging
↓
Info    // Normal operations (startup, config loaded)
↓
Warn    // Something unusual but not an error (deprecated API)
↓
Error   // Error that can be recovered (retry, fallback)
↓
Fatal   // Fatal error, program must exit
*/
package logger

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level      string // "debug", "info", "warn", "error", "fatal"
	Format     string // "json" or "console"
	OutputPath string // "stdout", "stderr", or file path

	// Metadata fields that will be added to every log entry
	ServiceName string // Name of the service/application
	Environment string // "development", "staging", "production"
	Version     string // Application version (e.g., "v1.0.0", git commit hash)
	Host        string // Hostname or instance identifier
	PID         int    // Process ID (0 = auto-detect)
}

func New(cfg Config) (*zap.Logger, error) {
	// 1. Parse log level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// 2. Setup encoder (format output)
	encoder := getEncoder(cfg.Format)

	// 3. Setup output writer
	writer, err := getWriter(cfg.OutputPath)
	if err != nil {
		return nil, err
	}

	// 4. Build core logger
	core := zapcore.NewCore(
		encoder, // How to encode (JSON/Console)
		writer,  // Where to write (stdout/file)
		level,   // Minimum level to log
	)

	// 5. Create logger with options
	logger := zap.New(core,
		zap.AddCaller(),                       // Add file:line info
		zap.AddStacktrace(zapcore.ErrorLevel), // Stacktrace untuk errors
	)

	// 6. Add metadata fields as persistent fields
	logger = addMetadataFields(logger, cfg)

	return logger, nil
}

// addMetadataFields adds persistent metadata to all log entries
func addMetadataFields(logger *zap.Logger, cfg Config) *zap.Logger {
	fields := make([]zap.Field, 0)

	if cfg.ServiceName != "" {
		fields = append(fields, zap.String("service", cfg.ServiceName))
	}

	if cfg.Environment != "" {
		fields = append(fields, zap.String("environment", cfg.Environment))
	}

	if cfg.Version != "" {
		fields = append(fields, zap.String("version", cfg.Version))
	}

	// Auto-detect hostname if not provided
	host := cfg.Host
	if host == "" {
		if h, err := os.Hostname(); err == nil {
			host = h
		}
	}
	if host != "" {
		fields = append(fields, zap.String("host", host))
	}

	// Auto-detect PID if not provided
	pid := cfg.PID
	if pid == 0 {
		pid = os.Getpid()
	}
	fields = append(fields, zap.Int("pid", pid))

	if len(fields) > 0 {
		logger = logger.With(fields...)
	}

	return logger
}

func parseLevel(levelStr string) (zapcore.Level, error) {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("invalid log level: %s", levelStr)
	}
}

func getEncoder(format string) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()

	// Customize time format
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Capitalize level names (INFO, DEBUG)
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if format == "console" {
		// Human-readable format untuk development
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Colored
		return zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Default: JSON format untuk production
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getWriter(outputPath string) (zapcore.WriteSyncer, error) {
	switch outputPath {
	case "stdout", "":
		return zapcore.AddSync(os.Stdout), nil
	case "stderr":
		return zapcore.AddSync(os.Stderr), nil
	default:
		// File output
		file, err := os.OpenFile(
			outputPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		return zapcore.AddSync(file), nil
	}
}

// Wrapper to load config dari ENV + create logger
func NewFromEnv() (*zap.Logger, error) {
	cfg := Config{
		Level:      getEnvOrDefault("LOG_LEVEL", "info"),
		Format:     getEnvOrDefault("LOG_FORMAT", "json"),
		OutputPath: getEnvOrDefault("LOG_OUTPUT", "stdout"),
	}

	return New(cfg)
}

// Helper to read ENV with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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
