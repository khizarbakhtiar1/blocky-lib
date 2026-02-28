package logging

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Level represents log levels
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelNone
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger is the interface for logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	WithContext(ctx context.Context) Logger
	WithFields(fields ...Field) Logger
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// Helper functions to create fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Uint64(key string, value uint64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// DefaultLogger is a simple implementation of Logger
type DefaultLogger struct {
	level  Level
	output io.Writer
	fields []Field
	mu     sync.Mutex
}

// NewLogger creates a new default logger
func NewLogger(level Level) *DefaultLogger {
	return &DefaultLogger{
		level:  level,
		output: os.Stdout,
		fields: nil,
	}
}

// SetOutput sets the output writer
func (l *DefaultLogger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// SetLevel sets the minimum log level
func (l *DefaultLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *DefaultLogger) log(level Level, msg string, fields []Field) {
	if level < l.level {
		return
	}
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	timestamp := time.Now().Format(time.RFC3339)
	
	// Combine default fields with provided fields
	allFields := append(l.fields, fields...)
	
	fieldStr := ""
	for _, f := range allFields {
		fieldStr += fmt.Sprintf(" %s=%v", f.Key, f.Value)
	}
	
	fmt.Fprintf(l.output, "%s [%s] %s%s\n", timestamp, level.String(), msg, fieldStr)
}

func (l *DefaultLogger) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, fields)
}

func (l *DefaultLogger) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, fields)
}

func (l *DefaultLogger) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, fields)
}

func (l *DefaultLogger) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields)
}

func (l *DefaultLogger) WithContext(ctx context.Context) Logger {
	// Could extract request ID or other context values
	return l
}

func (l *DefaultLogger) WithFields(fields ...Field) Logger {
	return &DefaultLogger{
		level:  l.level,
		output: l.output,
		fields: append(l.fields, fields...),
	}
}

// NopLogger is a logger that does nothing
type NopLogger struct{}

func (NopLogger) Debug(msg string, fields ...Field) {}
func (NopLogger) Info(msg string, fields ...Field)  {}
func (NopLogger) Warn(msg string, fields ...Field)  {}
func (NopLogger) Error(msg string, fields ...Field) {}
func (n NopLogger) WithContext(ctx context.Context) Logger { return n }
func (n NopLogger) WithFields(fields ...Field) Logger      { return n }

// Global default logger
var defaultLogger Logger = NewLogger(LevelInfo)
var defaultMu sync.RWMutex

// SetDefaultLogger sets the global default logger
func SetDefaultLogger(l Logger) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger = l
}

// GetDefaultLogger returns the global default logger
func GetDefaultLogger() Logger {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultLogger
}

// Package-level logging functions
func Debug(msg string, fields ...Field) {
	GetDefaultLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	GetDefaultLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetDefaultLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	GetDefaultLogger().Error(msg, fields...)
}

// StdLogAdapter adapts the standard library logger to our Logger interface
type StdLogAdapter struct {
	logger *log.Logger
	level  Level
}

// NewStdLogAdapter creates a new adapter for the standard log package
func NewStdLogAdapter(logger *log.Logger, level Level) *StdLogAdapter {
	return &StdLogAdapter{
		logger: logger,
		level:  level,
	}
}

func (a *StdLogAdapter) log(level Level, msg string, fields []Field) {
	if level < a.level {
		return
	}
	
	fieldStr := ""
	for _, f := range fields {
		fieldStr += fmt.Sprintf(" %s=%v", f.Key, f.Value)
	}
	
	a.logger.Printf("[%s] %s%s", level.String(), msg, fieldStr)
}

func (a *StdLogAdapter) Debug(msg string, fields ...Field) { a.log(LevelDebug, msg, fields) }
func (a *StdLogAdapter) Info(msg string, fields ...Field)  { a.log(LevelInfo, msg, fields) }
func (a *StdLogAdapter) Warn(msg string, fields ...Field)  { a.log(LevelWarn, msg, fields) }
func (a *StdLogAdapter) Error(msg string, fields ...Field) { a.log(LevelError, msg, fields) }
func (a *StdLogAdapter) WithContext(ctx context.Context) Logger { return a }
func (a *StdLogAdapter) WithFields(fields ...Field) Logger      { return a }
