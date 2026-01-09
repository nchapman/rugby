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

// FuncDecl represents a function definition
type FuncDecl struct {
	Name string
	Body []Statement
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
	Func string
	Args []Expression
}

func (c *CallExpr) node()     {}
func (c *CallExpr) exprNode() {}

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
