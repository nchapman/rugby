package runtime

import "testing"

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
	Times(5, func(i int) bool {
		collected = append(collected, i)
		return true
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
	Times(0, func(i int) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("Times(0): got %d iterations, want 0", count)
	}

	// Times with early break
	var partial []int
	Times(10, func(i int) bool {
		partial = append(partial, i)
		return i < 2
	})
	if len(partial) != 3 {
		t.Errorf("Times with break: got %v, want [0 1 2]", partial)
	}
}

func TestUpto(t *testing.T) {
	var collected []int
	Upto(1, 5, func(i int) bool {
		collected = append(collected, i)
		return true
	})
	expected := []int{1, 2, 3, 4, 5}
	if len(collected) != 5 {
		t.Errorf("Upto: got %v, want %v", collected, expected)
	}

	// Empty range
	var count int
	Upto(5, 1, func(i int) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("Upto empty: got %d iterations, want 0", count)
	}

	// With break
	var partial []int
	Upto(1, 10, func(i int) bool {
		partial = append(partial, i)
		return i < 3
	})
	if len(partial) != 3 {
		t.Errorf("Upto with break: got %v, want [1 2 3]", partial)
	}
}

func TestDownto(t *testing.T) {
	var collected []int
	Downto(5, 1, func(i int) bool {
		collected = append(collected, i)
		return true
	})
	expected := []int{5, 4, 3, 2, 1}
	if len(collected) != 5 {
		t.Errorf("Downto: got %v, want %v", collected, expected)
	}

	// Empty range
	var count int
	Downto(1, 5, func(i int) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("Downto empty: got %d iterations, want 0", count)
	}

	// With break
	var partial []int
	Downto(10, 1, func(i int) bool {
		partial = append(partial, i)
		return i > 8
	})
	if len(partial) != 3 {
		t.Errorf("Downto with break: got %v, want [10 9 8]", partial)
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
