package logger

import (
	"fmt"

	"go.uber.org/zap"
)

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
