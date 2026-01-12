package runtime

import (
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Chars splits a string into individual characters (as strings).
// Ruby: str.chars
// Note: strings.Split with empty separator splits on UTF-8 character boundaries.
func Chars(s string) []string {
	return strings.Split(s, "")
}

// CharLength returns the number of characters (runes) in the string.
// Ruby: str.length (for unicode-aware length)
func CharLength(s string) int {
	return utf8.RuneCountInString(s)
}

// StringReverse reverses a string at the rune level.
// Ruby: str.reverse
// Note: May produce unexpected results for combining characters or
// multi-rune grapheme clusters (e.g., some emoji with modifiers).
func StringReverse(s string) string {
	runes := []rune(s)
	slices.Reverse(runes)
	return string(runes)
}

// StringToInt parses a string as an integer, returning (value, error).
// Ruby: str.to_i() - use with ! for propagation or rescue for default
func StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// StringToFloat parses a string as a float64, returning (value, error).
// Ruby: str.to_f() - use with ! for propagation or rescue for default
func StringToFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// Split splits a string by a separator.
// Ruby: str.split(sep)
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// Strip removes leading and trailing whitespace.
// Ruby: str.strip
func Strip(s string) string {
	return strings.TrimSpace(s)
}

// Lstrip removes leading whitespace.
// Ruby: str.lstrip
func Lstrip(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}

// Rstrip removes trailing whitespace.
// Ruby: str.rstrip
func Rstrip(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// Upcase returns the string in uppercase.
// Ruby: str.upcase
func Upcase(s string) string {
	return strings.ToUpper(s)
}

// Downcase returns the string in lowercase.
// Ruby: str.downcase
func Downcase(s string) string {
	return strings.ToLower(s)
}

// Capitalize uppercases the first character and lowercases the rest.
// Ruby: str.capitalize
func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + strings.ToLower(s[size:])
}

// StringContains returns true if the string contains the substring.
// Ruby: str.include?(substr)
func StringContains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Replace replaces all occurrences of old with new.
// Ruby: str.gsub(old, new)
func Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}
