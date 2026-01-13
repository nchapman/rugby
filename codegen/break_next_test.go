package codegen

import (
	"testing"
)

func TestBreakInEachBlock(t *testing.T) {

	input := `def main

  [1, 2, 3].each do |x|

    if x == 2

      break

    end

    puts(x)

  end

end`

	output := compile(t, input)

	assertContains(t, output, `runtime.Each([]int{1, 2, 3}, func(x any) bool {`)

	assertContains(t, output, `if runtime.Equal(x, 2) {`)

	assertContains(t, output, `return false`)

	assertContains(t, output, `runtime.Puts(x)`)

	assertContains(t, output, `return true`)

}

func TestNextInEachBlock(t *testing.T) {

	input := `def main

  [1, 2, 3].each do |x|

    if x == 2

      next

    end

    puts(x)

  end

end`

	output := compile(t, input)

	assertContains(t, output, `if runtime.Equal(x, 2) {`)

	assertContains(t, output, `return true`)

	assertContains(t, output, `runtime.Puts(x)`)

	assertContains(t, output, `return true`)

}

func TestBreakInMapBlock(t *testing.T) {

	input := `def main

  result = [1, 2, 3].map do |x|

    if x == 2

      break

    end

    x * 10

  end

end`

	output := compile(t, input)

	// Element type is inferred from array literal as int
	assertContains(t, output, `runtime.Map([]int{1, 2, 3}, func(x int) (any, bool, bool) {`)

	assertContains(t, output, `if runtime.Equal(x, 2) {`)

	assertContains(t, output, `return nil, false, false`)

	assertContains(t, output, `return (x * 10), true, true`)

}

func TestNextInMapBlock(t *testing.T) {

	input := `def main

  result = [1, 2, 3].map do |x|

    if x == 2

      next

    end

    x * 10

  end

end`

	output := compile(t, input)

	assertContains(t, output, `if runtime.Equal(x, 2) {`)

	assertContains(t, output, `return nil, false, true`)

	assertContains(t, output, `return (x * 10), true, true`)

}

func TestBreakInSelectBlock(t *testing.T) {

	input := `def main

  result = [1, 2, 3].select do |x|

    if x == 2

      break

    end

    x > 1

  end

end`

	output := compile(t, input)

	// Element type is inferred from array literal as int
	assertContains(t, output, `runtime.Select([]int{1, 2, 3}, func(x int) (bool, bool) {`)

	assertContains(t, output, `if runtime.Equal(x, 2) {`)

	assertContains(t, output, `return false, false`)

	assertContains(t, output, `return (x > 1), true`)

}

func TestNextInMapSkipsElement(t *testing.T) {
	input := `def main
  result = [1, 2, 3].map do |x|
    if x == 2
      next
    end
    x * 10
  end
end`

	output := compile(t, input)
	// next should return (nil, false, true) - don't include, but continue
	assertContains(t, output, `return nil, false, true`)
	assertContains(t, output, `return (x * 10), true, true`)
}

func TestBreakInReduceBlock(t *testing.T) {
	input := `def main
  result = [1, 2, 3].reduce(0) do |acc, x|
    if x == 3
      break
    end
    acc + x
  end
end`

	output := compile(t, input)
	// acc type is inferred from initial value (int), element type from array literal (int)
	assertContains(t, output, `runtime.Reduce([]int{1, 2, 3}, 0, func(acc int, x int) (int, bool) {`)
	assertContains(t, output, `if runtime.Equal(x, 3) {`)
	assertContains(t, output, `return nil, false`)
	assertContains(t, output, `return (acc + x), true`)
}

func TestNextInReduceBlock(t *testing.T) {
	input := `def main
  result = [1, 2, 3].reduce(0) do |acc, x|
    if x == 2
      return acc, true
    end
    acc + x
  end
end`

	output := compile(t, input)
	assertContains(t, output, `return acc, true`)
	assertContains(t, output, `return (acc + x), true`)
}

func TestNestedLoopAndBlockBreak(t *testing.T) {

	input := `def main

  for i in [1, 2]

    [10, 20].each do |j|

      if j == 20

        break

      end

      puts(j)

    end

    if i == 2

      break

    end

  end

end`

	output := compile(t, input)

	// Inner block break should be return false

	assertContains(t, output, `runtime.Each([]int{10, 20}, func(j any) bool {`)

	assertContains(t, output, `if runtime.Equal(j, 20) {`)

	assertContains(t, output, `return false`)

	// Outer loop break should be break

	assertContains(t, output, `if runtime.Equal(i, 2) {`)

	assertContains(t, output, `break`)

}
