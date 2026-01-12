package runtime

import "testing"

func TestOption(t *testing.T) {
	// Test Some/None construction
	some := SomeOpt(42)
	none := NoneOpt[int]()

	if !some.IsPresent() {
		t.Error("Some value should be present")
	}
	if some.IsAbsent() {
		t.Error("Some value should not be absent")
	}
	if none.IsPresent() {
		t.Error("None should not be present")
	}
	if !none.IsAbsent() {
		t.Error("None should be absent")
	}
}

func TestOptionUnwrap(t *testing.T) {
	some := SomeOpt("hello")
	if some.Unwrap() != "hello" {
		t.Error("Unwrap should return the value")
	}
}

func TestOptionUnwrapOr(t *testing.T) {
	some := SomeOpt(10)
	none := NoneOpt[int]()

	if some.UnwrapOr(0) != 10 {
		t.Error("UnwrapOr on Some should return the value")
	}
	if none.UnwrapOr(99) != 99 {
		t.Error("UnwrapOr on None should return the default")
	}
}

func TestOptionToTuple(t *testing.T) {
	some := SomeOpt(5)
	none := NoneOpt[int]()

	v1, ok1 := some.ToTuple()
	if v1 != 5 || !ok1 {
		t.Error("ToTuple on Some should return (value, true)")
	}

	v2, ok2 := none.ToTuple()
	if v2 != 0 || ok2 {
		t.Error("ToTuple on None should return (zero, false)")
	}
}

func TestFromTuple(t *testing.T) {
	opt1 := FromTuple(42, true)
	opt2 := FromTuple(0, false)

	if !opt1.IsPresent() || opt1.Value != 42 {
		t.Error("FromTuple with true should create Some")
	}
	if opt2.IsPresent() {
		t.Error("FromTuple with false should create None")
	}
}
