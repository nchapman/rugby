package errors

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ColorMode controls when colors are used.
type ColorMode int

const (
	ColorAuto   ColorMode = iota // detect TTY
	ColorAlways                  // always use colors
	ColorNever                   // never use colors
)

// Formatter renders errors with source context and colors.
type Formatter struct {
	colorMode  ColorMode
	useColors  bool
	maxLineLen int // truncate lines longer than this (0 = no limit)
}

// NewFormatter creates a formatter with the given color mode.
func NewFormatter(mode ColorMode) *Formatter {
	f := &Formatter{
		colorMode:  mode,
		maxLineLen: 120,
	}
	f.useColors = f.shouldUseColors()
	return f
}

// shouldUseColors determines if colors should be used.
func (f *Formatter) shouldUseColors() bool {
	switch f.colorMode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default: // ColorAuto
		// Respect NO_COLOR environment variable
		if os.Getenv("NO_COLOR") != "" {
			return false
		}
		// Check if stdout is a terminal
		return term.IsTerminal(int(os.Stdout.Fd()))
	}
}

// ANSI color codes
const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	red       = "\033[31m"
	green     = "\033[32m"
	yellow    = "\033[33m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	white     = "\033[37m"
	boldRed   = "\033[1;31m"
	boldBlue  = "\033[1;34m"
	boldCyan  = "\033[1;36m"
	boldWhite = "\033[1;37m"
)

// color applies color if enabled.
func (f *Formatter) color(code, text string) string {
	if !f.useColors {
		return text
	}
	return code + text + reset
}

// FormatError formats an error with source context.
func (f *Formatter) FormatError(err error, src *Source) string {
	if err == nil {
		return ""
	}

	var sb strings.Builder

	// Get position information if available
	line, col := 0, 0
	if pe, ok := err.(PositionedError); ok {
		line, col = pe.Position()
	}

	// Format error header
	sb.WriteString(f.color(boldRed, "error"))
	sb.WriteString(f.color(bold, ": "))
	sb.WriteString(f.color(boldWhite, f.extractMessage(err)))
	sb.WriteString("\n")

	// Format location
	if src != nil && line > 0 {
		sb.WriteString(f.formatLocation(src.Filename, line, col))
		sb.WriteString("\n")
		sb.WriteString(f.formatSourceContext(src, line, col))
	}

	// Format help if available
	if he, ok := err.(ErrorWithHelp); ok {
		if help := he.Help(); help != "" {
			sb.WriteString(f.formatHelp(help))
		}
	}

	return sb.String()
}

// FormatParseError formats a parser error with source context.
func (f *Formatter) FormatParseError(msg string, src *Source, line, col int, hint string) string {
	var sb strings.Builder

	// Format error header
	sb.WriteString(f.color(boldRed, "error"))
	sb.WriteString(f.color(bold, ": "))
	sb.WriteString(f.color(boldWhite, msg))
	sb.WriteString("\n")

	// Format location
	if src != nil && line > 0 {
		sb.WriteString(f.formatLocation(src.Filename, line, col))
		sb.WriteString("\n")
		sb.WriteString(f.formatSourceContext(src, line, col))
	}

	// Format hint if available
	if hint != "" {
		sb.WriteString(f.formatHelp(hint))
	}

	return sb.String()
}

// FormatSemanticError formats a semantic analysis error with source context.
func (f *Formatter) FormatSemanticError(msg string, src *Source, line, col int) string {
	var sb strings.Builder

	// Format error header
	sb.WriteString(f.color(boldRed, "error"))
	sb.WriteString(f.color(bold, ": "))
	sb.WriteString(f.color(boldWhite, msg))
	sb.WriteString("\n")

	// Format location
	if src != nil && line > 0 {
		sb.WriteString(f.formatLocation(src.Filename, line, col))
		sb.WriteString("\n")
		sb.WriteString(f.formatSourceContext(src, line, col))
	}

	return sb.String()
}

// formatLocation formats the file:line:col location line.
func (f *Formatter) formatLocation(filename string, line, col int) string {
	var sb strings.Builder
	sb.WriteString("  ")
	sb.WriteString(f.color(boldBlue, "-->"))
	sb.WriteString(" ")
	sb.WriteString(f.color(bold, filename))
	sb.WriteString(":")
	sb.WriteString(fmt.Sprintf("%d", line))
	if col > 0 {
		sb.WriteString(fmt.Sprintf(":%d", col))
	}
	return sb.String()
}

// formatSourceContext formats 1-3 lines of source with the error location highlighted.
func (f *Formatter) formatSourceContext(src *Source, errorLine, errorCol int) string {
	var sb strings.Builder

	// Calculate gutter width based on line numbers we'll show
	maxLine := min(errorLine+1, src.LineCount())
	gutterWidth := max(len(fmt.Sprintf("%d", maxLine)), 2)

	// Show context: one line before (if exists), error line, one line after (if exists)
	startLine := max(errorLine-1, 1)
	endLine := min(errorLine+1, src.LineCount())

	// Empty gutter line at start
	sb.WriteString(f.formatGutter("", gutterWidth))
	sb.WriteString("\n")

	for lineNum := startLine; lineNum <= endLine; lineNum++ {
		line := src.Line(lineNum)
		line = f.truncateLine(line)

		// Format the line
		lineNumStr := fmt.Sprintf("%d", lineNum)
		sb.WriteString(f.formatGutter(lineNumStr, gutterWidth))
		sb.WriteString(f.color(boldBlue, " | "))
		sb.WriteString(line)
		sb.WriteString("\n")

		// If this is the error line, add the underline
		if lineNum == errorLine {
			sb.WriteString(f.formatGutter("", gutterWidth))
			sb.WriteString(f.color(boldBlue, " | "))
			sb.WriteString(f.formatUnderline(line, errorCol))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// formatGutter formats a line number in the gutter.
func (f *Formatter) formatGutter(lineNum string, width int) string {
	padded := fmt.Sprintf("%*s", width, lineNum)
	return f.color(dim, padded)
}

// formatUnderline creates the underline (^~~~) pointing to the error column.
func (f *Formatter) formatUnderline(line string, col int) string {
	if col < 1 {
		col = 1
	}

	// Convert to runes to handle Unicode correctly
	runes := []rune(line)

	// Calculate spaces needed (account for tabs in source)
	var spaces strings.Builder
	for i := 0; i < col-1 && i < len(runes); i++ {
		if runes[i] == '\t' {
			spaces.WriteRune('\t')
		} else {
			spaces.WriteRune(' ')
		}
	}

	// Find token length (simple heuristic: word characters or single char)
	tokenLen := 1
	if col-1 < len(runes) {
		start := col - 1
		for i := start; i < len(runes); i++ {
			ch := runes[i]
			if isWordRune(ch) {
				tokenLen = i - start + 1
			} else if i == start {
				tokenLen = 1
				break
			} else {
				break
			}
		}
	}

	// Build underline
	var underline strings.Builder
	underline.WriteRune('^')
	for i := 1; i < tokenLen; i++ {
		underline.WriteRune('~')
	}

	return spaces.String() + f.color(boldRed, underline.String())
}

// formatHelp formats a help message.
func (f *Formatter) formatHelp(help string) string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(f.color(boldCyan, "help"))
	sb.WriteString(f.color(bold, ": "))
	sb.WriteString(help)
	sb.WriteString("\n")
	return sb.String()
}

// truncateLine truncates long lines with ellipsis.
func (f *Formatter) truncateLine(line string) string {
	if f.maxLineLen <= 0 || len(line) <= f.maxLineLen {
		return line
	}
	return line[:f.maxLineLen-3] + "..."
}

// extractMessage extracts the core message from an error, stripping line prefixes.
func (f *Formatter) extractMessage(err error) string {
	return ExtractMessage(err.Error())
}

// ExtractMessage extracts the core message from an error string, stripping "line N: " prefixes.
func ExtractMessage(msg string) string {
	if strings.HasPrefix(msg, "line ") {
		if idx := strings.Index(msg, ": "); idx != -1 {
			return msg[idx+2:]
		}
	}
	return msg
}

// isWordRune returns true if ch is part of an identifier.
func isWordRune(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' || ch == '?' || ch == '!'
}
