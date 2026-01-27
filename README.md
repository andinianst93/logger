# Logger

A simple and configurable structured logging library for Go, built on top of [uber-go/zap](https://github.com/uber-go/zap).

## Features

- Multiple log levels: Debug, Info, Warn, Error, Fatal
- JSON and Console output formats
- Configurable output destination (stdout, stderr, or file)
- Environment variable configuration support
- Colored console output for development
- Automatic caller information and stack traces
- **Persistent metadata fields** (service name, environment, version, host, PID)
- **Request/Trace ID** support for distributed tracing
- Production-ready structured logging

## Installation

```bash
go get github.com/andinianst93/logger
```

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/andinianst93/logger"
)

func main() {
    // Create logger with custom config
    log, err := logger.New(logger.Config{
        Level:      "info",
        Format:     "console",
        OutputPath: "stdout",
    })
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    log.Info("Application started")
    log.Debug("This won't show because level is 'info'")
    log.Warn("Warning message")
    log.Error("Error occurred")
}
```

### Using Environment Variables

```go
package main

import (
    "github.com/andinianst93/logger"
)

func main() {
    // Create logger from environment variables
    log, err := logger.NewFromEnv()
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    log.Info("Logger initialized from environment")
}
```

Set environment variables:
```bash
export LOG_LEVEL=debug
export LOG_FORMAT=json
export LOG_OUTPUT=stdout
```

## Configuration

### Log Levels

- `debug` - Development, technical details, verbose logging
- `info` - Normal operations (startup, config loaded)
- `warn` - Something unusual but not an error
- `error` - Error that can be recovered
- `fatal` - Fatal error, program must exit

### Output Formats

- `json` - JSON format for production (default)
- `console` - Human-readable format with colors for development

### Output Destinations

- `stdout` - Standard output (default)
- `stderr` - Standard error
- `/path/to/file.log` - Write to file

### Metadata Fields

You can add persistent metadata fields that will appear in every log entry:

- `ServiceName` - Name of your service/application (useful in microservices)
- `Environment` - Environment name (development, staging, production)
- `Version` - Application version (e.g., "v1.0.0", git commit hash)
- `Host` - Hostname or instance identifier (auto-detected if empty)
- `PID` - Process ID (auto-detected if 0)

## Helper Functions

### WithTraceID / WithRequestID

Add trace/request ID to logger for distributed tracing:

```go
// Generate new trace ID
traceID := logger.NewTraceID()

// Add to logger
requestLog := logger.WithTraceID(log, traceID)
// or
requestLog := logger.WithRequestID(log, requestID)
```

### WithFields

Add custom fields to logger:

```go
userLog := logger.WithFields(log,
    zap.String("user_id", "12345"),
    zap.String("tenant_id", "acme-corp"),
)

// All logs from userLog will include these fields
userLog.Info("Processing user action")
```

## Examples

### Production Logger with Metadata

```go
import (
    "github.com/andinianst93/logger"
    "go.uber.org/zap"
)

log, _ := logger.New(logger.Config{
    Level:       "info",
    Format:      "json",
    OutputPath:  "/var/log/app.log",
    
    // Metadata fields (added to every log entry)
    ServiceName: "user-service",
    Environment: "production",
    Version:     "v1.2.3",
    Host:        "", // auto-detect hostname
    PID:         0,  // auto-detect process ID
})

log.Info("User logged in", 
    zap.String("user_id", "12345"),
    zap.String("ip", "192.168.1.1"),
)
```

Output:
```json
{
  "level":"INFO",
  "timestamp":"2026-01-27T20:15:30+07:00",
  "caller":"main.go:15",
  "msg":"User logged in",
  "service":"user-service",
  "environment":"production",
  "version":"v1.2.3",
  "host":"server-01",
  "pid":12345,
  "user_id":"12345",
  "ip":"192.168.1.1"
}
```

### Request Tracing with Trace ID

```go
import (
    "github.com/andinianst93/logger"
    "go.uber.org/zap"
)

// Create base logger
log, _ := logger.New(logger.Config{
    Level:       "info",
    Format:      "json",
    ServiceName: "api-gateway",
    Environment: "production",
})

// Add trace ID for this request
traceID := logger.NewTraceID() // generates random ID
requestLog := logger.WithTraceID(log, traceID)

// All logs from this request will have the same trace_id
requestLog.Info("Request started", zap.String("path", "/api/users"))
requestLog.Info("Database query executed")
requestLog.Info("Request completed", zap.Int("status", 200))
```

Output:
```json
{"level":"INFO","timestamp":"2026-01-27T20:15:30+07:00","msg":"Request started","service":"api-gateway","environment":"production","trace_id":"a1b2c3d4e5f6...","path":"/api/users"}
{"level":"INFO","timestamp":"2026-01-27T20:15:31+07:00","msg":"Database query executed","service":"api-gateway","environment":"production","trace_id":"a1b2c3d4e5f6..."}
{"level":"INFO","timestamp":"2026-01-27T20:15:32+07:00","msg":"Request completed","service":"api-gateway","environment":"production","trace_id":"a1b2c3d4e5f6...","status":200}
```

### HTTP Middleware Example

```go
func LoggingMiddleware(baseLogger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Generate or extract trace ID from header
            traceID := r.Header.Get("X-Trace-ID")
            if traceID == "" {
                traceID = logger.NewTraceID()
            }
            
            // Create request-scoped logger
            reqLog := logger.WithTraceID(baseLogger, traceID)
            reqLog = logger.WithFields(reqLog,
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.String("ip", r.RemoteAddr),
            )
            
            reqLog.Info("Request received")
            
            // Pass logger via context
            ctx := context.WithValue(r.Context(), "logger", reqLog)
            next.ServeHTTP(w, r.WithContext(ctx))
            
            reqLog.Info("Request completed")
        })
    }
}
```

### JSON Output for Production

```go
log, _ := logger.New(logger.Config{
    Level:      "info",
    Format:     "json",
    OutputPath: "/var/log/app.log",
})

log.Info("User logged in", 
    zap.String("user_id", "12345"),
    zap.String("ip", "192.168.1.1"),
)
```

Output:
```json
{"level":"INFO","timestamp":"2026-01-27T20:15:30+07:00","caller":"main.go:15","msg":"User logged in","user_id":"12345","ip":"192.168.1.1"}
```

### Console Output for Development

```go
log, _ := logger.New(logger.Config{
    Level:      "debug",
    Format:     "console",
    OutputPath: "stdout",
})

log.Debug("Processing request", zap.Int("request_id", 42))
```

Output (with colors):
```
2026-01-27T20:15:30+07:00	DEBUG	main.go:15	Processing request	{"request_id": 42}
```

## License

MIT
