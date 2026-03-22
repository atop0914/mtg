package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New creates a new logger with default settings
func New() zerolog.Logger {
	return NewWithLevel(zerolog.InfoLevel)
}

// NewWithLevel creates a new logger with specified level
func NewWithLevel(level zerolog.Level) zerolog.Logger {
	return zerolog.New(os.Stdout).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()
}

// NewWithOutput creates a new logger with custom output
func NewWithOutput(output *os.File) zerolog.Logger {
	return zerolog.New(output).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Caller().
		Logger()
}

// Logger is a global logger instance
var Logger zerolog.Logger

func init() {
	Logger = New()
}

// SetLevel sets the global logger level
func SetLevel(level string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		Logger.Warn().Str("level", level).Msg("Invalid log level, using info")
		lvl = zerolog.InfoLevel
	}
	Logger = Logger.Level(lvl)
}

// WithContext adds context to the logger
func WithContext(key, value string) zerolog.Logger {
	return Logger.With().Str(key, value).Logger()
}

// Sync syncs the logger (for file output)
func Sync() error {
	// zerolog console writer may need sync
	return nil
}

// TimeFormat is the format used for timestamps
const TimeFormat = time.RFC3339
