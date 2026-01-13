package codegen

import (
	"strings"
	"testing"

	"github.com/nchapman/rugby/lexer"
	"github.com/nchapman/rugby/parser"
)

func TestGoStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input: `go do
	puts("hello")
end`,
			expected: `go func() {`,
			desc:     "go with do block",
		},
		{
			input:    `go fetch_url(url)`,
			expected: `go fetchURL(url)`,
			desc:     "go with call expression",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
			continue
		}

		g := New()
		code, err := g.Generate(program)
		if err != nil {
			t.Errorf("%s: codegen error: %v", tt.desc, err)
			continue
		}

		if !strings.Contains(code, tt.expected) {
			t.Errorf("%s: expected code to contain '%s', got:\n%s", tt.desc, tt.expected, code)
		}
	}
}

func TestSpawnExpression(t *testing.T) {
	input := `t = spawn { 42 }`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	expected := "runtime.Spawn"
	if !strings.Contains(code, expected) {
		t.Errorf("expected code to contain '%s', got:\n%s", expected, code)
	}
}

func TestAwaitExpression(t *testing.T) {
	input := `result = await t`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	expected := "runtime.Await"
	if !strings.Contains(code, expected) {
		t.Errorf("expected code to contain '%s', got:\n%s", expected, code)
	}
}

func TestSelectStatement(t *testing.T) {
	input := `select
when val = ch1.receive
	puts(val)
when ch2 << 42
	puts("sent")
else
	puts("default")
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	// Should generate select { case ... }
	if !strings.Contains(code, "select {") {
		t.Errorf("expected 'select {', got:\n%s", code)
	}
	// Should have case with receive
	if !strings.Contains(code, "case val := <-") {
		t.Errorf("expected 'case val := <-', got:\n%s", code)
	}
	// Should have case with send
	if !strings.Contains(code, "ch2 <- 42") {
		t.Errorf("expected 'ch2 <- 42', got:\n%s", code)
	}
	// Should have default case
	if !strings.Contains(code, "default:") {
		t.Errorf("expected 'default:', got:\n%s", code)
	}
}

func TestConcurrentlyStatement(t *testing.T) {
	input := `concurrently do |scope|
	a = scope.spawn { 1 }
	x = await a
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	// Should generate NewScope
	if !strings.Contains(code, "runtime.NewScope") {
		t.Errorf("expected 'runtime.NewScope', got:\n%s", code)
	}
	// Should generate defer scope.Wait()
	if !strings.Contains(code, "defer scope.Wait()") {
		t.Errorf("expected 'defer scope.Wait()', got:\n%s", code)
	}
}

func TestConcurrentlyExpression(t *testing.T) {
	// Test concurrently as an expression (used in assignment)
	input := `result = concurrently do |scope|
	a = scope.spawn { 1 }
	await a
	42
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	// Should generate assignment with IIFE
	if !strings.Contains(code, "result := func() any") {
		t.Errorf("expected 'result := func() any', got:\n%s", code)
	}
	// Should generate NewScope
	if !strings.Contains(code, "runtime.NewScope") {
		t.Errorf("expected 'runtime.NewScope', got:\n%s", code)
	}
	// Should generate defer scope.Wait()
	if !strings.Contains(code, "defer scope.Wait()") {
		t.Errorf("expected 'defer scope.Wait()', got:\n%s", code)
	}
	// Should return the last expression
	if !strings.Contains(code, "return 42") {
		t.Errorf("expected 'return 42', got:\n%s", code)
	}
}

func TestChannelCreation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    `ch = Chan[Int].new(10)`,
			expected: "make(chan int, 10)",
			desc:     "buffered channel",
		},
		{
			input:    `ch = Chan[String].new`,
			expected: "make(chan string)",
			desc:     "unbuffered channel",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
			continue
		}

		g := New()
		code, err := g.Generate(program)
		if err != nil {
			t.Errorf("%s: codegen error: %v", tt.desc, err)
			continue
		}

		if !strings.Contains(code, tt.expected) {
			t.Errorf("%s: expected '%s', got:\n%s", tt.desc, tt.expected, code)
		}
	}
}

func TestChannelOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    `ch << 42`,
			expected: "runtime.ShiftLeft(ch, 42)",
			desc:     "channel send",
		},
		{
			input:    `val = ch.receive`,
			expected: "<-ch",
			desc:     "channel receive",
		},
		{
			input:    `val = ch.try_receive`,
			expected: "runtime.TryReceive(ch)",
			desc:     "non-blocking receive",
		},
		{
			input:    `ch.close`,
			expected: "close(ch)",
			desc:     "channel close",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
			continue
		}

		g := New()
		code, err := g.Generate(program)
		if err != nil {
			t.Errorf("%s: codegen error: %v", tt.desc, err)
			continue
		}

		if !strings.Contains(code, tt.expected) {
			t.Errorf("%s: expected '%s', got:\n%s", tt.desc, tt.expected, code)
		}
	}
}

func TestChannelIteration(t *testing.T) {
	input := `for msg in ch
	puts(msg)
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	// Channel iteration uses for range
	if !strings.Contains(code, "for _, msg := range ch") {
		t.Errorf("expected 'for _, msg := range ch', got:\n%s", code)
	}
}

// Additional comprehensive codegen tests

func TestScopedSpawnBlock(t *testing.T) {
	input := `concurrently do |scope|
	t = scope.spawn { compute() }
	x = await t
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	// Should generate scope.Spawn(func() any { ... })
	if !strings.Contains(code, "scope.Spawn(func() any {") {
		t.Errorf("expected 'scope.Spawn(func() any {', got:\n%s", code)
	}
}

func TestTaskAwaitMethod(t *testing.T) {
	input := `result = task.await`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	// Should generate runtime.Await(task)
	if !strings.Contains(code, "runtime.Await(task)") {
		t.Errorf("expected 'runtime.Await(task)', got:\n%s", code)
	}
}

func TestSpawnWithDoBlock(t *testing.T) {
	input := `t = spawn do
	x = setup()
	process(x)
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	if !strings.Contains(code, "runtime.Spawn(func() any {") {
		t.Errorf("expected 'runtime.Spawn(func() any {', got:\n%s", code)
	}
}

func TestChannelTypeMapping(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{`ch = Chan[Int].new`, "make(chan int)", "Int to int"},
		{`ch = Chan[Int64].new`, "make(chan int64)", "Int64 to int64"},
		{`ch = Chan[Float].new`, "make(chan float64)", "Float to float64"},
		{`ch = Chan[Bool].new`, "make(chan bool)", "Bool to bool"},
		{`ch = Chan[String].new`, "make(chan string)", "String to string"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
			continue
		}

		g := New()
		code, err := g.Generate(program)
		if err != nil {
			t.Errorf("%s: codegen error: %v", tt.desc, err)
			continue
		}

		if !strings.Contains(code, tt.expected) {
			t.Errorf("%s: expected '%s', got:\n%s", tt.desc, tt.expected, code)
		}
	}
}

func TestGoBlockClosesOver(t *testing.T) {
	input := `go do
	puts(x)
	puts(y)
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	// Should generate anonymous function
	if !strings.Contains(code, "go func() {") {
		t.Errorf("expected 'go func() {', got:\n%s", code)
	}
	// Should close block
	if !strings.Contains(code, "}()") {
		t.Errorf("expected '}()' to invoke anonymous func, got:\n%s", code)
	}
}

func TestSelectWithMultipleBodies(t *testing.T) {
	input := `select
when a = ch1.receive
	x = process(a)
	log(x)
when ch2 << value
	log("sent")
	cleanup()
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	if !strings.Contains(code, "case a := <-ch1:") {
		t.Errorf("expected 'case a := <-ch1:', got:\n%s", code)
	}
	if !strings.Contains(code, "case ch2 <- value:") {
		t.Errorf("expected 'case ch2 <- value:', got:\n%s", code)
	}
}

func TestSelectWithTryReceive(t *testing.T) {
	// try_receive in select should be treated as regular receive
	input := `select
when val = ch.try_receive
	process(val)
else
	default_action()
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	// Should generate case val := <-ch, not runtime.TryReceive
	if !strings.Contains(code, "case val := <-ch:") {
		t.Errorf("expected 'case val := <-ch:', got:\n%s", code)
	}
	if strings.Contains(code, "TryReceive") {
		t.Errorf("should not contain TryReceive in select case, got:\n%s", code)
	}
}

func TestConcurrentlyGeneratesDefer(t *testing.T) {
	input := `concurrently do |s|
	s.spawn { work() }
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	// Should have both NewScope and defer Wait
	if !strings.Contains(code, "s := runtime.NewScope()") {
		t.Errorf("expected 's := runtime.NewScope()', got:\n%s", code)
	}
	if !strings.Contains(code, "defer s.Wait()") {
		t.Errorf("expected 'defer s.Wait()', got:\n%s", code)
	}
}

func TestChannelSendExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{`ch << 42`, "runtime.ShiftLeft(ch, 42)", "literal"},
		{`ch << x + y`, "runtime.ShiftLeft(ch, (x + y))", "binary expr"},
		{`ch << compute()`, "runtime.ShiftLeft(ch, compute())", "function call"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
			continue
		}

		g := New()
		code, err := g.Generate(program)
		if err != nil {
			t.Errorf("%s: codegen error: %v", tt.desc, err)
			continue
		}

		if !strings.Contains(code, tt.expected) {
			t.Errorf("%s: expected '%s', got:\n%s", tt.desc, tt.expected, code)
		}
	}
}

func TestSpawnReturnsLastExpression(t *testing.T) {
	input := `t = spawn {
	x = 1
	y = 2
	x + y
}`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	// Should return the last expression (may have parentheses)
	if !strings.Contains(code, "return x + y") && !strings.Contains(code, "return (x + y)") {
		t.Errorf("expected 'return x + y' or 'return (x + y)', got:\n%s", code)
	}
}

func TestAwaitWithParentheses(t *testing.T) {
	input := `result = await(task)`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	if !strings.Contains(code, "runtime.Await(task)") {
		t.Errorf("expected 'runtime.Await(task)', got:\n%s", code)
	}
}

func TestChannelCloseWithParens(t *testing.T) {
	input := `ch.close()`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	if !strings.Contains(code, "close(ch)") {
		t.Errorf("expected 'close(ch)', got:\n%s", code)
	}
}

func TestMultipleChannelOperationsSequence(t *testing.T) {
	input := `ch << 1
ch << 2
val1 = ch.receive
val2 = ch.receive`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	if !strings.Contains(code, "runtime.ShiftLeft(ch, 1)") {
		t.Errorf("expected 'runtime.ShiftLeft(ch, 1)', got:\n%s", code)
	}
	if !strings.Contains(code, "runtime.ShiftLeft(ch, 2)") {
		t.Errorf("expected 'runtime.ShiftLeft(ch, 2)', got:\n%s", code)
	}
	if !strings.Contains(code, "val1 := <-ch") {
		t.Errorf("expected 'val1 := <-ch', got:\n%s", code)
	}
	if !strings.Contains(code, "val2 := <-ch") {
		t.Errorf("expected 'val2 := <-ch', got:\n%s", code)
	}
}

func TestScopedSpawnWithMultiStatements(t *testing.T) {
	input := `concurrently do |scope|
	t = scope.spawn do
		a = step1()
		b = step2(a)
		b
	end
	result = await t
end`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	if !strings.Contains(code, "scope.Spawn(func() any {") {
		t.Errorf("expected 'scope.Spawn(func() any {', got:\n%s", code)
	}
	// Should have return for last expression
	if !strings.Contains(code, "return b") {
		t.Errorf("expected 'return b', got:\n%s", code)
	}
}

func TestBufferedChannelCreation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{`ch = Chan[Int].new(0)`, "make(chan int, 0)", "zero buffer"},
		{`ch = Chan[Int].new(1)`, "make(chan int, 1)", "buffer 1"},
		{`ch = Chan[Int].new(100)`, "make(chan int, 100)", "buffer 100"},
		{`ch = Chan[Int].new(size)`, "make(chan int, size)", "variable buffer"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
			continue
		}

		g := New()
		code, err := g.Generate(program)
		if err != nil {
			t.Errorf("%s: codegen error: %v", tt.desc, err)
			continue
		}

		if !strings.Contains(code, tt.expected) {
			t.Errorf("%s: expected '%s', got:\n%s", tt.desc, tt.expected, code)
		}
	}
}

func TestGoWithArgumentCapture(t *testing.T) {
	// When go is used with a call, arguments should be captured
	input := `go process(x, y, z)`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	g := New()
	code, err := g.Generate(program)
	if err != nil {
		t.Errorf("codegen error: %v", err)
		return
	}

	if !strings.Contains(code, "go process(x, y, z)") {
		t.Errorf("expected 'go process(x, y, z)', got:\n%s", code)
	}
}
