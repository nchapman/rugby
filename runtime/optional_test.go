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

func TestOptionUnwrapPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Unwrap on None should panic")
		}
	}()
	none := NoneOpt[int]()
	none.Unwrap()
}

// Pointer-based optional helpers tests

func TestSomeNoneInt(t *testing.T) {
	some := SomeInt(42)
	none := NoneInt()

	if some == nil || *some != 42 {
		t.Error("SomeInt should return pointer to value")
	}
	if none != nil {
		t.Error("NoneInt should return nil")
	}
}

func TestToOptionalInt(t *testing.T) {
	some := ToOptionalInt(42, true)
	none := ToOptionalInt(0, false)

	if some == nil || *some != 42 {
		t.Error("ToOptionalInt with true should return pointer")
	}
	if none != nil {
		t.Error("ToOptionalInt with false should return nil")
	}
}

func TestSomeNoneInt64(t *testing.T) {
	some := SomeInt64(42)
	none := NoneInt64()

	if some == nil || *some != 42 {
		t.Error("SomeInt64 should return pointer to value")
	}
	if none != nil {
		t.Error("NoneInt64 should return nil")
	}
}

func TestToOptionalInt64(t *testing.T) {
	some := ToOptionalInt64(42, true)
	none := ToOptionalInt64(0, false)

	if some == nil || *some != 42 {
		t.Error("ToOptionalInt64 with true should return pointer")
	}
	if none != nil {
		t.Error("ToOptionalInt64 with false should return nil")
	}
}

func TestSomeNoneFloat(t *testing.T) {
	some := SomeFloat(3.14)
	none := NoneFloat()

	if some == nil || *some != 3.14 {
		t.Error("SomeFloat should return pointer to value")
	}
	if none != nil {
		t.Error("NoneFloat should return nil")
	}
}

func TestToOptionalFloat(t *testing.T) {
	some := ToOptionalFloat(3.14, true)
	none := ToOptionalFloat(0, false)

	if some == nil || *some != 3.14 {
		t.Error("ToOptionalFloat with true should return pointer")
	}
	if none != nil {
		t.Error("ToOptionalFloat with false should return nil")
	}
}

func TestSomeNoneString(t *testing.T) {
	some := SomeString("hello")
	none := NoneString()

	if some == nil || *some != "hello" {
		t.Error("SomeString should return pointer to value")
	}
	if none != nil {
		t.Error("NoneString should return nil")
	}
}

func TestToOptionalString(t *testing.T) {
	some := ToOptionalString("hello", true)
	none := ToOptionalString("", false)

	if some == nil || *some != "hello" {
		t.Error("ToOptionalString with true should return pointer")
	}
	if none != nil {
		t.Error("ToOptionalString with false should return nil")
	}
}

func TestSomeNoneBool(t *testing.T) {
	some := SomeBool(true)
	none := NoneBool()

	if some == nil || *some != true {
		t.Error("SomeBool should return pointer to value")
	}
	if none != nil {
		t.Error("NoneBool should return nil")
	}
}

func TestToOptionalBool(t *testing.T) {
	some := ToOptionalBool(true, true)
	none := ToOptionalBool(false, false)

	if some == nil || *some != true {
		t.Error("ToOptionalBool with true should return pointer")
	}
	if none != nil {
		t.Error("ToOptionalBool with false should return nil")
	}
}

// Coalesce function tests

func TestCoalesceInt(t *testing.T) {
	val := 42
	some := &val

	if CoalesceInt(some, 0) != 42 {
		t.Error("CoalesceInt with value should return value")
	}
	if CoalesceInt(nil, 99) != 99 {
		t.Error("CoalesceInt with nil should return default")
	}
}

func TestCoalesceInt64(t *testing.T) {
	val := int64(42)
	some := &val

	if CoalesceInt64(some, 0) != 42 {
		t.Error("CoalesceInt64 with value should return value")
	}
	if CoalesceInt64(nil, 99) != 99 {
		t.Error("CoalesceInt64 with nil should return default")
	}
}

func TestCoalesceFloat(t *testing.T) {
	val := 3.14
	some := &val

	if CoalesceFloat(some, 0) != 3.14 {
		t.Error("CoalesceFloat with value should return value")
	}
	if CoalesceFloat(nil, 2.71) != 2.71 {
		t.Error("CoalesceFloat with nil should return default")
	}
}

func TestCoalesceString(t *testing.T) {
	val := "hello"
	some := &val

	if CoalesceString(some, "") != "hello" {
		t.Error("CoalesceString with value should return value")
	}
	if CoalesceString(nil, "default") != "default" {
		t.Error("CoalesceString with nil should return default")
	}
}

func TestCoalesceBool(t *testing.T) {
	val := true
	some := &val

	if CoalesceBool(some, false) != true {
		t.Error("CoalesceBool with value should return value")
	}
	if CoalesceBool(nil, true) != true {
		t.Error("CoalesceBool with nil should return default")
	}
}
