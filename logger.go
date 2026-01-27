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
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Global logger instance
	globalLogger *zap.Logger
	globalMu     sync.RWMutex
)

func init() {
	// Initialize with default logger
	l, _ := New(Config{
		Level:      "info",
		Format:     "json",
		OutputPath: "stdout",
	})
	globalLogger = l
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *zap.Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *zap.Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalLogger
}

// L returns the global logger (shorthand for GetGlobalLogger)
func L() *zap.Logger {
	return GetGlobalLogger()
}

// Debug logs a debug message using global logger
func Debug(msg string, fields ...zap.Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

// Info logs an info message using global logger
func Info(msg string, fields ...zap.Field) {
	GetGlobalLogger().Info(msg, fields...)
}

// Warn logs a warning message using global logger
func Warn(msg string, fields ...zap.Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

// Error logs an error message using global logger
func Error(msg string, fields ...zap.Field) {
	GetGlobalLogger().Error(msg, fields...)
}

// Fatal logs a fatal message using global logger and exits
func Fatal(msg string, fields ...zap.Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}

// Sync flushes any buffered log entries (call before exit)
func Sync() error {
	return GetGlobalLogger().Sync()
}

// RotationConfig configures log file rotation
type RotationConfig struct {
	Filename   string // Full path to log file
	MaxSize    int    // Maximum size in megabytes before rotation (default: 100MB)
	MaxBackups int    // Maximum number of old log files to keep (default: 3)
	MaxAge     int    // Maximum days to keep old log files (default: 30)
	Compress   bool   // Compress old log files (default: true)
}

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

	// Advanced options
	Rotation       *RotationConfig // Enable log rotation (nil = disabled)
	AdditionalOuts []string        // Additional output paths (e.g., ["stdout", "/var/log/app.log"])
}

func New(cfg Config) (*zap.Logger, error) {
	// 1. Parse log level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// 2. Setup encoder (format output)
	encoder := getEncoder(cfg.Format)

	// 3. Setup output writer(s)
	var writers []zapcore.WriteSyncer

	// Primary output
	if cfg.OutputPath != "" {
		writer, err := getWriter(cfg.OutputPath, cfg.Rotation)
		if err != nil {
			return nil, err
		}
		writers = append(writers, writer)
	}

	// Additional outputs
	if len(cfg.AdditionalOuts) > 0 {
		for _, path := range cfg.AdditionalOuts {
			writer, err := getWriter(path, nil) // rotation only for primary
			if err != nil {
				return nil, fmt.Errorf("failed to setup additional output %s: %w", path, err)
			}
			writers = append(writers, writer)
		}
	}

	// Combine all writers
	var finalWriter zapcore.WriteSyncer
	if len(writers) == 1 {
		finalWriter = writers[0]
	} else {
		finalWriter = zapcore.NewMultiWriteSyncer(writers...)
	}

	// 4. Build core logger
	core := zapcore.NewCore(
		encoder,     // How to encode (JSON/Console)
		finalWriter, // Where to write (stdout/file)
		level,       // Minimum level to log
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

func getWriter(outputPath string, rotation *RotationConfig) (zapcore.WriteSyncer, error) {
	switch outputPath {
	case "stdout", "":
		return zapcore.AddSync(os.Stdout), nil
	case "stderr":
		return zapcore.AddSync(os.Stderr), nil
	default:
		// File output with optional rotation
		if rotation != nil {
			// Use lumberjack for log rotation
			return zapcore.AddSync(&lumberjack.Logger{
				Filename:   rotation.Filename,
				MaxSize:    getOrDefault(rotation.MaxSize, 100),  // 100MB default
				MaxBackups: getOrDefault(rotation.MaxBackups, 3), // keep 3 backups
				MaxAge:     getOrDefault(rotation.MaxAge, 30),    // 30 days
				Compress:   rotation.Compress,                    // compress rotated files
			}), nil
		}

		// Standard file without rotation
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

func getOrDefault(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
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

// LogError logs an error and returns it (useful for error handling)
func LogError(logger *zap.Logger, msg string, err error, fields ...zap.Field) error {
	if err == nil {
		return nil
	}
	allFields := append(fields, zap.Error(err))
	logger.Error(msg, allFields...)
	return err
}

// LogErrorAndExit logs an error and exits with status code 1
func LogErrorAndExit(logger *zap.Logger, msg string, err error, fields ...zap.Field) {
	if err != nil {
		allFields := append(fields, zap.Error(err))
		logger.Fatal(msg, allFields...)
	}
}

// Must panics if error is not nil (useful for initialization)
func Must(logger *zap.Logger, err error) {
	if err != nil {
		logger.Fatal("Fatal error", zap.Error(err))
	}
}

// MustLogger creates a logger or panics (useful for init)
func MustLogger(cfg Config) *zap.Logger {
	logger, err := New(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
}
