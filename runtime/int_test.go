package runtime

import (
	"math"
	"testing"
)

func TestAbs(t *testing.T) {
	if a := Abs(5); a != 5 {
		t.Errorf("Abs positive: got %d, want 5", a)
	}

	if a := Abs(-5); a != 5 {
		t.Errorf("Abs negative: got %d, want 5", a)
	}

	if a := Abs(0); a != 0 {
		t.Errorf("Abs zero: got %d, want 0", a)
	}
}

func TestClamp(t *testing.T) {
	// Within range
	if c := Clamp(5, 0, 10); c != 5 {
		t.Errorf("Clamp within: got %d, want 5", c)
	}

	// Below min
	if c := Clamp(-5, 0, 10); c != 0 {
		t.Errorf("Clamp below: got %d, want 0", c)
	}

	// Above max
	if c := Clamp(15, 0, 10); c != 10 {
		t.Errorf("Clamp above: got %d, want 10", c)
	}

	// At boundary
	if c := Clamp(0, 0, 10); c != 0 {
		t.Errorf("Clamp at min: got %d, want 0", c)
	}

	if c := Clamp(10, 0, 10); c != 10 {
		t.Errorf("Clamp at max: got %d, want 10", c)
	}
}

func TestAbsInt64(t *testing.T) {
	if a := AbsInt64(-100); a != 100 {
		t.Errorf("AbsInt64: got %d, want 100", a)
	}
}

func TestClampInt64(t *testing.T) {
	if c := ClampInt64(150, 0, 100); c != 100 {
		t.Errorf("ClampInt64: got %d, want 100", c)
	}
}

func TestTimes(t *testing.T) {
	var collected []int
	Times(5, func(i int) {
		collected = append(collected, i)
	})
	expected := []int{0, 1, 2, 3, 4}
	if len(collected) != 5 {
		t.Errorf("Times: got %v, want %v", collected, expected)
	}
	for i, v := range collected {
		if v != expected[i] {
			t.Errorf("Times[%d]: got %d, want %d", i, v, expected[i])
		}
	}

	// Times(0) should not iterate
	var count int
	Times(0, func(i int) {
		count++
	})
	if count != 0 {
		t.Errorf("Times(0): got %d iterations, want 0", count)
	}
}

func TestUpto(t *testing.T) {
	var collected []int
	Upto(1, 5, func(i int) {
		collected = append(collected, i)
	})
	expected := []int{1, 2, 3, 4, 5}
	if len(collected) != 5 {
		t.Errorf("Upto: got %v, want %v", collected, expected)
	}

	// Empty range
	var count int
	Upto(5, 1, func(i int) {
		count++
	})
	if count != 0 {
		t.Errorf("Upto empty: got %d iterations, want 0", count)
	}
}

func TestDownto(t *testing.T) {
	var collected []int
	Downto(5, 1, func(i int) {
		collected = append(collected, i)
	})
	expected := []int{5, 4, 3, 2, 1}
	if len(collected) != 5 {
		t.Errorf("Downto: got %v, want %v", collected, expected)
	}

	// Empty range
	var count int
	Downto(1, 5, func(i int) {
		count++
	})
	if count != 0 {
		t.Errorf("Downto empty: got %d iterations, want 0", count)
	}
}

func TestEven(t *testing.T) {
	tests := []struct {
		n        int
		expected bool
	}{
		{0, true},
		{2, true},
		{-2, true},
		{4, true},
		{1, false},
		{-1, false},
		{3, false},
	}

	for _, tt := range tests {
		got := Even(tt.n)
		if got != tt.expected {
			t.Errorf("Even(%d): got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestOdd(t *testing.T) {
	tests := []struct {
		n        int
		expected bool
	}{
		{1, true},
		{-1, true},
		{3, true},
		{0, false},
		{2, false},
		{-2, false},
		{4, false},
	}

	for _, tt := range tests {
		got := Odd(tt.n)
		if got != tt.expected {
			t.Errorf("Odd(%d): got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestPositive(t *testing.T) {
	tests := []struct {
		n        int
		expected bool
	}{
		{1, true},
		{100, true},
		{0, false},
		{-1, false},
		{-100, false},
	}

	for _, tt := range tests {
		got := Positive(tt.n)
		if got != tt.expected {
			t.Errorf("Positive(%d): got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestNegative(t *testing.T) {
	tests := []struct {
		n        int
		expected bool
	}{
		{-1, true},
		{-100, true},
		{0, false},
		{1, false},
		{100, false},
	}

	for _, tt := range tests {
		got := Negative(tt.n)
		if got != tt.expected {
			t.Errorf("Negative(%d): got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestZero(t *testing.T) {
	tests := []struct {
		n        int
		expected bool
	}{
		{0, true},
		{1, false},
		{-1, false},
		{100, false},
		{-100, false},
	}

	for _, tt := range tests {
		got := Zero(tt.n)
		if got != tt.expected {
			t.Errorf("Zero(%d): got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{0, "0"},
		{42, "42"},
		{-17, "-17"},
		{1000000, "1000000"},
	}

	for _, tt := range tests {
		got := IntToString(tt.n)
		if got != tt.expected {
			t.Errorf("IntToString(%d): got %q, want %q", tt.n, got, tt.expected)
		}
	}
}

func TestIntToFloat(t *testing.T) {
	tests := []struct {
		n        int
		expected float64
	}{
		{0, 0.0},
		{42, 42.0},
		{-17, -17.0},
	}

	for _, tt := range tests {
		got := IntToFloat(tt.n)
		if got != tt.expected {
			t.Errorf("IntToFloat(%d): got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestAbsInt64_Extended(t *testing.T) {
	// Additional coverage for edge cases
	if AbsInt64(-1) != 1 {
		t.Error("AbsInt64 -1 failed")
	}
}

func TestAbsOverflow(t *testing.T) {
	// Document behavior: Abs(math.MinInt) overflows and returns math.MinInt
	// This is noted in the function doc: "Note: Abs(math.MinInt) overflows and returns math.MinInt"
	// We test to ensure this documented behavior doesn't change unexpectedly
	result := Abs(math.MinInt)
	if result != math.MinInt {
		t.Errorf("Abs(math.MinInt): expected math.MinInt (overflow), got %d", result)
	}

	// Same for int64 version
	result64 := AbsInt64(math.MinInt64)
	if result64 != math.MinInt64 {
		t.Errorf("AbsInt64(math.MinInt64): expected math.MinInt64 (overflow), got %d", result64)
	}
}

func TestClampInt64_Extended(t *testing.T) {
	// Additional edge cases
	if ClampInt64(0, 0, 10) != 0 {
		t.Error("ClampInt64 at min boundary failed")
	}
	if ClampInt64(10, 0, 10) != 10 {
		t.Error("ClampInt64 at max boundary failed")
	}
}
