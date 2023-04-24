package logger

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog to provide custome log functions.
type Logger struct{}

// New return a Logger instance.
func New() *Logger {
	return &Logger{}
}

// Print formats a log for printing.
func (logger *Logger) Print(level zerolog.Level, args ...interface{}) {
	log.WithLevel(level).Msg(fmt.Sprint(args...))
}

// Printf formats a log for printing with string directive.
func (logger *Logger) Printf(ctx context.Context, format string, v ...interface{}) {
	log.WithLevel(zerolog.DebugLevel).Msgf(format, v...)
}

// Debug prints a log at Debug Level.
func (logger *Logger) Debug(arg ...interface{}) {
	logger.Print(zerolog.DebugLevel, arg...)
}

// Info prints a log at Info Level.
func (logger *Logger) Info(arg ...interface{}) {
	logger.Print(zerolog.InfoLevel, arg...)
}

// Warn prints a log at Warn Level.
func (logger *Logger) Warn(arg ...interface{}) {
	logger.Print(zerolog.WarnLevel, arg...)
}

// Error prints a log at Error Level.
func (logger *Logger) Error(arg ...interface{}) {
	logger.Print(zerolog.ErrorLevel, arg...)
}

// Fatal prints a log at Fatal Level,
// and terminates process with a status set of 1.
func (logger *Logger) Fatal(arg ...interface{}) {
	logger.Print(zerolog.FatalLevel, arg...)
}
