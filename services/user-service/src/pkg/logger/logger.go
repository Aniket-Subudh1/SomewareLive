package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/your-username/slido-clone/user-service/config"
)

// Init initializes the logger
func Init(config *config.Config) {
	// Set the global logger
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"

	// Configure logger output
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s |", i))
		},
	}

	// Set the global logger
	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Str("service", "user-service").
		Logger()

	// Set the log level
	level, err := zerolog.ParseLevel(config.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().Msg("Logger initialized")
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// Debug logs a debug message
func Debug(msg string, fields ...Field) {
	event := log.Debug()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Info logs an info message
func Info(msg string, fields ...Field) {
	event := log.Info()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Warn logs a warning message
func Warn(msg string, fields ...Field) {
	event := log.Warn()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Error logs an error message
func Error(msg string, err error, fields ...Field) {
	event := log.Error()
	if err != nil {
		event = event.Err(err)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, err error, fields ...Field) {
	event := log.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// F creates a new Field
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
