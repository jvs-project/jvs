// Package logging provides structured logging for JVS.
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents a log level.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Logger provides structured logging.
type Logger struct {
	mu     sync.Mutex
	level  Level
	output io.Writer
	fields map[string]any
}

// LogEntry represents a structured log entry.
type LogEntry struct {
	Timestamp string         `json:"timestamp"`
	Level     Level          `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// NewLogger creates a new logger with the specified level.
func NewLogger(level Level) *Logger {
	return &Logger{
		level:  level,
		output: os.Stderr,
		fields: make(map[string]any),
	}
}

// WithFields returns a new logger with additional fields.
func (l *Logger) WithFields(fields map[string]any) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]any)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		level:  l.level,
		output: l.output,
		fields: newFields,
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...map[string]any) {
	if l.level == LevelDebug {
		l.log(LevelDebug, msg, fields...)
	}
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields ...map[string]any) {
	if l.level == LevelDebug || l.level == LevelInfo {
		l.log(LevelInfo, msg, fields...)
	}
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields ...map[string]any) {
	if l.level != LevelError {
		l.log(LevelWarn, msg, fields...)
	}
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields ...map[string]any) {
	l.log(LevelError, msg, fields...)
}

// ErrorErr logs an error message with an error value.
func (l *Logger) ErrorErr(msg string, err error, fields ...map[string]any) {
	combined := map[string]any{"error": err.Error()}
	for _, f := range fields {
		for k, v := range f {
			combined[k] = v
		}
	}
	l.log(LevelError, msg, combined)
}

func (l *Logger) log(level Level, msg string, fields ...map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Message:   msg,
		Fields:    make(map[string]any),
	}

	// Add base fields
	for k, v := range l.fields {
		entry.Fields[k] = v
	}

	// Add call-specific fields
	for _, f := range fields {
		for k, v := range f {
			entry.Fields[k] = v
		}
	}

	// Remove fields if empty
	if len(entry.Fields) == 0 {
		entry.Fields = nil
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.output, `{"level":"error","message":"failed to marshal log entry"}`+"\n")
		return
	}

	l.output.Write(append(data, '\n'))
}

// SetOutput sets the output writer.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// SetLevel sets the log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Global logger instance
var global = NewLogger(LevelInfo)

// SetGlobal sets the global logger.
func SetGlobal(l *Logger) {
	global = l
}

// Debug logs to the global logger.
func Debug(msg string, fields ...map[string]any) {
	global.Debug(msg, fields...)
}

// Info logs to the global logger.
func Info(msg string, fields ...map[string]any) {
	global.Info(msg, fields...)
}

// Warn logs to the global logger.
func Warn(msg string, fields ...map[string]any) {
	global.Warn(msg, fields...)
}

// Error logs to the global logger.
func Error(msg string, fields ...map[string]any) {
	global.Error(msg, fields...)
}

// ErrorErr logs to the global logger with an error.
func ErrorErr(msg string, err error, fields ...map[string]any) {
	global.ErrorErr(msg, err, fields...)
}

// WithFields returns a new logger from global with additional fields.
func WithFields(fields map[string]any) *Logger {
	return global.WithFields(fields)
}
