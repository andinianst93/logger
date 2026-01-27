package logger

import (
	"sync"

	"go.uber.org/zap"
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
