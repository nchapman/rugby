package crypto

import (
	"bytes"
	"testing"
)

func TestSHA256(t *testing.T) {
	// Test vectors from NIST
	tests := []struct {
		input string
		want  string
	}{
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{"hello world", "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"},
	}

	for _, tt := range tests {
		got := SHA256([]byte(tt.input))
		if got != tt.want {
			t.Errorf("SHA256(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSHA256String(t *testing.T) {
	got := SHA256String("hello world")
	want := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if got != want {
		t.Errorf("SHA256String() = %q, want %q", got, want)
	}
}

func TestSHA256Bytes(t *testing.T) {
	got := SHA256Bytes([]byte("abc"))
	// First few bytes of SHA256("abc")
	if len(got) != 32 {
		t.Errorf("SHA256Bytes() length = %d, want 32", len(got))
	}
	if got[0] != 0xba || got[1] != 0x78 {
		t.Errorf("SHA256Bytes() first bytes = %x %x, want ba 78", got[0], got[1])
	}
}

func TestSHA512(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
		{"abc", "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"},
	}

	for _, tt := range tests {
		got := SHA512([]byte(tt.input))
		if got != tt.want {
			t.Errorf("SHA512(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSHA512String(t *testing.T) {
	got := SHA512String("abc")
	want := "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"
	if got != want {
		t.Errorf("SHA512String() = %q, want %q", got, want)
	}
}

func TestSHA512Bytes(t *testing.T) {
	got := SHA512Bytes([]byte("abc"))
	if len(got) != 64 {
		t.Errorf("SHA512Bytes() length = %d, want 64", len(got))
	}
}

func TestSHA1(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"abc", "a9993e364706816aba3e25717850c26c9cd0d89d"},
	}

	for _, tt := range tests {
		got := SHA1([]byte(tt.input))
		if got != tt.want {
			t.Errorf("SHA1(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSHA1String(t *testing.T) {
	got := SHA1String("abc")
	want := "a9993e364706816aba3e25717850c26c9cd0d89d"
	if got != want {
		t.Errorf("SHA1String() = %q, want %q", got, want)
	}
}

func TestSHA1Bytes(t *testing.T) {
	got := SHA1Bytes([]byte("abc"))
	if len(got) != 20 {
		t.Errorf("SHA1Bytes() length = %d, want 20", len(got))
	}
}

func TestMD5(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"abc", "900150983cd24fb0d6963f7d28e17f72"},
		{"hello world", "5eb63bbbe01eeed093cb22bb8f5acdc3"},
	}

	for _, tt := range tests {
		got := MD5([]byte(tt.input))
		if got != tt.want {
			t.Errorf("MD5(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMD5String(t *testing.T) {
	got := MD5String("hello world")
	want := "5eb63bbbe01eeed093cb22bb8f5acdc3"
	if got != want {
		t.Errorf("MD5String() = %q, want %q", got, want)
	}
}

func TestMD5Bytes(t *testing.T) {
	got := MD5Bytes([]byte("abc"))
	if len(got) != 16 {
		t.Errorf("MD5Bytes() length = %d, want 16", len(got))
	}
}

func TestRandomBytes(t *testing.T) {
	// Test various lengths
	for _, n := range []int{0, 1, 16, 32, 64, 256} {
		got, err := RandomBytes(n)
		if err != nil {
			t.Fatalf("RandomBytes(%d) error = %v", n, err)
		}
		if len(got) != n {
			t.Errorf("RandomBytes(%d) length = %d, want %d", n, len(got), n)
		}
	}

	// Sanity check: two calls should produce different results
	a, _ := RandomBytes(32)
	b, _ := RandomBytes(32)
	if bytes.Equal(a, b) {
		t.Error("RandomBytes() produced same output twice")
	}
}

func TestRandomBytesNegative(t *testing.T) {
	_, err := RandomBytes(-1)
	if err == nil {
		t.Error("RandomBytes(-1) should return an error")
	}
}

func TestRandomHex(t *testing.T) {
	got, err := RandomHex(16)
	if err != nil {
		t.Fatalf("RandomHex() error = %v", err)
	}
	// 16 bytes = 32 hex characters
	if len(got) != 32 {
		t.Errorf("RandomHex(16) length = %d, want 32", len(got))
	}

	// Should only contain hex characters
	for _, c := range got {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("RandomHex() contains invalid character: %c", c)
		}
	}
}

func TestConstantTimeCompare(t *testing.T) {
	tests := []struct {
		a, b []byte
		want bool
	}{
		{[]byte("abc"), []byte("abc"), true},
		{[]byte("abc"), []byte("abd"), false},
		{[]byte("abc"), []byte("ab"), false},
		{[]byte{}, []byte{}, true},
		{nil, nil, true},
	}

	for _, tt := range tests {
		got := ConstantTimeCompare(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("ConstantTimeCompare(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestConstantTimeCompareString(t *testing.T) {
	if !ConstantTimeCompareString("hello", "hello") {
		t.Error("ConstantTimeCompareString() should return true for equal strings")
	}
	if ConstantTimeCompareString("hello", "world") {
		t.Error("ConstantTimeCompareString() should return false for different strings")
	}
}

func TestHMACSHA256(t *testing.T) {
	// Test vector from RFC 4231
	key := []byte("key")
	data := []byte("The quick brown fox jumps over the lazy dog")
	got := HMACSHA256(key, data)
	want := "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	if got != want {
		t.Errorf("HMACSHA256() = %q, want %q", got, want)
	}
}

func TestHMACSHA256String(t *testing.T) {
	got := HMACSHA256String("key", "The quick brown fox jumps over the lazy dog")
	want := "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	if got != want {
		t.Errorf("HMACSHA256String() = %q, want %q", got, want)
	}
}

func TestHMACSHA256Bytes(t *testing.T) {
	got := HMACSHA256Bytes([]byte("key"), []byte("data"))
	if len(got) != 32 {
		t.Errorf("HMACSHA256Bytes() length = %d, want 32", len(got))
	}
}

func TestHMACSHA512(t *testing.T) {
	key := []byte("key")
	data := []byte("The quick brown fox jumps over the lazy dog")
	got := HMACSHA512(key, data)
	want := "b42af09057bac1e2d41708e48a902e09b5ff7f12ab428a4fe86653c73dd248fb82f948a549f7b791a5b41915ee4d1ec3935357e4e2317250d0372afa2ebeeb3a"
	if got != want {
		t.Errorf("HMACSHA512() = %q, want %q", got, want)
	}
}

func TestHMACSHA512String(t *testing.T) {
	got := HMACSHA512String("key", "The quick brown fox jumps over the lazy dog")
	want := "b42af09057bac1e2d41708e48a902e09b5ff7f12ab428a4fe86653c73dd248fb82f948a549f7b791a5b41915ee4d1ec3935357e4e2317250d0372afa2ebeeb3a"
	if got != want {
		t.Errorf("HMACSHA512String() = %q, want %q", got, want)
	}
}

func TestHMACSHA512Bytes(t *testing.T) {
	got := HMACSHA512Bytes([]byte("key"), []byte("data"))
	if len(got) != 64 {
		t.Errorf("HMACSHA512Bytes() length = %d, want 64", len(got))
	}
}

func TestVerifyHMACSHA256(t *testing.T) {
	key := []byte("secret")
	data := []byte("message")
	signature := HMACSHA256(key, data)

	if !VerifyHMACSHA256(key, data, signature) {
		t.Error("VerifyHMACSHA256() should return true for valid signature")
	}
	if VerifyHMACSHA256(key, data, "invalid") {
		t.Error("VerifyHMACSHA256() should return false for invalid signature")
	}
	if VerifyHMACSHA256([]byte("wrong"), data, signature) {
		t.Error("VerifyHMACSHA256() should return false for wrong key")
	}
}

func TestVerifyHMACSHA512(t *testing.T) {
	key := []byte("secret")
	data := []byte("message")
	signature := HMACSHA512(key, data)

	if !VerifyHMACSHA512(key, data, signature) {
		t.Error("VerifyHMACSHA512() should return true for valid signature")
	}
	if VerifyHMACSHA512(key, data, "invalid") {
		t.Error("VerifyHMACSHA512() should return false for invalid signature")
	}
}

func TestVerifyHMACSHA256Bytes(t *testing.T) {
	key := []byte("secret")
	data := []byte("message")
	signature := HMACSHA256Bytes(key, data)

	if !VerifyHMACSHA256Bytes(key, data, signature) {
		t.Error("VerifyHMACSHA256Bytes() should return true for valid signature")
	}
	if VerifyHMACSHA256Bytes(key, data, []byte("invalid")) {
		t.Error("VerifyHMACSHA256Bytes() should return false for invalid signature")
	}
	if VerifyHMACSHA256Bytes([]byte("wrong"), data, signature) {
		t.Error("VerifyHMACSHA256Bytes() should return false for wrong key")
	}
}

func TestVerifyHMACSHA512Bytes(t *testing.T) {
	key := []byte("secret")
	data := []byte("message")
	signature := HMACSHA512Bytes(key, data)

	if !VerifyHMACSHA512Bytes(key, data, signature) {
		t.Error("VerifyHMACSHA512Bytes() should return true for valid signature")
	}
	if VerifyHMACSHA512Bytes(key, data, []byte("invalid")) {
		t.Error("VerifyHMACSHA512Bytes() should return false for invalid signature")
	}
}
