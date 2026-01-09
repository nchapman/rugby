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
}

// FuncDecl represents a function definition
type FuncDecl struct {
	Name        string
	Params      []*Param
	ReturnTypes []string // empty if not specified
	Body        []Statement
}

func (f *FuncDecl) node()     {}
func (f *FuncDecl) stmtNode() {}

// ExprStmt wraps an expression as a statement
type ExprStmt struct {
	Expr Expression
}

func (e *ExprStmt) node()     {}
func (e *ExprStmt) stmtNode() {}

// CallExpr represents a function call
type CallExpr struct {
	Func Expression
	Args []Expression
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
	Value Expression
}

func (a *AssignStmt) node()     {}
func (a *AssignStmt) stmtNode() {}

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
