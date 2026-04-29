// Package logger provides a structured logging interface for cronitor-local.
package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

// Logger writes structured log lines to an output writer.
type Logger struct {
	out   io.Writer
	level Level
}

// New returns a Logger that writes to out at or above minLevel.
func New(out io.Writer, minLevel Level) *Logger {
	if out == nil {
		out = os.Stdout
	}
	return &Logger{out: out, level: minLevel}
}

func (l *Logger) log(level Level, job, msg string) {
	if level < l.level {
		return
	}
	timestamp := time.Now().UTC().Format(time.RFC3339)
	if job != "" {
		fmt.Fprintf(l.out, "%s [%s] job=%s %s\n", timestamp, levelNames[level], job, msg)
	} else {
		fmt.Fprintf(l.out, "%s [%s] %s\n", timestamp, levelNames[level], msg)
	}
}

// Debug logs a debug-level message.
func (l *Logger) Debug(job, msg string) { l.log(LevelDebug, job, msg) }

// Info logs an info-level message.
func (l *Logger) Info(job, msg string) { l.log(LevelInfo, job, msg) }

// Warn logs a warning-level message.
func (l *Logger) Warn(job, msg string) { l.log(LevelWarn, job, msg) }

// Error logs an error-level message.
func (l *Logger) Error(job, msg string) { l.log(LevelError, job, msg) }

// Infof logs a formatted info-level message with no job context.
func (l *Logger) Infof(format string, args ...any) {
	l.log(LevelInfo, "", fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error-level message with no job context.
func (l *Logger) Errorf(format string, args ...any) {
	l.log(LevelError, "", fmt.Sprintf(format, args...))
}
