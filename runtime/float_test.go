package runtime

import (
	"math"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		n        float64
		expected int
	}{
		{3.7, 3},
		{3.2, 3},
		{-3.7, -3},
		{-3.2, -3},
		{0.0, 0},
		{0.9, 0},
		{-0.9, 0},
		{100.999, 100},
		{-100.999, -100},
	}

	for _, tt := range tests {
		got := Truncate(tt.n)
		if got != tt.expected {
			t.Errorf("Truncate(%v): got %d, want %d", tt.n, got, tt.expected)
		}
	}
}

func TestFloatAbs(t *testing.T) {
	tests := []struct {
		n        float64
		expected float64
	}{
		{5.5, 5.5},
		{-5.5, 5.5},
		{0.0, 0.0},
		{math.MaxFloat64, math.MaxFloat64},
		{-math.MaxFloat64, math.MaxFloat64},
	}

	for _, tt := range tests {
		got := FloatAbs(tt.n)
		if got != tt.expected {
			t.Errorf("FloatAbs(%v): got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestFloatToString(t *testing.T) {
	tests := []struct {
		n        float64
		expected string
	}{
		{3.14, "3.14"},
		{-2.5, "-2.5"},
		{0.0, "0"},
		{100.0, "100"},
		{1.5, "1.5"},
	}

	for _, tt := range tests {
		got := FloatToString(tt.n)
		if got != tt.expected {
			t.Errorf("FloatToString(%v): got %q, want %q", tt.n, got, tt.expected)
		}
	}
}

func TestFloatToInt(t *testing.T) {
	tests := []struct {
		n        float64
		expected int
	}{
		{3.7, 3},
		{3.2, 3},
		{-3.7, -3},
		{-3.2, -3},
		{0.0, 0},
		{0.9, 0},
		{-0.9, 0},
	}

	for _, tt := range tests {
		got := FloatToInt(tt.n)
		if got != tt.expected {
			t.Errorf("FloatToInt(%v): got %d, want %d", tt.n, got, tt.expected)
		}
	}
}

func TestFloatPositive(t *testing.T) {
	tests := []struct {
		n        float64
		expected bool
	}{
		{1.0, true},
		{0.001, true},
		{math.MaxFloat64, true},
		{0.0, false},
		{-1.0, false},
		{-0.001, false},
	}

	for _, tt := range tests {
		got := FloatPositive(tt.n)
		if got != tt.expected {
			t.Errorf("FloatPositive(%v): got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestFloatNegative(t *testing.T) {
	tests := []struct {
		n        float64
		expected bool
	}{
		{-1.0, true},
		{-0.001, true},
		{-math.MaxFloat64, true},
		{0.0, false},
		{1.0, false},
		{0.001, false},
	}

	for _, tt := range tests {
		got := FloatNegative(tt.n)
		if got != tt.expected {
			t.Errorf("FloatNegative(%v): got %v, want %v", tt.n, got, tt.expected)
		}
	}
}

func TestFloatZero(t *testing.T) {
	tests := []struct {
		n        float64
		expected bool
	}{
		{0.0, true},
		{1.0, false},
		{-1.0, false},
		{0.001, false},
		{-0.001, false},
		{math.Inf(1), false},  // positive infinity
		{math.Inf(-1), false}, // negative infinity
	}

	for _, tt := range tests {
		got := FloatZero(tt.n)
		if got != tt.expected {
			t.Errorf("FloatZero(%v): got %v, want %v", tt.n, got, tt.expected)
		}
	}

	// NaN requires special handling - NaN != NaN in IEEE 754
	if FloatZero(math.NaN()) {
		t.Error("FloatZero(NaN): expected false")
	}
}

func TestFloatSpecialValues(t *testing.T) {
	// Test behavior with IEEE 754 special values
	inf := math.Inf(1)
	negInf := math.Inf(-1)
	nan := math.NaN()

	// Infinity absolute values
	if FloatAbs(negInf) != inf {
		t.Errorf("FloatAbs(-Inf): expected +Inf")
	}

	// Positive/Negative predicates with infinity
	if !FloatPositive(inf) {
		t.Error("FloatPositive(+Inf): expected true")
	}
	if !FloatNegative(negInf) {
		t.Error("FloatNegative(-Inf): expected true")
	}

	// NaN predicates - NaN is neither positive, negative, nor zero
	if FloatPositive(nan) {
		t.Error("FloatPositive(NaN): expected false")
	}
	if FloatNegative(nan) {
		t.Error("FloatNegative(NaN): expected false")
	}
}
