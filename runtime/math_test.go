package runtime

import (
	"math"
	"testing"
)

func TestSqrt(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{4.0, 2.0},
		{9.0, 3.0},
		{2.0, 1.4142135623730951},
		{0.0, 0.0},
		{1.0, 1.0},
	}

	for _, tt := range tests {
		got := Sqrt(tt.input)
		if got != tt.expected {
			t.Errorf("Sqrt(%f) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestSqrtNaN(t *testing.T) {
	result := Sqrt(-1.0)
	if !math.IsNaN(result) {
		t.Errorf("Sqrt(-1.0) = %f, want NaN", result)
	}
}

func TestPow(t *testing.T) {
	tests := []struct {
		base, exp float64
		expected  float64
	}{
		{2.0, 3.0, 8.0},
		{3.0, 2.0, 9.0},
		{10.0, 0.0, 1.0},
		{2.0, -1.0, 0.5},
		{4.0, 0.5, 2.0},
		{0.0, 5.0, 0.0},
		{1.0, 100.0, 1.0},
	}

	for _, tt := range tests {
		got := Pow(tt.base, tt.exp)
		if got != tt.expected {
			t.Errorf("Pow(%f, %f) = %f, want %f", tt.base, tt.exp, got, tt.expected)
		}
	}
}

func TestCeil(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{1.1, 2},
		{1.9, 2},
		{1.0, 1},
		{-1.1, -1},
		{-1.9, -1},
		{0.0, 0},
		{0.5, 1},
		{-0.5, 0},
	}

	for _, tt := range tests {
		got := Ceil(tt.input)
		if got != tt.expected {
			t.Errorf("Ceil(%f) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestCeilPanicsOnInf(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Ceil(+Inf) should panic")
		}
	}()
	Ceil(math.Inf(1))
}

func TestCeilPanicsOnNaN(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Ceil(NaN) should panic")
		}
	}()
	Ceil(math.NaN())
}

func TestFloor(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{1.1, 1},
		{1.9, 1},
		{1.0, 1},
		{-1.1, -2},
		{-1.9, -2},
		{0.0, 0},
		{0.5, 0},
		{-0.5, -1},
	}

	for _, tt := range tests {
		got := Floor(tt.input)
		if got != tt.expected {
			t.Errorf("Floor(%f) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestFloorPanicsOnInf(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Floor(-Inf) should panic")
		}
	}()
	Floor(math.Inf(-1))
}

func TestRound(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{1.4, 1},
		{1.5, 2},
		{1.6, 2},
		{2.5, 3},
		{-1.4, -1},
		{-1.5, -2},
		{-1.6, -2},
		{0.0, 0},
	}

	for _, tt := range tests {
		got := Round(tt.input)
		if got != tt.expected {
			t.Errorf("Round(%f) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestRoundPanicsOnNaN(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Round(NaN) should panic")
		}
	}()
	Round(math.NaN())
}
