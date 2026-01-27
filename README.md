# Logger

A simple and configurable structured logging library for Go, built on top of [uber-go/zap](https://github.com/uber-go/zap).

## Features

- Multiple log levels: Debug, Info, Warn, Error, Fatal
- JSON and Console output formats
- Configurable output destination (stdout, stderr, or file)
- Environment variable configuration support
- Colored console output for development
- Automatic caller information and stack traces

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

## Examples

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
