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

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new logger with the given configuration
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
