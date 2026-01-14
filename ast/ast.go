// Package ast defines the abstract syntax tree nodes for Rugby programs.
package ast

import "strings"

// Comment represents a single comment in the source code.
type Comment struct {
	Text string // the comment text including the # prefix
	Line int    // 1-indexed line number
	Col  int    // 0-indexed column position
}

// CommentGroup represents a sequence of comments with no blank lines between them.
type CommentGroup struct {
	List []*Comment
}

// Text returns the text of the comment group with # prefixes and leading space removed.
func (g *CommentGroup) Text() string {
	if g == nil || len(g.List) == 0 {
		return ""
	}
	var lines []string
	for _, c := range g.List {
		text := c.Text
		if after, found := strings.CutPrefix(text, "#"); found {
			text = strings.TrimPrefix(after, " ") // remove single leading space
		}
		lines = append(lines, text)
	}
	return strings.Join(lines, "\n")
}

// Node is the base interface for all AST nodes
type Node interface {
	node()
}

// Statement is the base interface for statement nodes
type Statement interface {
	Node
	stmtNode()
}

// Expression is the base interface for expression nodes
type Expression interface {
	Node
	exprNode()
}

// Program is the root node of every AST
type Program struct {
	Imports      []*ImportDecl
	Declarations []Statement
	Comments     []*CommentGroup // all comments in file, for free-floating comment handling
}

func (p *Program) node() {}

// ImportDecl represents an import statement
type ImportDecl struct {
	Path    string
	Alias   string        // optional alias
	Line    int           // source line number (1-indexed)
	Doc     *CommentGroup // leading comments
	Comment *CommentGroup // trailing comment on same line
}

func (i *ImportDecl) node()     {}
func (i *ImportDecl) stmtNode() {}

// Param represents a function parameter
type Param struct {
	Name string
	Type string // optional type annotation, empty if not specified
}

// FuncDecl represents a function definition
type FuncDecl struct {
	Name        string
	Params      []*Param
	ReturnTypes []string // empty if not specified
	Body        []Statement
	Pub         bool          // true if exported (pub def)
	Line        int           // source line number (1-indexed)
	Doc         *CommentGroup // leading comments
	Comment     *CommentGroup // trailing comment on same line
}

func (f *FuncDecl) node()     {}
func (f *FuncDecl) stmtNode() {}

// ExprStmt wraps an expression as a statement
type ExprStmt struct {
	Expr Expression
	// Statement modifiers (e.g., "puts x unless valid?")
	Condition Expression    // nil if no modifier
	IsUnless  bool          // true for "unless", false for "if"
	Line      int           // source line number (1-indexed)
	Doc       *CommentGroup // leading comments
	Comment   *CommentGroup // trailing comment on same line
}

func (e *ExprStmt) node()     {}
func (e *ExprStmt) stmtNode() {}

// BlockExpr represents a block: do |params| ... end or { |params| ... }
type BlockExpr struct {
	Params []string    // block parameter names (e.g., |x| or |v, i|)
	Body   []Statement // the statements in the block
}

func (b *BlockExpr) node()     {}
func (b *BlockExpr) exprNode() {}

// CallExpr represents a function call, optionally with a block
type CallExpr struct {
	Func  Expression // function being called
	Args  []Expression
	Block *BlockExpr // optional block argument (nil if no block)
}

func (c *CallExpr) node()     {}
func (c *CallExpr) exprNode() {}

// SelectorKind indicates what a selector expression refers to.
// This is resolved during semantic analysis and used by codegen
// to determine whether to generate field access or method call.
type SelectorKind int

const (
	SelectorUnknown  SelectorKind = iota // not yet resolved
	SelectorField                        // class field access (no parens needed)
	SelectorMethod                       // class method (needs () call)
	SelectorGetter                       // accessor getter (generated method, needs ())
	SelectorSetter                       // used in assignment context (generates setX() call)
	SelectorGoField                      // Go struct field (no parens needed)
	SelectorGoMethod                     // Go method (needs () call)
)

// SelectorExpr represents a selector expression like x.Y or a.b.c
type SelectorExpr struct {
	X   Expression // the expression before the dot
	Sel string     // the selected identifier

	// Resolved during semantic analysis (optional - zero value means unknown)
	ResolvedKind SelectorKind // field, method, getter, setter, etc.
}

func (s *SelectorExpr) node()     {}
func (s *SelectorExpr) exprNode() {}

// StringLit represents a string literal
type StringLit struct {
	Value string
}

func (s *StringLit) node()     {}
func (s *StringLit) exprNode() {}

// InterpolatedString represents a string with embedded expressions: "hello #{name}"
// Parts alternate between string literals (string) and expressions (Expression)
type InterpolatedString struct {
	Parts []any // string or Expression
}

func (i *InterpolatedString) node()     {}
func (i *InterpolatedString) exprNode() {}

// Ident represents an identifier
type Ident struct {
	Name   string
	Line   int // 1-indexed line number
	Column int // 1-indexed column number
}

func (i *Ident) node()     {}
func (i *Ident) exprNode() {}

// IntLit represents an integer literal
type IntLit struct {
	Value  int64
	Line   int // 1-indexed line number
	Column int // 1-indexed column number
}

func (i *IntLit) node()     {}
func (i *IntLit) exprNode() {}

// FloatLit represents a float literal
type FloatLit struct {
	Value  float64
	Line   int // 1-indexed line number
	Column int // 1-indexed column number
}

func (f *FloatLit) node()     {}
func (f *FloatLit) exprNode() {}

// BoolLit represents a boolean literal
type BoolLit struct {
	Value  bool
	Line   int // 1-indexed line number
	Column int // 1-indexed column number
}

func (b *BoolLit) node()     {}
func (b *BoolLit) exprNode() {}

// NilLit represents the nil literal
type NilLit struct {
	Line   int // 1-indexed line number
	Column int // 1-indexed column number
}

func (n *NilLit) node()          {}
func (n *NilLit) exprNode()      {}
func (n *NilLit) String() string { return "nil" }

// SymbolLit represents a symbol literal (:foo, :status, etc.)
type SymbolLit struct {
	Value  string // the symbol name without the colon
	Line   int    // 1-indexed line number
	Column int    // 1-indexed column number
}

func (s *SymbolLit) node()     {}
func (s *SymbolLit) exprNode() {}

// ArrayLit represents an array literal
type ArrayLit struct {
	Elements []Expression
	Line     int // 1-indexed line number
	Column   int // 1-indexed column number
}

func (a *ArrayLit) node()     {}
func (a *ArrayLit) exprNode() {}

// SplatExpr represents a splat expression: *arr
// Used in array literals to spread elements: [1, *rest, 3]
type SplatExpr struct {
	Expr Expression // the expression being splatted
}

func (s *SplatExpr) node()     {}
func (s *SplatExpr) exprNode() {}

// DoubleSplatExpr represents a double splat expression: **hash
// Used in map literals to spread key-value pairs: {**defaults, key: value}
type DoubleSplatExpr struct {
	Expr Expression // the map being splatted
}

func (d *DoubleSplatExpr) node()     {}
func (d *DoubleSplatExpr) exprNode() {}

// IndexExpr represents an index expression like arr[0]
type IndexExpr struct {
	Left  Expression // the expression being indexed
	Index Expression // the index expression
	Line  int        // source line number (1-indexed)
}

func (i *IndexExpr) node()     {}
func (i *IndexExpr) exprNode() {}

// MapEntry represents a key-value pair or a splat in a map literal
type MapEntry struct {
	Key   Expression // nil for splat entries
	Value Expression // nil for splat entries
	Splat Expression // non-nil for **expr splat entries
}

// MapLit represents a map literal like {"a" => 1, "b" => 2}
type MapLit struct {
	Entries []MapEntry
}

func (m *MapLit) node()     {}
func (m *MapLit) exprNode() {}

// RangeLit represents a range literal like 1..10 or 0...n
type RangeLit struct {
	Start     Expression // start of range
	End       Expression // end of range
	Exclusive bool       // true for ... (exclusive), false for .. (inclusive)
	Line      int        // line number for error reporting
}

func (r *RangeLit) node()     {}
func (r *RangeLit) exprNode() {}

// BinaryExpr represents a binary operation
type BinaryExpr struct {
	Left  Expression
	Op    string
	Right Expression
}

func (b *BinaryExpr) node()     {}
func (b *BinaryExpr) exprNode() {}

// UnaryExpr represents a unary operation
type UnaryExpr struct {
	Op   string
	Expr Expression
}

func (u *UnaryExpr) node()     {}
func (u *UnaryExpr) exprNode() {}

// BangExpr represents error propagation: call()!
// Unwraps (T, error) and propagates error on failure
type BangExpr struct {
	Expr Expression // must be a CallExpr
	Line int        // source line number (1-indexed)
}

func (b *BangExpr) node()     {}
func (b *BangExpr) exprNode() {}

// RescueExpr represents error recovery: call() rescue default
type RescueExpr struct {
	Expr    Expression // the fallible call
	Default Expression // default value (for inline form), nil for block form
	Block   *BlockExpr // block form (nil for inline)
	ErrName string     // error binding name (for => err do form), empty if no binding
}

func (r *RescueExpr) node()     {}
func (r *RescueExpr) exprNode() {}

// NilCoalesceExpr represents the ?? operator: expr ?? default
// Returns expr if present, otherwise default
type NilCoalesceExpr struct {
	Left  Expression // optional expression (T?)
	Right Expression // default value (T)
}

func (n *NilCoalesceExpr) node()     {}
func (n *NilCoalesceExpr) exprNode() {}

// TernaryExpr represents the ternary conditional: cond ? then : else
// Ruby: status = valid? ? "ok" : "error"
type TernaryExpr struct {
	Condition Expression // the condition to test
	Then      Expression // value if condition is truthy
	Else      Expression // value if condition is falsy
	Line      int        // source line number (1-indexed)
}

func (t *TernaryExpr) node()     {}
func (t *TernaryExpr) exprNode() {}

// SymbolToProcExpr represents the &:method syntax
// Converts a method name into a block/lambda that calls that method on its argument
// Example: names.map(&:upcase) → names.map { |x| x.upcase }
type SymbolToProcExpr struct {
	Method string // the method name (without leading colon)
}

func (s *SymbolToProcExpr) node()     {}
func (s *SymbolToProcExpr) exprNode() {}

// SafeNavExpr represents the &. operator: expr&.method
// Calls method only if expr is present, returns optional
type SafeNavExpr struct {
	Receiver Expression // optional expression (T?)
	Selector string     // method/property name
}

func (s *SafeNavExpr) node()     {}
func (s *SafeNavExpr) exprNode() {}

// AssignStmt represents variable assignment
type AssignStmt struct {
	Name    string
	Type    string // optional type annotation, empty if not specified
	Value   Expression
	Line    int           // source line number (1-indexed)
	Doc     *CommentGroup // leading comments
	Comment *CommentGroup // trailing comment on same line
}

func (a *AssignStmt) node()     {}
func (a *AssignStmt) stmtNode() {}

// SelectorAssignStmt represents setter assignment: obj.field = value
// This generates a setter method call: obj.setField(value)
type SelectorAssignStmt struct {
	Object Expression // the receiver object
	Field  string     // the field/property name
	Value  Expression // the value being assigned
	Line   int        // source line number (1-indexed)
}

func (s *SelectorAssignStmt) node()     {}
func (s *SelectorAssignStmt) stmtNode() {}

// OrAssignStmt represents x ||= y (logical or assignment)
type OrAssignStmt struct {
	Name  string
	Value Expression
	Line  int // source line number (1-indexed)
}

func (o *OrAssignStmt) node()     {}
func (o *OrAssignStmt) stmtNode() {}

// CompoundAssignStmt represents x += y, x -= y, x *= y, x /= y
type CompoundAssignStmt struct {
	Name  string // variable name
	Op    string // operator: "+", "-", "*", "/"
	Value Expression
	Line  int // source line number (1-indexed)
}

func (c *CompoundAssignStmt) node()     {}
func (c *CompoundAssignStmt) stmtNode() {}

// MultiAssignStmt represents tuple unpacking: val, ok = expr
type MultiAssignStmt struct {
	Names []string   // variable names (e.g., ["val", "ok"])
	Value Expression // expression returning multiple values
	Line  int        // source line number (1-indexed)
}

func (m *MultiAssignStmt) node()     {}
func (m *MultiAssignStmt) stmtNode() {}

// IfStmt represents an if/elsif/else statement or unless/else statement
type IfStmt struct {
	// Optional assignment in condition: if (name = expr)
	// When set, generates: if name, ok := expr; ok { ... }
	AssignName string     // variable name for assignment (empty if none)
	AssignExpr Expression // expression to assign (nil if none)

	IsUnless bool // true for "unless", false for "if" (inverts condition)
	Cond     Expression
	Then     []Statement
	ElseIfs  []ElseIfClause
	Else     []Statement
	Line     int           // source line number (1-indexed)
	Doc      *CommentGroup // leading comments
	Comment  *CommentGroup // trailing comment on same line
}

func (i *IfStmt) node()     {}
func (i *IfStmt) stmtNode() {}

// ElseIfClause represents an elsif branch
type ElseIfClause struct {
	Cond Expression
	Body []Statement
}

// CaseStmt represents a case/when/else statement
type CaseStmt struct {
	Subject     Expression    // the expression being matched (nil for case without subject)
	WhenClauses []WhenClause  // one or more when branches
	Else        []Statement   // optional else clause
	Line        int           // source line number (1-indexed)
	Doc         *CommentGroup // leading comments
	Comment     *CommentGroup // trailing comment on same line
}

func (c *CaseStmt) node()     {}
func (c *CaseStmt) stmtNode() {}

// WhenClause represents a single when branch in a case statement
type WhenClause struct {
	Values []Expression // one or more values to match (can be comma-separated)
	Body   []Statement  // statements in this when branch
}

// CaseTypeStmt represents a type switch statement: case_type x when String ... when Int ... end
type CaseTypeStmt struct {
	Subject     Expression       // the expression being type-matched
	WhenClauses []TypeWhenClause // one or more type branches
	Else        []Statement      // optional else clause
	Line        int              // source line number (1-indexed)
	Doc         *CommentGroup    // leading comments
	Comment     *CommentGroup    // trailing comment on same line
}

func (c *CaseTypeStmt) node()     {}
func (c *CaseTypeStmt) stmtNode() {}

// TypeWhenClause represents a single when branch in a case_type statement
type TypeWhenClause struct {
	Type string      // the type to match (String, Int, etc.)
	Body []Statement // statements in this when branch
}

// WhileStmt represents a while loop
type WhileStmt struct {
	Cond    Expression
	Body    []Statement
	Line    int           // source line number (1-indexed)
	Doc     *CommentGroup // leading comments
	Comment *CommentGroup // trailing comment on same line
}

func (w *WhileStmt) node()     {}
func (w *WhileStmt) stmtNode() {}

// UntilStmt represents an until loop (inverse of while)
type UntilStmt struct {
	Cond    Expression
	Body    []Statement
	Line    int           // source line number (1-indexed)
	Doc     *CommentGroup // leading comments
	Comment *CommentGroup // trailing comment on same line
}

func (u *UntilStmt) node()     {}
func (u *UntilStmt) stmtNode() {}

// ForStmt represents a for...in loop: for item in items ... end
type ForStmt struct {
	Var      string     // loop variable name
	Iterable Expression // the collection to iterate over
	Body     []Statement
	Line     int           // source line number (1-indexed)
	Doc      *CommentGroup // leading comments
	Comment  *CommentGroup // trailing comment on same line
}

func (f *ForStmt) node()     {}
func (f *ForStmt) stmtNode() {}

// BreakStmt represents a break statement (exits loop)
type BreakStmt struct {
	// Statement modifiers (e.g., "break if x == 2")
	Condition Expression // nil if no modifier
	IsUnless  bool       // true for "unless", false for "if"
	Line      int        // source line number (1-indexed)
}

func (b *BreakStmt) node()     {}
func (b *BreakStmt) stmtNode() {}

// NextStmt represents a next statement (continues to next iteration)
type NextStmt struct {
	// Statement modifiers (e.g., "next unless valid?")
	Condition Expression // nil if no modifier
	IsUnless  bool       // true for "unless", false for "if"
	Line      int        // source line number (1-indexed)
}

func (n *NextStmt) node()     {}
func (n *NextStmt) stmtNode() {}

// ReturnStmt represents a return statement
type ReturnStmt struct {
	Values []Expression // empty if no return values
	// Statement modifiers (e.g., "return if error")
	Condition Expression    // nil if no modifier
	IsUnless  bool          // true for "unless", false for "if"
	Line      int           // source line number (1-indexed)
	Doc       *CommentGroup // leading comments
	Comment   *CommentGroup // trailing comment on same line
}

func (r *ReturnStmt) node()     {}
func (r *ReturnStmt) stmtNode() {}

// PanicStmt represents a panic: panic "message" [if|unless cond]
type PanicStmt struct {
	Message   Expression // the panic message/value
	Condition Expression // optional condition for statement modifier (panic "msg" if cond)
	IsUnless  bool       // true if modifier is "unless" instead of "if"
	Line      int        // source line number
}

func (r *PanicStmt) node()     {}
func (r *PanicStmt) stmtNode() {}

// DeferStmt represents a defer statement
type DeferStmt struct {
	Call *CallExpr
	Line int // source line number (1-indexed)
}

func (d *DeferStmt) node()     {}
func (d *DeferStmt) stmtNode() {}

// ClassDecl represents a class definition
type ClassDecl struct {
	Name       string          // class name (e.g., "User")
	Embeds     []string        // embedded types (Go struct embedding), empty if none
	Implements []string        // interfaces this class explicitly implements
	Fields     []*FieldDecl    // fields inferred from initialize
	Methods    []*MethodDecl   // methods defined in class
	Accessors  []*AccessorDecl // accessor declarations (getter, setter, property)
	Includes   []string        // modules to include
	Pub        bool            // true if exported (pub class)
	Line       int             // source line number (1-indexed)
	Doc        *CommentGroup   // leading comments
	Comment    *CommentGroup   // trailing comment on same line
}

func (c *ClassDecl) node()     {}
func (c *ClassDecl) stmtNode() {}

// FieldDecl represents a struct field (inferred from @var in initialize)
type FieldDecl struct {
	Name string // field name (without @)
	Type string // inferred from initialize params or explicit annotation
}

// MethodDecl represents a method definition within a class
type MethodDecl struct {
	Name          string        // method name (may end with ? for predicates)
	Params        []*Param      // parameters
	ReturnTypes   []string      // return types
	Body          []Statement   // method body
	Pub           bool          // true if exported (pub def)
	IsClassMethod bool          // true if class method (def self.method)
	Line          int           // source line number (1-indexed)
	Doc           *CommentGroup // leading comments
	Comment       *CommentGroup // trailing comment on same line
}

func (m *MethodDecl) node()     {}
func (m *MethodDecl) stmtNode() {}

// InstanceVar represents an instance variable reference (@name)
type InstanceVar struct {
	Name string // variable name without @
}

func (i *InstanceVar) node()     {}
func (i *InstanceVar) exprNode() {}

// InstanceVarAssign represents @name = value
type InstanceVarAssign struct {
	Name  string // variable name without @
	Value Expression
}

func (i *InstanceVarAssign) node()     {}
func (i *InstanceVarAssign) stmtNode() {}

// InstanceVarOrAssign represents @name ||= value
type InstanceVarOrAssign struct {
	Name  string // variable name without @
	Value Expression
}

func (i *InstanceVarOrAssign) node()     {}
func (i *InstanceVarOrAssign) stmtNode() {}

// InstanceVarCompoundAssign represents @name += value, @name -= value, etc.
type InstanceVarCompoundAssign struct {
	Name  string // variable name without @
	Op    string // operator: "+", "-", "*", "/"
	Value Expression
	Line  int
}

func (i *InstanceVarCompoundAssign) node()     {}
func (i *InstanceVarCompoundAssign) stmtNode() {}

// InterfaceDecl represents an interface definition
type InterfaceDecl struct {
	Name    string        // interface name (e.g., "Speaker")
	Parents []string      // parent interfaces for embedding (e.g., ["Reader", "Writer"])
	Methods []*MethodSig  // method signatures (no body)
	Pub     bool          // true if exported (pub interface)
	Line    int           // source line number (1-indexed)
	Doc     *CommentGroup // leading comments
	Comment *CommentGroup // trailing comment on same line
}

func (i *InterfaceDecl) node()     {}
func (i *InterfaceDecl) stmtNode() {}

// MethodSig represents a method signature in an interface (no body)
type MethodSig struct {
	Name        string   // method name
	Params      []*Param // parameters
	ReturnTypes []string // return types
}

// DescribeStmt represents a describe block for grouping tests
// describe "name" do ... end
type DescribeStmt struct {
	Name   string        // description string
	Body   []Statement   // test cases (it, describe, before, after)
	Line   int           // source line number
	Parent *DescribeStmt // parent describe block for nesting (nil if top-level)
}

func (d *DescribeStmt) node()     {}
func (d *DescribeStmt) stmtNode() {}

// ItStmt represents a single test case inside a describe block
// it "description" do |t| ... end
type ItStmt struct {
	Name string      // test description
	Body []Statement // test body
	Line int         // source line number
}

func (i *ItStmt) node()     {}
func (i *ItStmt) stmtNode() {}

// TestStmt represents a standalone test (top-level, no describe)
// test "name" do |t| ... end
type TestStmt struct {
	Name string      // test name
	Body []Statement // test body
	Line int         // source line number
}

func (t *TestStmt) node()     {}
func (t *TestStmt) stmtNode() {}

// TableStmt represents a table-driven test
// table "name" do |tt| ... end
type TableStmt struct {
	Name string      // test name
	Body []Statement // table test body (contains tt.case and tt.run calls)
	Line int         // source line number
}

func (t *TableStmt) node()     {}
func (t *TableStmt) stmtNode() {}

// BeforeStmt represents a before hook for test setup
// before do |t| ... end
type BeforeStmt struct {
	Body []Statement // setup code
	Line int         // source line number
}

func (b *BeforeStmt) node()     {}
func (b *BeforeStmt) stmtNode() {}

// AfterStmt represents an after hook for test teardown
// after do |t, ctx| ... end
type AfterStmt struct {
	Body []Statement // teardown code
	Line int         // source line number
}

func (a *AfterStmt) node()     {}
func (a *AfterStmt) stmtNode() {}

// AccessorDecl represents a getter, setter, or property declaration in a class
// getter name : Type → declares @name : Type + generates def name -> Type
// setter name : Type → declares @name : Type + generates def name=(v : Type)
// property name : Type → declares @name : Type + generates both getter and setter
type AccessorDecl struct {
	Kind string // "getter", "setter", or "property"
	Name string // field name (without @)
	Type string // field type
	Line int    // source line number
}

func (a *AccessorDecl) node()     {}
func (a *AccessorDecl) stmtNode() {}

// SuperExpr represents a call to the parent's implementation
// super or super(args...)
type SuperExpr struct {
	Args []Expression // arguments to pass to parent method
	Line int          // source line number
}

func (s *SuperExpr) node()     {}
func (s *SuperExpr) exprNode() {}

// GoStmt represents a goroutine spawn: go expr or go do ... end
type GoStmt struct {
	Call  Expression // call expression to execute in goroutine
	Block *BlockExpr // optional block form: go do ... end
	Line  int        // source line number
}

func (g *GoStmt) node()     {}
func (g *GoStmt) stmtNode() {}

// ChanSendStmt represents a channel send: ch << value
type ChanSendStmt struct {
	Chan  Expression // the channel
	Value Expression // the value to send
	Line  int        // source line number
}

func (c *ChanSendStmt) node()     {}
func (c *ChanSendStmt) stmtNode() {}

// SelectStmt represents a select statement for channel operations
type SelectStmt struct {
	Cases []SelectCase // when clauses
	Else  []Statement  // optional else/default clause
	Line  int          // source line number
}

func (s *SelectStmt) node()     {}
func (s *SelectStmt) stmtNode() {}

// SelectCase represents a single case in a select statement
type SelectCase struct {
	// Either a receive (val = ch.receive) or send (ch << val)
	AssignName string     // variable name for receive assignment (empty if none)
	Chan       Expression // the channel expression
	IsSend     bool       // true for send, false for receive
	Value      Expression // value to send (for send case) or nil (for receive)
	Body       []Statement
}

// ModuleDecl represents a module definition
type ModuleDecl struct {
	Name      string          // module name
	Fields    []*FieldDecl    // fields defined in module
	Methods   []*MethodDecl   // methods defined in module
	Accessors []*AccessorDecl // accessor declarations
	Pub       bool            // true if exported
	Line      int             // source line number
	Doc       *CommentGroup   // associated doc comment
}

func (m *ModuleDecl) node()     {}
func (m *ModuleDecl) stmtNode() {}

// IncludeStmt represents including a module in a class
// include ModuleName
type IncludeStmt struct {
	Module string // module name to include
	Line   int    // source line number
}

func (i *IncludeStmt) node()     {}
func (i *IncludeStmt) stmtNode() {}

// SpawnExpr represents spawning a concurrent task: spawn { expr }
// Returns Task[T] where T is the block's return type
type SpawnExpr struct {
	Block *BlockExpr // the block to execute concurrently
	Line  int        // source line number
}

func (s *SpawnExpr) node()     {}
func (s *SpawnExpr) exprNode() {}

// AwaitExpr represents awaiting a task: await t or await(t)
// Blocks until the task completes and returns its value
type AwaitExpr struct {
	Task Expression // the Task[T] to await
	Line int        // source line number
}

func (a *AwaitExpr) node()     {}
func (a *AwaitExpr) exprNode() {}

// ConcurrentlyStmt represents structured concurrency: concurrently do |scope| ... end
// All tasks spawned within the scope are awaited when the block exits
type ConcurrentlyStmt struct {
	ScopeVar string      // the scope variable name (e.g., "scope")
	Body     []Statement // statements inside the block
	Line     int         // source line number
}

func (c *ConcurrentlyStmt) node()     {}
func (c *ConcurrentlyStmt) stmtNode() {}

// ConcurrentlyExpr is the expression form of concurrently: result = concurrently do |scope| ... end
// Returns the last expression in the block
type ConcurrentlyExpr struct {
	ScopeVar string      // the scope variable name (e.g., "scope")
	Body     []Statement // statements inside the block
	Line     int         // source line number
}

func (c *ConcurrentlyExpr) node()     {}
func (c *ConcurrentlyExpr) exprNode() {}
