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
