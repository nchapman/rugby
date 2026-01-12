package runtime

import (
	"reflect"
	"testing"
)

func TestChars(t *testing.T) {
	chars := Chars("hello")
	expected := []string{"h", "e", "l", "l", "o"}
	if !reflect.DeepEqual(chars, expected) {
		t.Errorf("Chars: got %v, want %v", chars, expected)
	}

	// Unicode
	unicodeChars := Chars("日本語")
	expectedUnicode := []string{"日", "本", "語"}
	if !reflect.DeepEqual(unicodeChars, expectedUnicode) {
		t.Errorf("Chars unicode: got %v, want %v", unicodeChars, expectedUnicode)
	}

	// Empty
	empty := Chars("")
	if len(empty) != 0 {
		t.Errorf("Chars empty: got %v, want empty", empty)
	}
}

func TestCharLength(t *testing.T) {
	if n := CharLength("hello"); n != 5 {
		t.Errorf("CharLength ascii: got %d, want 5", n)
	}

	// Unicode - 3 characters but more bytes
	if n := CharLength("日本語"); n != 3 {
		t.Errorf("CharLength unicode: got %d, want 3", n)
	}

	// Emoji
	if n := CharLength("👋🌍"); n != 2 {
		t.Errorf("CharLength emoji: got %d, want 2", n)
	}
}

func TestStringReverse(t *testing.T) {
	if r := StringReverse("hello"); r != "olleh" {
		t.Errorf("StringReverse: got %q, want %q", r, "olleh")
	}

	// Unicode
	if r := StringReverse("日本語"); r != "語本日" {
		t.Errorf("StringReverse unicode: got %q, want %q", r, "語本日")
	}

	// Empty
	if r := StringReverse(""); r != "" {
		t.Errorf("StringReverse empty: got %q, want empty", r)
	}
}

func TestStringToInt(t *testing.T) {
	// Valid
	n, err := StringToInt("42")
	if err != nil || n != 42 {
		t.Errorf("StringToInt valid: got (%d, %v), want (42, nil)", n, err)
	}

	// Negative
	n, err = StringToInt("-123")
	if err != nil || n != -123 {
		t.Errorf("StringToInt negative: got (%d, %v), want (-123, nil)", n, err)
	}

	// Invalid
	_, err = StringToInt("abc")
	if err == nil {
		t.Error("StringToInt invalid: got nil error, want error")
	}

	// Empty
	_, err = StringToInt("")
	if err == nil {
		t.Error("StringToInt empty: got nil error, want error")
	}
}

func TestStringToFloat(t *testing.T) {
	// Valid
	f, err := StringToFloat("3.14")
	if err != nil || f != 3.14 {
		t.Errorf("StringToFloat valid: got (%f, %v), want (3.14, nil)", f, err)
	}

	// Integer form
	f, err = StringToFloat("42")
	if err != nil || f != 42.0 {
		t.Errorf("StringToFloat integer: got (%f, %v), want (42.0, nil)", f, err)
	}

	// Invalid
	_, err = StringToFloat("abc")
	if err == nil {
		t.Error("StringToFloat invalid: got nil error, want error")
	}
}

func TestSplit(t *testing.T) {
	result := Split("a,b,c", ",")
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Split: got %v, want %v", result, expected)
	}

	// No separator found
	result = Split("hello", ",")
	if !reflect.DeepEqual(result, []string{"hello"}) {
		t.Errorf("Split no match: got %v, want [hello]", result)
	}

	// Empty string
	result = Split("", ",")
	if !reflect.DeepEqual(result, []string{""}) {
		t.Errorf("Split empty: got %v, want ['']", result)
	}

	// Multi-char separator
	result = Split("a::b::c", "::")
	if !reflect.DeepEqual(result, []string{"a", "b", "c"}) {
		t.Errorf("Split multi-char: got %v, want [a b c]", result)
	}
}

func TestStrip(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"hello", "hello"},
		{"\t\nhello\n\t", "hello"},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		result := Strip(tt.input)
		if result != tt.expected {
			t.Errorf("Strip(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestLstrip(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello  "},
		{"hello", "hello"},
		{"\t\nhello", "hello"},
		{"", ""},
	}

	for _, tt := range tests {
		result := Lstrip(tt.input)
		if result != tt.expected {
			t.Errorf("Lstrip(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRstrip(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "  hello"},
		{"hello", "hello"},
		{"hello\t\n", "hello"},
		{"", ""},
	}

	for _, tt := range tests {
		result := Rstrip(tt.input)
		if result != tt.expected {
			t.Errorf("Rstrip(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestUpcase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "HELLO"},
		{"Hello World", "HELLO WORLD"},
		{"ALREADY", "ALREADY"},
		{"", ""},
		{"123abc", "123ABC"},
	}

	for _, tt := range tests {
		result := Upcase(tt.input)
		if result != tt.expected {
			t.Errorf("Upcase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDowncase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"already", "already"},
		{"", ""},
		{"123ABC", "123abc"},
	}

	for _, tt := range tests {
		result := Downcase(tt.input)
		if result != tt.expected {
			t.Errorf("Downcase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"HELLO", "Hello"},
		{"hELLO", "Hello"},
		{"", ""},
		{"a", "A"},
		{"1abc", "1abc"},
	}

	for _, tt := range tests {
		result := Capitalize(tt.input)
		if result != tt.expected {
			t.Errorf("Capitalize(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestStringContains(t *testing.T) {
	tests := []struct {
		s, substr string
		expected  bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"hello", "", true},
		{"", "a", false},
		{"", "", true},
	}

	for _, tt := range tests {
		result := StringContains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("StringContains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestReplace(t *testing.T) {
	tests := []struct {
		s, old, new string
		expected    string
	}{
		{"hello world", "world", "there", "hello there"},
		{"aaa", "a", "b", "bbb"},
		{"hello", "x", "y", "hello"},
		{"", "a", "b", ""},
	}

	for _, tt := range tests {
		result := Replace(tt.s, tt.old, tt.new)
		if result != tt.expected {
			t.Errorf("Replace(%q, %q, %q) = %q, want %q", tt.s, tt.old, tt.new, result, tt.expected)
		}
	}
}
