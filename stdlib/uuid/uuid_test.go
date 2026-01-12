package uuid

import (
	"strings"
	"testing"
	"time"
)

func TestV4(t *testing.T) {
	u, err := V4()
	if err != nil {
		t.Fatalf("V4() error = %v", err)
	}

	// Check format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(u) != 36 {
		t.Errorf("V4() length = %d, want 36", len(u))
	}

	// Check hyphens at correct positions
	if u[8] != '-' || u[13] != '-' || u[18] != '-' || u[23] != '-' {
		t.Errorf("V4() has incorrect hyphen positions: %s", u)
	}

	// Check version is 4
	if u[14] != '4' {
		t.Errorf("V4() version byte = %c, want '4'", u[14])
	}

	// Check variant is 8, 9, a, or b
	variant := u[19]
	if variant != '8' && variant != '9' && variant != 'a' && variant != 'b' {
		t.Errorf("V4() variant byte = %c, want 8/9/a/b", variant)
	}

	// Check lowercase
	if u != strings.ToLower(u) {
		t.Errorf("V4() not lowercase: %s", u)
	}
}

func TestV4Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for range 1000 {
		u, err := V4()
		if err != nil {
			t.Fatalf("V4() error = %v", err)
		}
		if seen[u] {
			t.Errorf("V4() generated duplicate: %s", u)
		}
		seen[u] = true
	}
}

func TestV7(t *testing.T) {
	before := time.Now().UnixMilli()
	u, err := V7()
	after := time.Now().UnixMilli()

	if err != nil {
		t.Fatalf("V7() error = %v", err)
	}

	// Check format
	if len(u) != 36 {
		t.Errorf("V7() length = %d, want 36", len(u))
	}

	// Check version is 7
	if u[14] != '7' {
		t.Errorf("V7() version byte = %c, want '7'", u[14])
	}

	// Check timestamp is reasonable
	ts := Timestamp(u)
	if ts < before || ts > after {
		t.Errorf("V7() timestamp = %d, want between %d and %d", ts, before, after)
	}
}

func TestV7Ordering(t *testing.T) {
	u1, _ := V7()
	time.Sleep(2 * time.Millisecond)
	u2, _ := V7()

	// V7 UUIDs should be ordered by creation time
	if Compare(u1, u2) != -1 {
		t.Errorf("V7() not properly ordered: %s should be < %s", u1, u2)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "standard lowercase",
			input: "550e8400-e29b-41d4-a716-446655440000",
			want:  "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:  "uppercase",
			input: "550E8400-E29B-41D4-A716-446655440000",
			want:  "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:  "no hyphens",
			input: "550e8400e29b41d4a716446655440000",
			want:  "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:  "mixed case no hyphens",
			input: "550E8400E29B41D4A716446655440000",
			want:  "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:    "too short",
			input:   "550e8400-e29b-41d4",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "550e8400-e29b-41d4-a716-446655440000extra",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			input:   "550e8400-e29b-41d4-a716-44665544000g",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"550E8400-E29B-41D4-A716-446655440000", true},
		{"550e8400e29b41d4a716446655440000", true},
		{"invalid", false},
		{"", false},
		{"550e8400-e29b-41d4-a716-44665544000g", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := Valid(tt.input); got != tt.want {
				t.Errorf("Valid(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNil(t *testing.T) {
	n := Nil()
	if n != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("Nil() = %s, want all zeros", n)
	}
}

func TestIsNil(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"00000000-0000-0000-0000-000000000000", true},
		{"00000000000000000000000000000000", true},
		{"550e8400-e29b-41d4-a716-446655440000", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsNil(tt.input); got != tt.want {
				t.Errorf("IsNil(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"550e8400-e29b-41d4-a716-446655440000", 4},
		{"550e8400-e29b-71d4-a716-446655440000", 7},
		{"550e8400-e29b-11d4-a716-446655440000", 1},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := Version(tt.input); got != tt.want {
				t.Errorf("Version(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}

	// Generated UUIDs should have correct versions
	v4, _ := V4()
	if Version(v4) != 4 {
		t.Errorf("V4() version = %d, want 4", Version(v4))
	}

	v7, _ := V7()
	if Version(v7) != 7 {
		t.Errorf("V7() version = %d, want 7", Version(v7))
	}
}

func TestBytes(t *testing.T) {
	input := "550e8400-e29b-41d4-a716-446655440000"
	b, err := Bytes(input)
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}
	if len(b) != 16 {
		t.Errorf("Bytes() length = %d, want 16", len(b))
	}

	// Check first byte
	if b[0] != 0x55 {
		t.Errorf("Bytes()[0] = %x, want 0x55", b[0])
	}
}

func TestFromBytes(t *testing.T) {
	original := "550e8400-e29b-41d4-a716-446655440000"
	b, _ := Bytes(original)

	result, err := FromBytes(b)
	if err != nil {
		t.Fatalf("FromBytes() error = %v", err)
	}
	if result != original {
		t.Errorf("FromBytes() = %s, want %s", result, original)
	}

	// Test invalid length
	_, err = FromBytes([]byte{1, 2, 3})
	if err == nil {
		t.Error("FromBytes() expected error for invalid length")
	}
}

func TestTimestamp(t *testing.T) {
	v7, _ := V7()
	ts := Timestamp(v7)
	now := time.Now().UnixMilli()

	// Should be within 1 second
	if ts < now-1000 || ts > now+1000 {
		t.Errorf("Timestamp() = %d, expected close to %d", ts, now)
	}

	// V4 should return 0
	v4, _ := V4()
	if Timestamp(v4) != 0 {
		t.Errorf("Timestamp(v4) = %d, want 0", Timestamp(v4))
	}
}

func TestTime(t *testing.T) {
	before := time.Now()
	v7, _ := V7()
	after := time.Now()

	tm := Time(v7)
	if tm.Before(before.Add(-time.Millisecond)) || tm.After(after.Add(time.Millisecond)) {
		t.Errorf("Time() = %v, expected between %v and %v", tm, before, after)
	}
}

func TestCompare(t *testing.T) {
	a := "00000000-0000-0000-0000-000000000001"
	b := "00000000-0000-0000-0000-000000000002"

	if Compare(a, b) != -1 {
		t.Errorf("Compare(%s, %s) != -1", a, b)
	}
	if Compare(b, a) != 1 {
		t.Errorf("Compare(%s, %s) != 1", b, a)
	}
	if Compare(a, a) != 0 {
		t.Errorf("Compare(%s, %s) != 0", a, a)
	}
	if Compare("invalid", a) != 0 {
		t.Error("Compare(invalid, valid) != 0")
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000", true},
		{"550e8400-e29b-41d4-a716-446655440000", "550E8400-E29B-41D4-A716-446655440000", true},
		{"550e8400e29b41d4a716446655440000", "550e8400-e29b-41d4-a716-446655440000", true},
		{"550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440001", false},
		{"invalid", "550e8400-e29b-41d4-a716-446655440000", false},
	}

	for _, tt := range tests {
		if got := Equal(tt.a, tt.b); got != tt.want {
			t.Errorf("Equal(%s, %s) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestDeterministic(t *testing.T) {
	// Same inputs should produce same output
	u1 := Deterministic("namespace", "name")
	u2 := Deterministic("namespace", "name")
	if u1 != u2 {
		t.Errorf("Deterministic() not deterministic: %s != %s", u1, u2)
	}

	// Different inputs should produce different output
	u3 := Deterministic("namespace", "other")
	if u1 == u3 {
		t.Errorf("Deterministic() produced same UUID for different inputs")
	}

	// Output should be valid UUID
	if !Valid(u1) {
		t.Errorf("Deterministic() produced invalid UUID: %s", u1)
	}
}

func TestMustV4(t *testing.T) {
	// Should not panic under normal conditions
	u := MustV4()
	if !Valid(u) {
		t.Errorf("MustV4() produced invalid UUID: %s", u)
	}
}

func TestMustV7(t *testing.T) {
	// Should not panic under normal conditions
	u := MustV7()
	if !Valid(u) {
		t.Errorf("MustV7() produced invalid UUID: %s", u)
	}
}
