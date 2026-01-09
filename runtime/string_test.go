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
	n, ok := StringToInt("42")
	if !ok || n != 42 {
		t.Errorf("StringToInt valid: got (%d, %v), want (42, true)", n, ok)
	}

	// Negative
	n, ok = StringToInt("-123")
	if !ok || n != -123 {
		t.Errorf("StringToInt negative: got (%d, %v), want (-123, true)", n, ok)
	}

	// Invalid
	_, ok = StringToInt("abc")
	if ok {
		t.Error("StringToInt invalid: got true, want false")
	}

	// Empty
	_, ok = StringToInt("")
	if ok {
		t.Error("StringToInt empty: got true, want false")
	}
}

func TestStringToFloat(t *testing.T) {
	// Valid
	f, ok := StringToFloat("3.14")
	if !ok || f != 3.14 {
		t.Errorf("StringToFloat valid: got (%f, %v), want (3.14, true)", f, ok)
	}

	// Integer form
	f, ok = StringToFloat("42")
	if !ok || f != 42.0 {
		t.Errorf("StringToFloat integer: got (%f, %v), want (42.0, true)", f, ok)
	}

	// Invalid
	_, ok = StringToFloat("abc")
	if ok {
		t.Error("StringToFloat invalid: got true, want false")
	}
}

func TestMustAtoi(t *testing.T) {
	// Valid
	if n := MustAtoi("42"); n != 42 {
		t.Errorf("MustAtoi: got %d, want 42", n)
	}

	// Invalid - should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustAtoi invalid: expected panic")
		}
	}()
	MustAtoi("abc")
}
