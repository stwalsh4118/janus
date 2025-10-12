package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is the global logger instance
var Logger zerolog.Logger

// Init initializes the global logger with pretty console output
func Init(logLevel string) {
	// Parse log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Set global log level
	zerolog.SetGlobalLevel(level)

	// Configure pretty console output with colors
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false, // Enable colors
	}

	// Create logger with pretty output
	Logger = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	// Set as global logger
	log.Logger = Logger

	Logger.Info().
		Str("level", level.String()).
		Msg("Logger initialized")
}

// InitJSON initializes the logger with JSON output (for production)
func InitJSON(logLevel string, output io.Writer) {
	// Parse log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Set global log level
	zerolog.SetGlobalLevel(level)

	// Create logger with JSON output
	Logger = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	// Set as global logger
	log.Logger = Logger

	Logger.Info().
		Str("level", level.String()).
		Msg("Logger initialized")
}

// Get returns the global logger instance
func Get() *zerolog.Logger {
	return &Logger
}
