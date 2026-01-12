// Package log provides structured logging with levels for Rugby programs.
// Rugby: import rugby/log
//
// Example:
//
//	log.info("user logged in", {"user_id": 123})
//	log.error("failed to connect", {"host": "db.example.com"})
//	log.set_level(:debug)
package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Level represents a logging level.
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	// Off disables all logging
	Off Level = 100
)

// String returns the level name.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Format represents the output format.
type Format int

const (
	TextFormat Format = iota
	JSONFormat
)

// Fields is a map of structured log fields.
type Fields map[string]any

// Logger is a structured logger.
type Logger struct {
	mu      sync.Mutex
	out     io.Writer
	level   Level
	format  Format
	fields  Fields
	name    string
	timeFn  func() time.Time
}

// New creates a new logger with default settings.
// Ruby: log.new()
func New() *Logger {
	return &Logger{
		out:    os.Stderr,
		level:  InfoLevel,
		format: TextFormat,
		fields: make(Fields),
		timeFn: time.Now,
	}
}

// global logger
var std = New()

// SetOutput sets the output destination.
// Ruby: logger.output = io
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// SetLevel sets the minimum log level.
// Ruby: logger.level = :info
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level.
// Ruby: logger.level
func (l *Logger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// SetFormat sets the output format.
// Ruby: logger.format = :json
func (l *Logger) SetFormat(f Format) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.format = f
}

// SetName sets the logger name (appears in output).
// Ruby: logger.name = "myapp"
func (l *Logger) SetName(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.name = name
}

// With returns a new logger with additional fields.
// Ruby: logger.with({"request_id": "abc123"})
func (l *Logger) With(fields Fields) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		out:    l.out,
		level:  l.level,
		format: l.format,
		fields: newFields,
		name:   l.name,
		timeFn: l.timeFn,
	}
}

// Debug logs at debug level.
// Ruby: log.debug(msg) or log.debug(msg, fields)
func (l *Logger) Debug(msg string, fields ...Fields) {
	l.log(DebugLevel, msg, mergeFields(fields))
}

// Info logs at info level.
// Ruby: log.info(msg) or log.info(msg, fields)
func (l *Logger) Info(msg string, fields ...Fields) {
	l.log(InfoLevel, msg, mergeFields(fields))
}

// Warn logs at warn level.
// Ruby: log.warn(msg) or log.warn(msg, fields)
func (l *Logger) Warn(msg string, fields ...Fields) {
	l.log(WarnLevel, msg, mergeFields(fields))
}

// Error logs at error level.
// Ruby: log.error(msg) or log.error(msg, fields)
func (l *Logger) Error(msg string, fields ...Fields) {
	l.log(ErrorLevel, msg, mergeFields(fields))
}

// Fatal logs at fatal level and exits with status 1.
// Ruby: log.fatal(msg) or log.fatal(msg, fields)
func (l *Logger) Fatal(msg string, fields ...Fields) {
	l.log(FatalLevel, msg, mergeFields(fields))
	os.Exit(1)
}

func mergeFields(fields []Fields) Fields {
	if len(fields) == 0 {
		return nil
	}
	result := make(Fields)
	for _, f := range fields {
		for k, v := range f {
			result[k] = v
		}
	}
	return result
}

func (l *Logger) log(level Level, msg string, fields Fields) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// Merge logger fields with call-site fields
	allFields := make(Fields)
	for k, v := range l.fields {
		allFields[k] = v
	}
	for k, v := range fields {
		allFields[k] = v
	}

	now := l.timeFn()

	var output string
	if l.format == JSONFormat {
		output = l.formatJSON(level, msg, allFields, now)
	} else {
		output = l.formatText(level, msg, allFields, now)
	}

	fmt.Fprintln(l.out, output)
}

func (l *Logger) formatText(level Level, msg string, fields Fields, t time.Time) string {
	var b strings.Builder

	// Timestamp
	b.WriteString(t.Format("2006-01-02T15:04:05.000Z07:00"))
	b.WriteString(" ")

	// Level
	b.WriteString(fmt.Sprintf("%-5s", level.String()))
	b.WriteString(" ")

	// Logger name
	if l.name != "" {
		b.WriteString("[")
		b.WriteString(l.name)
		b.WriteString("] ")
	}

	// Message
	b.WriteString(msg)

	// Fields (sorted for deterministic output)
	if len(fields) > 0 {
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		b.WriteString(" ")
		for i, k := range keys {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString(k)
			b.WriteString("=")
			b.WriteString(formatValue(fields[k]))
		}
	}

	return b.String()
}

func (l *Logger) formatJSON(level Level, msg string, fields Fields, t time.Time) string {
	entry := make(map[string]any)

	// Add user fields first, so reserved keys can't be overwritten
	for k, v := range fields {
		entry[k] = v
	}

	// Reserved keys (added last to ensure they aren't overwritten)
	entry["time"] = t.Format(time.RFC3339Nano)
	entry["level"] = strings.ToLower(level.String())
	entry["msg"] = msg

	if l.name != "" {
		entry["logger"] = l.name
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Sprintf(`{"error":"failed to marshal log entry: %s"}`, err)
	}
	return string(data)
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		if strings.ContainsAny(val, " \t\n\"") {
			return fmt.Sprintf("%q", val)
		}
		return val
	case error:
		return val.Error()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Global logger functions

// SetOutput sets the output for the global logger.
// Ruby: log.output = io
func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

// SetLevel sets the level for the global logger.
// Ruby: log.set_level(:debug)
func SetLevel(level Level) {
	std.SetLevel(level)
}

// GetLevel returns the level of the global logger.
// Ruby: log.level
func GetLevel() Level {
	return std.GetLevel()
}

// SetFormat sets the format for the global logger.
// Ruby: log.set_format(:json)
func SetFormat(f Format) {
	std.SetFormat(f)
}

// SetName sets the name for the global logger.
// Ruby: log.set_name("myapp")
func SetName(name string) {
	std.SetName(name)
}

// With returns a new logger with additional fields.
// Ruby: log.with(fields)
func With(fields Fields) *Logger {
	return std.With(fields)
}

// Debug logs at debug level on the global logger.
// Ruby: log.debug(msg) or log.debug(msg, fields)
func Debug(msg string, fields ...Fields) {
	std.Debug(msg, fields...)
}

// Info logs at info level on the global logger.
// Ruby: log.info(msg) or log.info(msg, fields)
func Info(msg string, fields ...Fields) {
	std.Info(msg, fields...)
}

// Warn logs at warn level on the global logger.
// Ruby: log.warn(msg) or log.warn(msg, fields)
func Warn(msg string, fields ...Fields) {
	std.Warn(msg, fields...)
}

// Error logs at error level on the global logger.
// Ruby: log.error(msg) or log.error(msg, fields)
func Error(msg string, fields ...Fields) {
	std.Error(msg, fields...)
}

// Fatal logs at fatal level and exits.
// Ruby: log.fatal(msg) or log.fatal(msg, fields)
func Fatal(msg string, fields ...Fields) {
	std.Fatal(msg, fields...)
}

// ParseLevel parses a level string.
// Ruby: log.parse_level("debug")
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "off":
		return Off, nil
	default:
		return InfoLevel, fmt.Errorf("log: unknown level %q", s)
	}
}

// LevelFromSymbol converts a Rugby symbol to a level.
// Ruby: log.set_level(:debug)
func LevelFromSymbol(sym string) Level {
	level, err := ParseLevel(sym)
	if err != nil {
		return InfoLevel
	}
	return level
}

// IsEnabled reports whether logging is enabled at the given level.
// Ruby: log.enabled?(:debug)
func (l *Logger) IsEnabled(level Level) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return level >= l.level
}

// IsEnabled reports whether the global logger is enabled at the given level.
// Ruby: log.enabled?(:debug)
func IsEnabled(level Level) bool {
	return std.IsEnabled(level)
}

// Reset resets the global logger to defaults.
// Useful for testing.
func Reset() {
	std = New()
}
