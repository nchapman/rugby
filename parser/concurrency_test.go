package parser

import (
	"testing"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
)

func TestParseGoStatement(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{`go do
			puts("hello")
		end`, "go with do block"},
		{`go fetch_url(url)`, "go with call expression"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
			continue
		}

		if len(program.Declarations) != 1 {
			t.Errorf("%s: expected 1 declaration, got %d", tt.desc, len(program.Declarations))
			continue
		}

		exprStmt, ok := program.Declarations[0].(*ast.ExprStmt)
		if !ok {
			// Could be a GoStmt directly
			if _, ok := program.Declarations[0].(*ast.GoStmt); !ok {
				t.Errorf("%s: expected GoStmt or ExprStmt, got %T", tt.desc, program.Declarations[0])
			}
		} else {
			_ = exprStmt
		}
	}
}

func TestParseSpawnExpression(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{`t = spawn { 42 }`, "spawn with brace block"},
		{`t = spawn do
			compute()
		end`, "spawn with do block"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
			continue
		}

		if len(program.Declarations) != 1 {
			t.Errorf("%s: expected 1 declaration, got %d", tt.desc, len(program.Declarations))
			continue
		}

		assignStmt, ok := program.Declarations[0].(*ast.AssignStmt)
		if !ok {
			t.Errorf("%s: expected AssignStmt, got %T", tt.desc, program.Declarations[0])
			continue
		}

		_, ok = assignStmt.Value.(*ast.SpawnExpr)
		if !ok {
			t.Errorf("%s: expected SpawnExpr, got %T", tt.desc, assignStmt.Value)
		}
	}
}

func TestParseAwaitExpression(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{`result = await t`, "await without parens"},
		{`result = await(t)`, "await with parens"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
			continue
		}

		if len(program.Declarations) != 1 {
			t.Errorf("%s: expected 1 declaration, got %d", tt.desc, len(program.Declarations))
			continue
		}

		assignStmt, ok := program.Declarations[0].(*ast.AssignStmt)
		if !ok {
			t.Errorf("%s: expected AssignStmt, got %T", tt.desc, program.Declarations[0])
			continue
		}

		_, ok = assignStmt.Value.(*ast.AwaitExpr)
		if !ok {
			t.Errorf("%s: expected AwaitExpr, got %T", tt.desc, assignStmt.Value)
		}
	}
}

func TestParseSelectStatement(t *testing.T) {
	input := `select
when val = ch1.receive
	puts(val)
when ch2 << 42
	puts("sent")
else
	puts("default")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	if len(program.Declarations) != 1 {
		t.Errorf("expected 1 declaration, got %d", len(program.Declarations))
		return
	}

	selectStmt, ok := program.Declarations[0].(*ast.SelectStmt)
	if !ok {
		t.Errorf("expected SelectStmt, got %T", program.Declarations[0])
		return
	}

	if len(selectStmt.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(selectStmt.Cases))
	}

	// First case: receive
	if selectStmt.Cases[0].IsSend {
		t.Error("expected first case to be receive")
	}
	if selectStmt.Cases[0].AssignName != "val" {
		t.Errorf("expected assign name 'val', got '%s'", selectStmt.Cases[0].AssignName)
	}

	// Second case: send
	if !selectStmt.Cases[1].IsSend {
		t.Error("expected second case to be send")
	}

	// Else clause
	if len(selectStmt.Else) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(selectStmt.Else))
	}
}

func TestParseConcurrentlyStatement(t *testing.T) {
	input := `concurrently do |scope|
	a = scope.spawn { fetch_a() }
	b = scope.spawn { fetch_b() }
	x = await a
	y = await b
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	if len(program.Declarations) != 1 {
		t.Errorf("expected 1 declaration, got %d", len(program.Declarations))
		return
	}

	concStmt, ok := program.Declarations[0].(*ast.ConcurrentlyStmt)
	if !ok {
		t.Errorf("expected ConcurrentlyStmt, got %T", program.Declarations[0])
		return
	}

	if concStmt.ScopeVar != "scope" {
		t.Errorf("expected scope var 'scope', got '%s'", concStmt.ScopeVar)
	}

	if len(concStmt.Body) != 4 {
		t.Errorf("expected 4 statements in body, got %d", len(concStmt.Body))
	}
}

func TestParseChannelOperations(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{`ch = Chan[Int].new(10)`, "buffered channel creation"},
		{`ch = Chan[String].new`, "unbuffered channel creation"},
		{`ch << 42`, "channel send"},
		{`val = ch.receive`, "channel receive"},
		{`val = ch.try_receive`, "non-blocking receive"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
			continue
		}

		if len(program.Declarations) != 1 {
			t.Errorf("%s: expected 1 declaration, got %d", tt.desc, len(program.Declarations))
		}
	}
}

func TestParseChannelIteration(t *testing.T) {
	input := `for msg in ch
	puts(msg)
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	if len(program.Declarations) != 1 {
		t.Errorf("expected 1 declaration, got %d", len(program.Declarations))
		return
	}

	forStmt, ok := program.Declarations[0].(*ast.ForStmt)
	if !ok {
		t.Errorf("expected ForStmt, got %T", program.Declarations[0])
		return
	}

	if forStmt.Var != "msg" {
		t.Errorf("expected loop var 'msg', got '%s'", forStmt.Var)
	}
}

// Additional comprehensive tests for edge cases

func TestParseSelectWithoutElse(t *testing.T) {
	input := `select
when val = ch.receive
	process(val)
when ch2 << data
	log("sent")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	selectStmt, ok := program.Declarations[0].(*ast.SelectStmt)
	if !ok {
		t.Errorf("expected SelectStmt, got %T", program.Declarations[0])
		return
	}

	if len(selectStmt.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(selectStmt.Cases))
	}

	if len(selectStmt.Else) != 0 {
		t.Errorf("expected no else clause, got %d statements", len(selectStmt.Else))
	}
}

func TestParseSelectReceiveOnly(t *testing.T) {
	input := `select
when a = ch1.receive
	handle_a(a)
when b = ch2.receive
	handle_b(b)
when c = ch3.receive
	handle_c(c)
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	selectStmt := program.Declarations[0].(*ast.SelectStmt)

	if len(selectStmt.Cases) != 3 {
		t.Errorf("expected 3 cases, got %d", len(selectStmt.Cases))
	}

	for i, c := range selectStmt.Cases {
		if c.IsSend {
			t.Errorf("case %d: expected receive, got send", i)
		}
	}
}

func TestParseSelectSendOnly(t *testing.T) {
	input := `select
when ch1 << 1
	log("sent 1")
when ch2 << 2
	log("sent 2")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	selectStmt := program.Declarations[0].(*ast.SelectStmt)

	if len(selectStmt.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(selectStmt.Cases))
	}

	for i, c := range selectStmt.Cases {
		if !c.IsSend {
			t.Errorf("case %d: expected send, got receive", i)
		}
	}
}

func TestParseGoWithMethodChain(t *testing.T) {
	input := `go server.handler.process(request)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	if len(program.Declarations) != 1 {
		t.Errorf("expected 1 declaration, got %d", len(program.Declarations))
	}

	_, ok := program.Declarations[0].(*ast.GoStmt)
	if !ok {
		t.Errorf("expected GoStmt, got %T", program.Declarations[0])
	}
}

func TestParseSpawnWithMultipleStatements(t *testing.T) {
	input := `t = spawn do
	x = compute_step1()
	y = compute_step2(x)
	z = compute_step3(y)
	z * 2
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	assignStmt := program.Declarations[0].(*ast.AssignStmt)
	spawnExpr, ok := assignStmt.Value.(*ast.SpawnExpr)
	if !ok {
		t.Errorf("expected SpawnExpr, got %T", assignStmt.Value)
		return
	}

	if len(spawnExpr.Block.Body) != 4 {
		t.Errorf("expected 4 statements in spawn block, got %d", len(spawnExpr.Block.Body))
	}
}

func TestParseMultipleSpawnAwait(t *testing.T) {
	input := `t1 = spawn { fetch_a() }
t2 = spawn { fetch_b() }
t3 = spawn { fetch_c() }
a = await t1
b = await t2
c = await t3`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	if len(program.Declarations) != 6 {
		t.Errorf("expected 6 declarations, got %d", len(program.Declarations))
	}
}

func TestParseChannelTypesVariety(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{`ch = Chan[Int].new`, "Int channel"},
		{`ch = Chan[String].new`, "String channel"},
		{`ch = Chan[Bool].new`, "Bool channel"},
		{`ch = Chan[Float].new`, "Float channel"},
		{`ch = Chan[User].new`, "Custom type channel"},
		{`ch = Chan[Int].new(0)`, "zero buffer"},
		{`ch = Chan[Int].new(1)`, "buffer 1"},
		{`ch = Chan[Int].new(100)`, "buffer 100"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		_ = p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
		}
	}
}

func TestParseConcurrentlyWithComplexBody(t *testing.T) {
	input := `concurrently do |s|
	a = s.spawn { fetch_a() }
	b = s.spawn { fetch_b() }
	x = await a
	y = await b
	z = x + y
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	concStmt, ok := program.Declarations[0].(*ast.ConcurrentlyStmt)
	if !ok {
		t.Errorf("expected ConcurrentlyStmt, got %T", program.Declarations[0])
		return
	}

	if concStmt.ScopeVar != "s" {
		t.Errorf("expected scope var 's', got '%s'", concStmt.ScopeVar)
	}

	if len(concStmt.Body) != 5 {
		t.Errorf("expected 5 statements in body, got %d", len(concStmt.Body))
	}
}

func TestParseChannelSendExpressionValue(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{`ch << 42`, "send literal"},
		{`ch << x`, "send variable"},
		{`ch << compute()`, "send function result"},
		{`ch << a + b`, "send expression"},
		{`ch << "hello"`, "send string"},
		{`ch << arr[0]`, "send array element"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		_ = p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
		}
	}
}

func TestParseAwaitInExpression(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{`x = await t`, "simple await"},
		{`x = await(t)`, "await with parens"},
		{`x = (await t) + 1`, "await in expression"},
		{`puts(await t)`, "await as argument"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		_ = p.ParseProgram()

		if len(p.Errors()) != 0 {
			t.Errorf("%s: parser errors: %v", tt.desc, p.Errors())
		}
	}
}

func TestParseSelectWithTryReceive(t *testing.T) {
	input := `select
when val = ch.try_receive
	process(val)
else
	do_something_else()
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	selectStmt := program.Declarations[0].(*ast.SelectStmt)
	if len(selectStmt.Cases) != 1 {
		t.Errorf("expected 1 case, got %d", len(selectStmt.Cases))
	}
}

func TestParseSpawnEmptyBlock(t *testing.T) {
	input := `t = spawn { nil }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	assignStmt := program.Declarations[0].(*ast.AssignStmt)
	_, ok := assignStmt.Value.(*ast.SpawnExpr)
	if !ok {
		t.Errorf("expected SpawnExpr, got %T", assignStmt.Value)
	}
}

func TestParseGoDoWithMultipleStatements(t *testing.T) {
	input := `go do
	setup()
	process()
	cleanup()
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	goStmt, ok := program.Declarations[0].(*ast.GoStmt)
	if !ok {
		t.Errorf("expected GoStmt, got %T", program.Declarations[0])
		return
	}

	if goStmt.Block == nil {
		t.Error("expected block in go statement")
		return
	}

	if len(goStmt.Block.Body) != 3 {
		t.Errorf("expected 3 statements in block, got %d", len(goStmt.Block.Body))
	}
}

func TestParseChannelMethodChain(t *testing.T) {
	// This tests that channel operations can be part of larger expressions
	input := `val = ch.receive`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Errorf("parser errors: %v", p.Errors())
		return
	}

	assignStmt := program.Declarations[0].(*ast.AssignStmt)
	selExpr, ok := assignStmt.Value.(*ast.SelectorExpr)
	if !ok {
		t.Errorf("expected SelectorExpr, got %T", assignStmt.Value)
		return
	}

	if selExpr.Sel != "receive" {
		t.Errorf("expected selector 'receive', got '%s'", selExpr.Sel)
	}
}
