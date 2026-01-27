package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// getWriter creates a WriteSyncer based on output path
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

// getEncoder creates encoder based on format
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

// parseLevel converts string level to zapcore.Level
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

// getOrDefault returns value or defaultValue if value is 0
func getOrDefault(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

// getEnvOrDefault reads environment variable with default fallback
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewFromEnv creates a logger from environment variables
func NewFromEnv() (*zap.Logger, error) {
	cfg := Config{
		Level:      getEnvOrDefault("LOG_LEVEL", "info"),
		Format:     getEnvOrDefault("LOG_FORMAT", "json"),
		OutputPath: getEnvOrDefault("LOG_OUTPUT", "stdout"),
	}

	return New(cfg)
}
