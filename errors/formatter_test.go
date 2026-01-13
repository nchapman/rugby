package errors

import (
	"errors"
	"strings"
	"testing"
)

// testError implements PositionedError for testing
type testError struct {
	msg    string
	line   int
	column int
}

func (e *testError) Error() string {
	return e.msg
}

func (e *testError) Position() (int, int) {
	return e.line, e.column
}

// testErrorWithHelp implements PositionedError and ErrorWithHelp
type testErrorWithHelp struct {
	testError
	help string
}

func (e *testErrorWithHelp) Help() string {
	return e.help
}

func TestSource_Line(t *testing.T) {
	src := NewSource("test.rg", "line one\nline two\nline three")

	tests := []struct {
		line int
		want string
	}{
		{1, "line one"},
		{2, "line two"},
		{3, "line three"},
		{0, ""},
		{4, ""},
		{-1, ""},
	}

	for _, tt := range tests {
		got := src.Line(tt.line)
		if got != tt.want {
			t.Errorf("Line(%d) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestSource_LineCount(t *testing.T) {
	tests := []struct {
		content string
		want    int
	}{
		{"one\ntwo\nthree", 3},
		{"single", 1},
		{"", 1}, // empty string splits to one empty element
		{"a\nb", 2},
	}

	for _, tt := range tests {
		src := NewSource("test.rg", tt.content)
		got := src.LineCount()
		if got != tt.want {
			t.Errorf("LineCount(%q) = %d, want %d", tt.content, got, tt.want)
		}
	}
}

func TestFormatter_FormatError_NoColors(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = 5 + \"hello\"")
	err := &testError{
		msg:    "type mismatch: cannot add Int and String",
		line:   1,
		column: 5,
	}

	result := f.FormatError(err, src)

	// Check for key parts without colors
	if !strings.Contains(result, "error:") {
		t.Error("expected 'error:' in output")
	}
	if !strings.Contains(result, "type mismatch") {
		t.Error("expected error message in output")
	}
	if !strings.Contains(result, "test.rg:1:5") {
		t.Error("expected file location in output")
	}
	if !strings.Contains(result, "x = 5 + \"hello\"") {
		t.Error("expected source line in output")
	}
	if !strings.Contains(result, "^") {
		t.Error("expected caret in output")
	}
}

func TestFormatter_FormatError_WithHelp(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "greet(user.age)")
	err := &testErrorWithHelp{
		testError: testError{
			msg:    "type mismatch: expected String, got Int",
			line:   1,
			column: 7,
		},
		help: "convert with .to_s",
	}

	result := f.FormatError(err, src)

	if !strings.Contains(result, "help:") {
		t.Error("expected 'help:' in output")
	}
	if !strings.Contains(result, "convert with .to_s") {
		t.Error("expected help message in output")
	}
}

func TestFormatter_FormatError_WithEmptyHelp(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = y")
	err := &testErrorWithHelp{
		testError: testError{
			msg:    "undefined: y",
			line:   1,
			column: 5,
		},
		help: "", // empty help should not show "help:" label
	}

	result := f.FormatError(err, src)

	if strings.Contains(result, "help:") {
		t.Error("expected no 'help:' in output for empty help string")
	}
}

func TestFormatter_FormatError_MultiLine(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "def greet(name)\n  puts name\nend")
	err := &testError{
		msg:    "undefined: puts",
		line:   2,
		column: 3,
	}

	result := f.FormatError(err, src)

	// Should show line before, error line, and line after
	if !strings.Contains(result, "def greet(name)") {
		t.Error("expected line before in output")
	}
	if !strings.Contains(result, "puts name") {
		t.Error("expected error line in output")
	}
	if !strings.Contains(result, "end") {
		t.Error("expected line after in output")
	}
}

func TestFormatter_FormatError_FirstLine(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = y\nputs x")
	err := &testError{
		msg:    "undefined: y",
		line:   1,
		column: 5,
	}

	result := f.FormatError(err, src)

	// Should not show line before (doesn't exist)
	if !strings.Contains(result, "x = y") {
		t.Error("expected error line in output")
	}
	if !strings.Contains(result, "puts x") {
		t.Error("expected line after in output")
	}
}

func TestFormatter_FormatError_LastLine(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = 1\ny = x + z")
	err := &testError{
		msg:    "undefined: z",
		line:   2,
		column: 9,
	}

	result := f.FormatError(err, src)

	// Should show line before and error line
	if !strings.Contains(result, "x = 1") {
		t.Error("expected line before in output")
	}
	if !strings.Contains(result, "y = x + z") {
		t.Error("expected error line in output")
	}
}

func TestFormatter_FormatParseError(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "if true\n  puts \"hello\"")

	result := f.FormatParseError("expected 'end'", src, 2, 14, "add 'end' to close the if block")

	if !strings.Contains(result, "error:") {
		t.Error("expected 'error:' in output")
	}
	if !strings.Contains(result, "expected 'end'") {
		t.Error("expected error message in output")
	}
	if !strings.Contains(result, "help:") {
		t.Error("expected 'help:' in output")
	}
}

func TestFormatter_TruncateLine(t *testing.T) {
	f := NewFormatter(ColorNever)
	f.maxLineLen = 20

	tests := []struct {
		input string
		want  string
	}{
		{"short", "short"},
		{"exactly twenty char", "exactly twenty char"},
		{"this line is way too long", "this line is way ..."},
	}

	for _, tt := range tests {
		got := f.truncateLine(tt.input)
		if got != tt.want {
			t.Errorf("truncateLine(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatter_ExtractMessage(t *testing.T) {
	f := NewFormatter(ColorNever)

	tests := []struct {
		input string
		want  string
	}{
		{"line 5: undefined: x", "undefined: x"},
		{"line 10: type mismatch", "type mismatch"},
		{"no prefix here", "no prefix here"},
		{"line: not a line number", "line: not a line number"},
	}

	for _, tt := range tests {
		err := errors.New(tt.input)
		got := f.extractMessage(err)
		if got != tt.want {
			t.Errorf("extractMessage(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSpan_IsSingleLine(t *testing.T) {
	tests := []struct {
		span Span
		want bool
	}{
		{Span{Line: 1, Column: 5}, true},
		{Span{Line: 1, Column: 5, EndLine: 0}, true},
		{Span{Line: 1, Column: 5, EndLine: 1}, true},
		{Span{Line: 1, Column: 5, EndLine: 2}, false},
	}

	for _, tt := range tests {
		got := tt.span.IsSingleLine()
		if got != tt.want {
			t.Errorf("IsSingleLine() for %+v = %v, want %v", tt.span, got, tt.want)
		}
	}
}

func TestFormatter_FormatSemanticError(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = 5 + y")

	result := f.FormatSemanticError("undefined: 'y'", src, 1, 9)

	if !strings.Contains(result, "error:") {
		t.Error("expected 'error:' in output")
	}
	if !strings.Contains(result, "undefined: 'y'") {
		t.Error("expected error message in output")
	}
	if !strings.Contains(result, "test.rg:1:9") {
		t.Error("expected file location in output")
	}
	if !strings.Contains(result, "x = 5 + y") {
		t.Error("expected source line in output")
	}
}

func TestFormatter_FormatError_NilError(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = 5")

	result := f.FormatError(nil, src)

	if result != "" {
		t.Errorf("expected empty string for nil error, got %q", result)
	}
}

func TestFormatter_FormatError_NilSource(t *testing.T) {
	f := NewFormatter(ColorNever)
	err := &testError{
		msg:    "some error",
		line:   1,
		column: 5,
	}

	result := f.FormatError(err, nil)

	// Should still format the error, just without source context
	if !strings.Contains(result, "error:") {
		t.Error("expected 'error:' in output")
	}
	if !strings.Contains(result, "some error") {
		t.Error("expected error message in output")
	}
}

func TestFormatter_FormatError_NoPosition(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = 5")
	// Regular error without Position() method
	err := errors.New("generic error")

	result := f.FormatError(err, src)

	if !strings.Contains(result, "error:") {
		t.Error("expected 'error:' in output")
	}
	if !strings.Contains(result, "generic error") {
		t.Error("expected error message in output")
	}
}

func TestFormatter_Underline_TabHandling(t *testing.T) {
	f := NewFormatter(ColorNever)
	// Source with tab before error location
	src := NewSource("test.rg", "\tx = y")
	err := &testError{
		msg:    "undefined: y",
		line:   1,
		column: 6, // after tab and "x = "
	}

	result := f.FormatError(err, src)

	// Should contain tab in the underline spacing
	if !strings.Contains(result, "\t") {
		t.Error("expected tab character preserved in output")
	}
}

func TestFormatter_Underline_MultiCharToken(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = undefined_variable")
	err := &testError{
		msg:    "undefined: undefined_variable",
		line:   1,
		column: 5,
	}

	result := f.FormatError(err, src)

	// Should have multi-character underline for the identifier
	if !strings.Contains(result, "^~~~") {
		t.Error("expected multi-character underline for identifier")
	}
}

func TestFormatter_Underline_SingleChar(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = +")
	err := &testError{
		msg:    "unexpected operator",
		line:   1,
		column: 5,
	}

	result := f.FormatError(err, src)

	// Should have single caret for non-word character
	if !strings.Contains(result, "^") {
		t.Error("expected caret in output")
	}
}

func TestFormatter_ColorAlways(t *testing.T) {
	f := NewFormatter(ColorAlways)
	src := NewSource("test.rg", "x = y")
	err := &testError{
		msg:    "undefined: y",
		line:   1,
		column: 5,
	}

	result := f.FormatError(err, src)

	// Should contain ANSI color codes
	if !strings.Contains(result, "\033[") {
		t.Error("expected ANSI color codes in output with ColorAlways")
	}
}

func TestSource_NilSource(t *testing.T) {
	var src *Source

	// Should handle nil gracefully
	if line := src.Line(1); line != "" {
		t.Errorf("expected empty string for nil source Line(), got %q", line)
	}
	if count := src.LineCount(); count != 0 {
		t.Errorf("expected 0 for nil source LineCount(), got %d", count)
	}
}

func TestExtractMessage_Public(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"line 5: undefined: x", "undefined: x"},
		{"line 10: type mismatch", "type mismatch"},
		{"no prefix here", "no prefix here"},
	}

	for _, tt := range tests {
		got := ExtractMessage(tt.input)
		if got != tt.want {
			t.Errorf("ExtractMessage(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatter_FormatError_ZeroColumn(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = y")
	err := &testError{
		msg:    "error at start",
		line:   1,
		column: 0, // zero column should be treated as 1
	}

	result := f.FormatError(err, src)

	// Should still format without panic
	if !strings.Contains(result, "error:") {
		t.Error("expected 'error:' in output")
	}
	if !strings.Contains(result, "^") {
		t.Error("expected caret in output")
	}
}

func TestFormatter_FormatError_ColumnBeyondLine(t *testing.T) {
	f := NewFormatter(ColorNever)
	src := NewSource("test.rg", "x = y")
	err := &testError{
		msg:    "error past end",
		line:   1,
		column: 100, // way past end of line
	}

	result := f.FormatError(err, src)

	// Should handle gracefully without panic
	if !strings.Contains(result, "error:") {
		t.Error("expected 'error:' in output")
	}
}
