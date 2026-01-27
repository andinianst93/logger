package logger

// Config configures the logger behavior and output
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

// RotationConfig configures log file rotation
type RotationConfig struct {
	Filename   string // Full path to log file
	MaxSize    int    // Maximum size in megabytes before rotation (default: 100MB)
	MaxBackups int    // Maximum number of old log files to keep (default: 3)
	MaxAge     int    // Maximum days to keep old log files (default: 30)
	Compress   bool   // Compress old log files (default: true)
}
