package parser

import (
	"testing"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/lexer"
)

func TestParseDescribe(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantItCount int
	}{
		{
			name: "simple describe",
			input: `describe "User" do
  it "has a name" do |t|
    puts("test")
  end
end`,
			wantName:    "User",
			wantItCount: 1,
		},
		{
			name: "describe with multiple its",
			input: `describe "Math" do
  it "adds numbers" do |t|
    puts("add")
  end
  it "subtracts numbers" do |t|
    puts("sub")
  end
end`,
			wantName:    "Math",
			wantItCount: 2,
		},
		{
			name: "nested describe",
			input: `describe "User" do
  describe "Validation" do
    it "requires name" do |t|
      puts("test")
    end
  end
end`,
			wantName:    "User",
			wantItCount: 0, // The it is inside nested describe
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			if len(program.Declarations) != 1 {
				t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
			}

			desc, ok := program.Declarations[0].(*ast.DescribeStmt)
			if !ok {
				t.Fatalf("expected DescribeStmt, got %T", program.Declarations[0])
			}

			if desc.Name != tt.wantName {
				t.Errorf("name = %q, want %q", desc.Name, tt.wantName)
			}

			// Count it statements directly in this describe (not nested)
			itCount := 0
			for _, stmt := range desc.Body {
				if _, ok := stmt.(*ast.ItStmt); ok {
					itCount++
				}
			}
			if itCount != tt.wantItCount {
				t.Errorf("it count = %d, want %d", itCount, tt.wantItCount)
			}
		})
	}
}

func TestParseTest(t *testing.T) {
	input := `test "math/add" do |t|
  x = 1 + 2
  puts(x)
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	testStmt, ok := program.Declarations[0].(*ast.TestStmt)
	if !ok {
		t.Fatalf("expected TestStmt, got %T", program.Declarations[0])
	}

	if testStmt.Name != "math/add" {
		t.Errorf("name = %q, want %q", testStmt.Name, "math/add")
	}

	if len(testStmt.Body) != 2 {
		t.Errorf("body length = %d, want 2", len(testStmt.Body))
	}
}

func TestParseTable(t *testing.T) {
	input := `table "String#to_i?" do |tt|
  puts("case 1")
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	tableStmt, ok := program.Declarations[0].(*ast.TableStmt)
	if !ok {
		t.Fatalf("expected TableStmt, got %T", program.Declarations[0])
	}

	if tableStmt.Name != "String#to_i?" {
		t.Errorf("name = %q, want %q", tableStmt.Name, "String#to_i?")
	}
}

func TestParseBeforeAfter(t *testing.T) {
	input := `describe "Database" do
  before do |t|
    puts("setup")
  end

  after do |t|
    puts("teardown")
  end

  it "works" do |t|
    puts("test")
  end
end`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	desc, ok := program.Declarations[0].(*ast.DescribeStmt)
	if !ok {
		t.Fatalf("expected DescribeStmt, got %T", program.Declarations[0])
	}

	var beforeCount, afterCount, itCount int
	for _, stmt := range desc.Body {
		switch stmt.(type) {
		case *ast.BeforeStmt:
			beforeCount++
		case *ast.AfterStmt:
			afterCount++
		case *ast.ItStmt:
			itCount++
		}
	}

	if beforeCount != 1 {
		t.Errorf("before count = %d, want 1", beforeCount)
	}
	if afterCount != 1 {
		t.Errorf("after count = %d, want 1", afterCount)
	}
	if itCount != 1 {
		t.Errorf("it count = %d, want 1", itCount)
	}
}
