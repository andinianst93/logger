# Logger Examples

This directory contains comprehensive examples demonstrating all features of the logger library.

## Running the Examples

```bash
cd example
go run main.go
```

## Examples Included

### 1. Global Logger
Shows the simplest way to use the logger with a global instance. No need to pass the logger around.

```go
logger.SetGlobalLogger(log)
logger.Info("Hello world")
```

### 2. Metadata Fields
Demonstrates persistent metadata fields that appear in every log entry (service name, environment, version, host, PID).

### 3. Request Tracing
Shows how to use trace IDs to track requests across your application.

```go
traceID := logger.NewTraceID()
reqLog := logger.WithTraceID(log, traceID)
reqLog.Info("Processing request")
```

### 4. Log Rotation
Production-ready log rotation based on file size, with automatic compression of old files.

```go
Rotation: &logger.RotationConfig{
    Filename:   "/var/log/app.log",
    MaxSize:    100,  // 100 MB
    MaxBackups: 7,
    MaxAge:     30,
    Compress:   true,
}
```

### 5. Multiple Outputs
Log to multiple destinations simultaneously (e.g., console + file).

```go
OutputPath:  "/var/log/app.log",
AdditionalOuts: []string{"stdout"},
```

### 6. Error Handling Helpers
Simplified error logging patterns for common use cases.

```go
// Log and return error
return logger.LogError(log, "Failed to connect", err)

// Panic if error (for initialization)
logger.Must(log, initConfig())

// Create logger or panic
log := logger.MustLogger(config)
```

## Output

The example produces various log formats:
- JSON format for production monitoring
- Console format with colors for development
- Structured fields for filtering and searching
- Stack traces for errors
- Trace IDs for request correlation

## Try It Yourself

Modify the examples to experiment with:
- Different log levels (debug, info, warn, error, fatal)
- Custom metadata fields
- Your own service names and environments
- Different output formats
