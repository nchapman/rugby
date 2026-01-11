package base64

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		input   []byte
		encoded string
	}{
		{[]byte("hello"), "aGVsbG8="},
		{[]byte("hello world"), "aGVsbG8gd29ybGQ="},
		{[]byte(""), ""},
		{[]byte{0x00, 0x01, 0x02}, "AAEC"},
	}

	for _, tt := range tests {
		encoded := Encode(tt.input)
		if encoded != tt.encoded {
			t.Errorf("Encode(%v) = %q, want %q", tt.input, encoded, tt.encoded)
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Errorf("Decode(%q) error = %v", encoded, err)
		}
		if !bytes.Equal(decoded, tt.input) {
			t.Errorf("Decode(%q) = %v, want %v", encoded, decoded, tt.input)
		}
	}
}

func TestEncodeDecodeString(t *testing.T) {
	input := "hello world"
	encoded := EncodeString(input)

	decoded, err := DecodeString(encoded)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	if decoded != input {
		t.Errorf("DecodeString() = %q, want %q", decoded, input)
	}
}

func TestURLEncodeDecode(t *testing.T) {
	// Data that produces + and / in standard base64
	input := []byte{0xfb, 0xff, 0xfe}
	standardEncoded := Encode(input)
	urlEncoded := URLEncode(input)

	// URL encoding should use - instead of + and _ instead of /
	if urlEncoded == standardEncoded {
		t.Error("URLEncode should differ from Encode for this input")
	}

	decoded, err := URLDecode(urlEncoded)
	if err != nil {
		t.Fatalf("URLDecode() error = %v", err)
	}
	if !bytes.Equal(decoded, input) {
		t.Errorf("URLDecode() = %v, want %v", decoded, input)
	}
}

func TestURLEncodeDecodeString(t *testing.T) {
	input := "hello world"
	encoded := URLEncodeString(input)

	decoded, err := URLDecodeString(encoded)
	if err != nil {
		t.Fatalf("URLDecodeString() error = %v", err)
	}
	if decoded != input {
		t.Errorf("URLDecodeString() = %q, want %q", decoded, input)
	}
}

func TestRawEncodeDecode(t *testing.T) {
	input := []byte("hello")
	encoded := RawEncode(input)

	// Raw encoding should not have padding
	if encoded[len(encoded)-1] == '=' {
		t.Error("RawEncode should not have padding")
	}

	decoded, err := RawDecode(encoded)
	if err != nil {
		t.Fatalf("RawDecode() error = %v", err)
	}
	if !bytes.Equal(decoded, input) {
		t.Errorf("RawDecode() = %v, want %v", decoded, input)
	}
}

func TestRawURLEncodeDecode(t *testing.T) {
	input := []byte{0xfb, 0xff, 0xfe}
	encoded := RawURLEncode(input)

	// Should not have padding
	if len(encoded) > 0 && encoded[len(encoded)-1] == '=' {
		t.Error("RawURLEncode should not have padding")
	}

	decoded, err := RawURLDecode(encoded)
	if err != nil {
		t.Fatalf("RawURLDecode() error = %v", err)
	}
	if !bytes.Equal(decoded, input) {
		t.Errorf("RawURLDecode() = %v, want %v", decoded, input)
	}
}

func TestDecodeInvalid(t *testing.T) {
	tests := []string{
		"!!!",
		"not-valid-base64!!!",
		"aGVsbG8", // Missing padding (valid for raw, invalid for standard)
	}

	for _, tt := range tests {
		_, err := Decode(tt)
		if err == nil {
			t.Errorf("Decode(%q) should return error", tt)
		}
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"aGVsbG8=", true},
		{"aGVsbG8gd29ybGQ=", true},
		{"", true},
		{"!!!", false},
		{"aGVsbG8", false}, // Missing padding
	}

	for _, tt := range tests {
		got := Valid(tt.input)
		if got != tt.want {
			t.Errorf("Valid(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestURLValid(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"aGVsbG8=", true},
		{"-_-_", true}, // URL-safe characters
		{"!!!", false},
	}

	for _, tt := range tests {
		got := URLValid(tt.input)
		if got != tt.want {
			t.Errorf("URLValid(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRawEncodeDecodeString(t *testing.T) {
	input := "hello world"
	encoded := RawEncodeString(input)

	decoded, err := RawDecodeString(encoded)
	if err != nil {
		t.Fatalf("RawDecodeString() error = %v", err)
	}
	if decoded != input {
		t.Errorf("RawDecodeString() = %q, want %q", decoded, input)
	}
}

func TestRawURLEncodeDecodeString(t *testing.T) {
	input := "hello world"
	encoded := RawURLEncodeString(input)

	decoded, err := RawURLDecodeString(encoded)
	if err != nil {
		t.Fatalf("RawURLDecodeString() error = %v", err)
	}
	if decoded != input {
		t.Errorf("RawURLDecodeString() = %q, want %q", decoded, input)
	}
}

func TestRawValid(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"aGVsbG8", true},  // "hello" without padding
		{"aGVsbG8=", false}, // With padding (invalid for raw)
		{"!!!", false},
	}

	for _, tt := range tests {
		got := RawValid(tt.input)
		if got != tt.want {
			t.Errorf("RawValid(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRawURLValid(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"aGVsbG8", true}, // "hello" without padding
		{"-_-_", true},    // URL-safe characters
		{"!!!", false},
	}

	for _, tt := range tests {
		got := RawURLValid(tt.input)
		if got != tt.want {
			t.Errorf("RawURLValid(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	// Test various lengths to exercise padding
	for length := 0; length < 10; length++ {
		input := make([]byte, length)
		for i := range input {
			input[i] = byte(i)
		}

		// Standard
		encoded := Encode(input)
		decoded, err := Decode(encoded)
		if err != nil {
			t.Errorf("Decode(Encode(%v)) error = %v", input, err)
		}
		if !bytes.Equal(decoded, input) {
			t.Errorf("roundtrip failed for length %d", length)
		}

		// URL
		urlEncoded := URLEncode(input)
		urlDecoded, err := URLDecode(urlEncoded)
		if err != nil {
			t.Errorf("URLDecode(URLEncode(%v)) error = %v", input, err)
		}
		if !bytes.Equal(urlDecoded, input) {
			t.Errorf("URL roundtrip failed for length %d", length)
		}
	}
}
