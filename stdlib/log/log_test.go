package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	l := New()
	if l == nil {
		t.Fatal("New() returned nil")
	}
	if l.level != InfoLevel {
		t.Errorf("default level = %v, want InfoLevel", l.level)
	}
	if l.format != TextFormat {
		t.Errorf("default format = %v, want TextFormat", l.format)
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestSetLevel(t *testing.T) {
	l := New()
	l.SetLevel(DebugLevel)
	if l.GetLevel() != DebugLevel {
		t.Errorf("GetLevel() = %v, want DebugLevel", l.GetLevel())
	}
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetLevel(WarnLevel)

	// These should be filtered
	l.Debug("debug message")
	l.Info("info message")

	// These should appear
	l.Warn("warn message")
	l.Error("error message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should appear")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should appear")
	}
}

func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.timeFn = func() time.Time {
		return time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	}

	l.Info("test message", Fields{"key": "value"})

	output := buf.String()

	// Check timestamp
	if !strings.Contains(output, "2024-01-15T10:30:00.000Z") {
		t.Errorf("Missing timestamp in output: %s", output)
	}

	// Check level
	if !strings.Contains(output, "INFO") {
		t.Errorf("Missing level in output: %s", output)
	}

	// Check message
	if !strings.Contains(output, "test message") {
		t.Errorf("Missing message in output: %s", output)
	}

	// Check field
	if !strings.Contains(output, "key=value") {
		t.Errorf("Missing field in output: %s", output)
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetFormat(JSONFormat)
	l.timeFn = func() time.Time {
		return time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	}

	l.Info("test message", Fields{"user_id": 123})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if entry["level"] != "info" {
		t.Errorf("level = %v, want info", entry["level"])
	}
	if entry["msg"] != "test message" {
		t.Errorf("msg = %v, want 'test message'", entry["msg"])
	}
	if entry["user_id"] != float64(123) {
		t.Errorf("user_id = %v, want 123", entry["user_id"])
	}
}

func TestLoggerName(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetName("myapp")

	l.Info("test")

	output := buf.String()
	if !strings.Contains(output, "[myapp]") {
		t.Errorf("Missing logger name in output: %s", output)
	}
}

func TestLoggerNameJSON(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetFormat(JSONFormat)
	l.SetName("myapp")

	l.Info("test")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry["logger"] != "myapp" {
		t.Errorf("logger = %v, want myapp", entry["logger"])
	}
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)

	child := l.With(Fields{"request_id": "abc123"})
	child.Info("request started")

	output := buf.String()
	if !strings.Contains(output, "request_id=abc123") {
		t.Errorf("Missing inherited field in output: %s", output)
	}
}

func TestWithDoesNotMutateParent(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)

	_ = l.With(Fields{"child_field": "value"})

	buf.Reset()
	l.Info("parent log")

	output := buf.String()
	if strings.Contains(output, "child_field") {
		t.Error("Child field should not appear in parent logs")
	}
}

func TestWithMergesFields(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)

	child := l.With(Fields{"field1": "value1"})
	grandchild := child.With(Fields{"field2": "value2"})

	grandchild.Info("test")

	output := buf.String()
	if !strings.Contains(output, "field1=value1") {
		t.Error("Missing field1 in output")
	}
	if !strings.Contains(output, "field2=value2") {
		t.Error("Missing field2 in output")
	}
}

func TestCallSiteFieldsOverrideLoggerFields(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetFormat(JSONFormat)

	child := l.With(Fields{"key": "logger_value"})
	child.Info("test", Fields{"key": "call_value"})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry["key"] != "call_value" {
		t.Errorf("key = %v, want call_value", entry["key"])
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input   string
		want    Level
		wantErr bool
	}{
		{"debug", DebugLevel, false},
		{"DEBUG", DebugLevel, false},
		{"info", InfoLevel, false},
		{"INFO", InfoLevel, false},
		{"warn", WarnLevel, false},
		{"warning", WarnLevel, false},
		{"error", ErrorLevel, false},
		{"fatal", FatalLevel, false},
		{"off", Off, false},
		{"invalid", InfoLevel, true},
	}

	for _, tt := range tests {
		got, err := ParseLevel(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseLevel(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestLevelFromSymbol(t *testing.T) {
	if LevelFromSymbol("debug") != DebugLevel {
		t.Error("LevelFromSymbol(debug) != DebugLevel")
	}
	if LevelFromSymbol("invalid") != InfoLevel {
		t.Error("LevelFromSymbol(invalid) should default to InfoLevel")
	}
}

func TestIsEnabled(t *testing.T) {
	l := New()
	l.SetLevel(WarnLevel)

	if l.IsEnabled(DebugLevel) {
		t.Error("Debug should not be enabled")
	}
	if l.IsEnabled(InfoLevel) {
		t.Error("Info should not be enabled")
	}
	if !l.IsEnabled(WarnLevel) {
		t.Error("Warn should be enabled")
	}
	if !l.IsEnabled(ErrorLevel) {
		t.Error("Error should be enabled")
	}
}

func TestGlobalLogger(t *testing.T) {
	Reset()
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(DebugLevel)
	SetFormat(TextFormat)
	SetName("global")

	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")

	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Error("Missing DEBUG in global output")
	}
	if !strings.Contains(output, "[global]") {
		t.Error("Missing logger name in global output")
	}
}

func TestGlobalWith(t *testing.T) {
	Reset()
	var buf bytes.Buffer
	SetOutput(&buf)

	child := With(Fields{"trace_id": "xyz"})
	child.Info("traced")

	output := buf.String()
	if !strings.Contains(output, "trace_id=xyz") {
		t.Error("Missing trace_id in output")
	}
}

func TestGlobalGetLevel(t *testing.T) {
	Reset()
	SetLevel(ErrorLevel)
	if GetLevel() != ErrorLevel {
		t.Errorf("GetLevel() = %v, want ErrorLevel", GetLevel())
	}
}

func TestGlobalIsEnabled(t *testing.T) {
	Reset()
	SetLevel(WarnLevel)
	if IsEnabled(DebugLevel) {
		t.Error("Debug should not be enabled")
	}
	if !IsEnabled(ErrorLevel) {
		t.Error("Error should be enabled")
	}
}

func TestFormatValueString(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{"simple", "simple"},
		{"has space", `"has space"`},
		{"has\ttab", `"has\ttab"`},
		{"has\nnewline", `"has\nnewline"`},
		{"has\"quote", `"has\"quote"`},
		{123, "123"},
		{true, "true"},
		{errors.New("something failed"), "something failed"},
	}

	for _, tt := range tests {
		got := formatValue(tt.input)
		if got != tt.want {
			t.Errorf("formatValue(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMultipleFields(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)

	l.Info("test", Fields{"a": 1}, Fields{"b": 2})

	output := buf.String()
	if !strings.Contains(output, "a=1") {
		t.Error("Missing a=1")
	}
	if !strings.Contains(output, "b=2") {
		t.Error("Missing b=2")
	}
}

func TestNoFields(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)

	l.Info("message only")

	output := buf.String()
	if !strings.Contains(output, "message only") {
		t.Error("Missing message")
	}
}

func TestOff(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetLevel(Off)

	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")

	if buf.Len() > 0 {
		t.Error("No output expected when level is Off")
	}
}

func TestReservedKeysCannotBeOverwritten(t *testing.T) {
	var buf bytes.Buffer
	l := New()
	l.SetOutput(&buf)
	l.SetFormat(JSONFormat)

	// Try to overwrite reserved keys
	l.Info("test", Fields{"level": "custom", "msg": "custom", "time": "custom"})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Reserved keys should not be overwritten
	if entry["level"] != "info" {
		t.Errorf("level = %v, want info (reserved key was overwritten)", entry["level"])
	}
	if entry["msg"] != "test" {
		t.Errorf("msg = %v, want test (reserved key was overwritten)", entry["msg"])
	}
	// time should be a valid timestamp, not "custom"
	if entry["time"] == "custom" {
		t.Error("time was overwritten by user field")
	}
}

func TestFieldOrderDeterministic(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	l1 := New()
	l1.SetOutput(&buf1)
	l1.timeFn = func() time.Time {
		return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	l2 := New()
	l2.SetOutput(&buf2)
	l2.timeFn = func() time.Time {
		return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	fields := Fields{"z": 1, "a": 2, "m": 3}
	l1.Info("test", fields)
	l2.Info("test", fields)

	if buf1.String() != buf2.String() {
		t.Errorf("Field order not deterministic:\n%s\nvs\n%s", buf1.String(), buf2.String())
	}
}
