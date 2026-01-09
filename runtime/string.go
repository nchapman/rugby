package runtime

import (
	"strconv"
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
