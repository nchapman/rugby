package runtime

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestRandInt(t *testing.T) {
	// Test bounds
	for i := 0; i < 100; i++ {
		r := RandInt(10)
		if r < 0 || r >= 10 {
			t.Errorf("RandInt(10) = %d, want [0, 10)", r)
		}
	}

	// Edge case: 0 or negative
	if r := RandInt(0); r != 0 {
		t.Errorf("RandInt(0) = %d, want 0", r)
	}
	if r := RandInt(-5); r != 0 {
		t.Errorf("RandInt(-5) = %d, want 0", r)
	}

	// RandInt(1) should always return 0
	for i := 0; i < 10; i++ {
		if r := RandInt(1); r != 0 {
			t.Errorf("RandInt(1) = %d, want 0", r)
		}
	}
}

func TestRandFloat(t *testing.T) {
	for i := 0; i < 100; i++ {
		r := RandFloat()
		if r < 0.0 || r >= 1.0 {
			t.Errorf("RandFloat() = %f, want [0.0, 1.0)", r)
		}
	}
}

func TestRandRange(t *testing.T) {
	// Normal range
	for i := 0; i < 100; i++ {
		r := RandRange(5, 10)
		if r < 5 || r > 10 {
			t.Errorf("RandRange(5, 10) = %d, want [5, 10]", r)
		}
	}

	// Inverted range (should still work)
	for i := 0; i < 100; i++ {
		r := RandRange(10, 5)
		if r < 5 || r > 10 {
			t.Errorf("RandRange(10, 5) = %d, want [5, 10]", r)
		}
	}

	// Single value range
	for i := 0; i < 10; i++ {
		if r := RandRange(7, 7); r != 7 {
			t.Errorf("RandRange(7, 7) = %d, want 7", r)
		}
	}
}

func TestSleep(t *testing.T) {
	// Just verify it doesn't panic - actual timing is hard to test
	Sleep(0.001) // 1ms
	SleepMs(1)
}

// captureStdout captures stdout during fn execution
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPuts(t *testing.T) {
	out := captureStdout(func() {
		Puts("hello")
	})
	if out != "hello\n" {
		t.Errorf("Puts: got %q, want %q", out, "hello\n")
	}

	// Multiple args - each gets its own line
	out = captureStdout(func() {
		Puts("a", "b", "c")
	})
	if out != "a\nb\nc\n" {
		t.Errorf("Puts multiple: got %q, want %q", out, "a\nb\nc\n")
	}
}

func TestPrint(t *testing.T) {
	out := captureStdout(func() {
		Print("hello")
	})
	if out != "hello" {
		t.Errorf("Print: got %q, want %q", out, "hello")
	}

	// Multiple args - no separators
	out = captureStdout(func() {
		Print("a", "b", "c")
	})
	if out != "abc" {
		t.Errorf("Print multiple: got %q, want %q", out, "abc")
	}
}

func TestP(t *testing.T) {
	out := captureStdout(func() {
		P("hello", 42, true)
	})
	expected := "\"hello\" 42 true\n"
	if out != expected {
		t.Errorf("P: got %q, want %q", out, expected)
	}
}

// Exit is not tested as it terminates the program
// Gets is not tested as it requires stdin mocking
