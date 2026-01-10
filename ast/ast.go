package ast

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
}

func (p *Program) node() {}

// ImportDecl represents an import statement
type ImportDecl struct {
	Path  string
	Alias string // optional alias
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
	Pub         bool // true if exported (pub def)
}

func (f *FuncDecl) node()     {}
func (f *FuncDecl) stmtNode() {}

// ExprStmt wraps an expression as a statement
type ExprStmt struct {
	Expr Expression
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

// SelectorExpr represents a selector expression like x.Y or a.b.c
type SelectorExpr struct {
	X   Expression // the expression before the dot
	Sel string     // the selected identifier
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
	Parts []interface{} // string or Expression
}

func (i *InterpolatedString) node()     {}
func (i *InterpolatedString) exprNode() {}

// Ident represents an identifier
type Ident struct {
	Name string
}

func (i *Ident) node()     {}
func (i *Ident) exprNode() {}

// IntLit represents an integer literal
type IntLit struct {
	Value int64
}

func (i *IntLit) node()     {}
func (i *IntLit) exprNode() {}

// FloatLit represents a float literal
type FloatLit struct {
	Value float64
}

func (f *FloatLit) node()     {}
func (f *FloatLit) exprNode() {}

// BoolLit represents a boolean literal
type BoolLit struct {
	Value bool
}

func (b *BoolLit) node()     {}
func (b *BoolLit) exprNode() {}

// ArrayLit represents an array literal
type ArrayLit struct {
	Elements []Expression
}

func (a *ArrayLit) node()     {}
func (a *ArrayLit) exprNode() {}

// IndexExpr represents an index expression like arr[0]
type IndexExpr struct {
	Left  Expression // the expression being indexed
	Index Expression // the index expression
}

func (i *IndexExpr) node()     {}
func (i *IndexExpr) exprNode() {}

// MapEntry represents a key-value pair in a map literal
type MapEntry struct {
	Key   Expression
	Value Expression
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

// AssignStmt represents variable assignment
type AssignStmt struct {
	Name  string
	Type  string // optional type annotation, empty if not specified
	Value Expression
}

func (a *AssignStmt) node()     {}
func (a *AssignStmt) stmtNode() {}

// OrAssignStmt represents x ||= y (logical or assignment)
type OrAssignStmt struct {
	Name  string
	Value Expression
}

func (o *OrAssignStmt) node()     {}
func (o *OrAssignStmt) stmtNode() {}

// IfStmt represents an if/elsif/else statement
type IfStmt struct {
	Cond    Expression
	Then    []Statement
	ElseIfs []ElseIfClause
	Else    []Statement
}

func (i *IfStmt) node()     {}
func (i *IfStmt) stmtNode() {}

// ElseIfClause represents an elsif branch
type ElseIfClause struct {
	Cond Expression
	Body []Statement
}

// WhileStmt represents a while loop
type WhileStmt struct {
	Cond Expression
	Body []Statement
}

func (w *WhileStmt) node()     {}
func (w *WhileStmt) stmtNode() {}

// ForStmt represents a for...in loop: for item in items ... end
type ForStmt struct {
	Var      string     // loop variable name
	Iterable Expression // the collection to iterate over
	Body     []Statement
}

func (f *ForStmt) node()     {}
func (f *ForStmt) stmtNode() {}

// BreakStmt represents a break statement (exits loop)
type BreakStmt struct{}

func (b *BreakStmt) node()     {}
func (b *BreakStmt) stmtNode() {}

// NextStmt represents a next statement (continues to next iteration)
type NextStmt struct{}

func (n *NextStmt) node()     {}
func (n *NextStmt) stmtNode() {}

// ReturnStmt represents a return statement
type ReturnStmt struct {
	Values []Expression // empty if no return values
}

func (r *ReturnStmt) node()     {}
func (r *ReturnStmt) stmtNode() {}

// DeferStmt represents a defer statement
type DeferStmt struct {
	Call *CallExpr
}

func (d *DeferStmt) node()     {}
func (d *DeferStmt) stmtNode() {}

// ClassDecl represents a class definition
type ClassDecl struct {
	Name    string        // class name (e.g., "User")
	Embeds  []string      // embedded types (Go struct embedding), empty if none
	Fields  []*FieldDecl  // fields inferred from initialize
	Methods []*MethodDecl // methods defined in class
	Pub     bool          // true if exported (pub class)
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
	Name        string      // method name (may end with ! for pointer receiver)
	Params      []*Param    // parameters
	ReturnTypes []string    // return types
	Body        []Statement // method body
	Pub         bool        // true if exported (pub def)
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
	Name  string     // variable name without @
	Value Expression
}

func (i *InstanceVarAssign) node()     {}
func (i *InstanceVarAssign) stmtNode() {}

// InstanceVarOrAssign represents @name ||= value
type InstanceVarOrAssign struct {
	Name  string     // variable name without @
	Value Expression
}

func (i *InstanceVarOrAssign) node()     {}
func (i *InstanceVarOrAssign) stmtNode() {}

// InterfaceDecl represents an interface definition
type InterfaceDecl struct {
	Name    string        // interface name (e.g., "Speaker")
	Methods []*MethodSig  // method signatures (no body)
	Pub     bool          // true if exported (pub interface)
}

func (i *InterfaceDecl) node()     {}
func (i *InterfaceDecl) stmtNode() {}

// MethodSig represents a method signature in an interface (no body)
type MethodSig struct {
	Name        string   // method name
	Params      []*Param // parameters
	ReturnTypes []string // return types
}
