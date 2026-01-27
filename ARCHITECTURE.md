# Logger Architecture

This document explains the structure and organization of the logger library.

## File Structure

The library is organized into separate files based on their concerns:

```
logger/
├── logger.go       # Core logger creation logic (1.9KB)
├── config.go       # Configuration structs (1.2KB)
├── global.go       # Global logger instance (1.6KB)
├── context.go      # Request tracing & context (1.1KB)
├── error.go        # Error handling helpers (983B)
├── helpers.go      # Internal helper functions (3.8KB)
├── logger_test.go  # Tests
└── example/        # Usage examples
```

## File Responsibilities

### 1. `logger.go` - Core Logger Creation
**Purpose**: Main entry point for creating logger instances

**Functions**:
- `New(cfg Config) (*zap.Logger, error)` - Creates a configured logger

**Flow**:
1. Parse log level
2. Setup encoder (JSON/Console)
3. Setup output writers (stdout/file/rotation)
4. Combine multiple outputs if needed
5. Create zap core
6. Add metadata fields

### 2. `config.go` - Configuration
**Purpose**: Define configuration structs for logger setup

**Structs**:
- `Config` - Main logger configuration
  - Basic: Level, Format, OutputPath
  - Metadata: ServiceName, Environment, Version, Host, PID
  - Advanced: Rotation, AdditionalOuts
- `RotationConfig` - Log rotation settings

### 3. `global.go` - Global Logger
**Purpose**: Provide global logger instance for convenience

**Features**:
- Thread-safe global logger with RWMutex
- Convenience functions: `L()`, `Debug()`, `Info()`, `Warn()`, `Error()`, `Fatal()`
- `SetGlobalLogger()` / `GetGlobalLogger()`
- Auto-initialized with default logger

**Use Case**: When you don't want to pass logger everywhere

### 4. `context.go` - Request Tracing
**Purpose**: Add contextual information to logs

**Functions**:
- `NewTraceID()` - Generate random trace ID
- `WithTraceID()` - Add trace_id to logger
- `WithRequestID()` - Add request_id to logger  
- `WithFields()` - Add custom fields to logger

**Use Case**: Distributed tracing, request correlation

### 5. `error.go` - Error Handling
**Purpose**: Simplified error logging patterns

**Functions**:
- `LogError()` - Log error and return it (useful in return statements)
- `LogErrorAndExit()` - Log error and exit program
- `Must()` - Panic if error (for initialization)
- `MustLogger()` - Create logger or panic

**Use Case**: Cleaner error handling code

### 6. `helpers.go` - Internal Helpers
**Purpose**: Internal utility functions (not exported)

**Functions**:
- `getWriter()` - Create WriteSyncer from output path
- `getEncoder()` - Create encoder from format
- `parseLevel()` - Convert string to log level
- `addMetadataFields()` - Add metadata to logger
- `getOrDefault()` - Helper for default values
- `getEnvOrDefault()` - Read environment variables
- `NewFromEnv()` - Create logger from env vars

## Separation of Concerns

### Why This Structure?

1. **Readability**: Each file has a clear, single purpose
2. **Maintainability**: Easy to find and modify specific functionality
3. **Testability**: Can test each component independently
4. **Discoverability**: New contributors can quickly understand structure

### Design Principles

1. **Core vs Convenience**
   - `logger.go` = Core functionality
   - `global.go` = Convenience wrappers
   
2. **Configuration vs Logic**
   - `config.go` = Data structures
   - `logger.go` + `helpers.go` = Logic
   
3. **Public vs Internal**
   - Exported functions in `logger.go`, `global.go`, `context.go`, `error.go`
   - Internal helpers in `helpers.go` (lowercase functions)

## Usage Patterns

### Pattern 1: Instance-based (Explicit)
```go
// Create logger
log := logger.MustLogger(logger.Config{...})

// Pass to functions
doWork(log)

// Use in function
func doWork(log *zap.Logger) {
    log.Info("working")
}
```

**Files used**: `logger.go`, `config.go`, `error.go`

### Pattern 2: Global Logger (Convenient)
```go
// Setup once
logger.SetGlobalLogger(myLogger)

// Use anywhere
logger.Info("hello")
```

**Files used**: `global.go`

### Pattern 3: Request Tracing
```go
// Add trace ID per request
reqLog := logger.WithTraceID(log, traceID)
reqLog.Info("processing")
```

**Files used**: `context.go`

### Pattern 4: Error Handling
```go
// Log and return error
return logger.LogError(log, "failed", err)
```

**Files used**: `error.go`

## Adding New Features

### Where to add new code?

- **New config option** → `config.go`
- **New logger creation logic** → `logger.go`
- **New convenience function** → `global.go`
- **New context helper** → `context.go`
- **New error pattern** → `error.go`
- **New internal helper** → `helpers.go`

## Dependencies

```
logger.go
  ├── config.go (Config, RotationConfig)
  └── helpers.go (getWriter, getEncoder, parseLevel, addMetadataFields)

global.go
  └── logger.go (New)

context.go
  └── (no internal dependencies)

error.go
  └── logger.go (New)

helpers.go
  ├── config.go (Config, RotationConfig)
  └── external: zap, zapcore, lumberjack
```

## Migration Note

Previously all code was in a single `logger.go` (378 lines). Now split into:
- `logger.go`: 80 lines (core)
- `config.go`: 28 lines (config)
- `global.go`: 74 lines (global)
- `context.go`: 42 lines (context)
- `error.go`: 40 lines (errors)
- `helpers.go`: 155 lines (internals)

**Total**: ~419 lines (small increase due to better organization and comments)
