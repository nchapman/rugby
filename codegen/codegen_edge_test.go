package codegen

import (
	"testing"
)

func TestOptionalIntCodeGen(t *testing.T) {
	input := `def main
  x : Int? = 5
end`
	output := compile(t, input)

	// We expect runtime.OptionalInt usage
	assertContains(t, output, `runtime.OptionalInt`)
	assertContains(t, output, `runtime.SomeInt(5)`)
}

func TestOptionalStringOrAssign(t *testing.T) {
	input := `def main
  x : String? = nil
  x ||= "default"
end`
	output := compile(t, input)

	assertContains(t, output, `runtime.OptionalString`)
	assertContains(t, output, `if !x.Valid {`)
	assertContains(t, output, `x = runtime.SomeString("default")`)
}

func TestReferenceTypeOrAssign(t *testing.T) {
	input := `def main
  x : Array[Int] = nil
  x ||= [1, 2, 3]
end`
	output := compile(t, input)

	// Array[Int] -> []int
	// x ||= ... -> if x == nil { x = ... }
	assertContains(t, output, `var x []int`) // or similar declaration
	assertContains(t, output, `if x == nil {`)
	assertContains(t, output, `x = []int{1, 2, 3}`)
}

func TestOptionalReturn(t *testing.T) {
	input := `def find(n : Int) -> Int?
  if n < 0
    return nil
  end
  n
end`
	output := compile(t, input)

	assertContains(t, output, `func find(n int) runtime.OptionalInt`)
	assertContains(t, output, `return runtime.NoneInt()`)
	assertContains(t, output, `return runtime.SomeInt(n)`)
}
