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
