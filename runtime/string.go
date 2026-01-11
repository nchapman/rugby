package runtime

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Chars splits a string into individual characters (as strings).
// Ruby: str.chars
func Chars(s string) []string {
	runes := []rune(s)
	result := make([]string, len(runes))
	for i, r := range runes {
		result[i] = string(r)
	}
	return result
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
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// StringToInt parses a string as an integer, returning (value, ok).
// Ruby: str.to_i? (optional variant)
func StringToInt(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

// StringToFloat parses a string as a float64, returning (value, ok).
// Ruby: str.to_f? (optional variant)
func StringToFloat(s string) (float64, bool) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// MustAtoi parses a string as an integer, panicking on failure.
// Ruby: str.to_i (assumes valid input)
func MustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic("MustAtoi: invalid integer: " + s)
	}
	return n
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
