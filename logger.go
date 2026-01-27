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
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level      string // "debug", "info", "warn", "error", "fatal"
	Format     string // "json" or "console"
	OutputPath string // "stdout", "stderr", or file path
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

	return logger, nil
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
