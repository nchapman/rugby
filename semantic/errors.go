package semantic

import "fmt"

// Error represents a semantic analysis error with position information.
type Error struct {
	Line    int
	Column  int
	Message string
}

func (e *Error) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, e.Message)
	}
	return e.Message
}

// UndefinedError is returned when a symbol is not defined.
type UndefinedError struct {
	Name       string
	Line       int
	Column     int
	Candidates []string // possible corrections (for "Did you mean?")
}

func (e *UndefinedError) Error() string {
	msg := fmt.Sprintf("undefined: '%s'", e.Name)
	if len(e.Candidates) > 0 {
		msg += fmt.Sprintf(" (did you mean '%s'?)", e.Candidates[0])
	}
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// TypeMismatchError is returned when types don't match.
type TypeMismatchError struct {
	Expected *Type
	Got      *Type
	Line     int
	Column   int
	Context  string // e.g., "assignment", "argument", "return"
}

func (e *TypeMismatchError) Error() string {
	msg := fmt.Sprintf("type mismatch in %s: expected %s, got %s", e.Context, e.Expected, e.Got)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// ArityMismatchError is returned when argument count doesn't match parameter count.
type ArityMismatchError struct {
	Name     string
	Expected int
	Got      int
	Line     int
	Column   int
}

func (e *ArityMismatchError) Error() string {
	msg := fmt.Sprintf("wrong number of arguments for '%s': expected %d, got %d", e.Name, e.Expected, e.Got)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// ReturnOutsideFunctionError is returned for return statements outside functions.
type ReturnOutsideFunctionError struct {
	Line   int
	Column int
}

func (e *ReturnOutsideFunctionError) Error() string {
	msg := "return statement outside function"
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// BreakOutsideLoopError is returned for break statements outside loops.
type BreakOutsideLoopError struct {
	Line   int
	Column int
}

func (e *BreakOutsideLoopError) Error() string {
	msg := "break statement outside loop"
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// SelfOutsideClassError is returned when 'self' is used outside a class.
type SelfOutsideClassError struct {
	Line   int
	Column int
}

func (e *SelfOutsideClassError) Error() string {
	msg := "'self' used outside class"
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// InstanceVarOutsideClassError is returned when instance variables are used outside a class.
type InstanceVarOutsideClassError struct {
	Name   string
	Line   int
	Column int
}

func (e *InstanceVarOutsideClassError) Error() string {
	msg := fmt.Sprintf("instance variable '@%s' used outside class", e.Name)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// OperatorTypeMismatchError is returned when operator operands have incompatible types.
type OperatorTypeMismatchError struct {
	Op        string
	LeftType  *Type
	RightType *Type
	Line      int
	Column    int
}

func (e *OperatorTypeMismatchError) Error() string {
	msg := fmt.Sprintf("cannot use '%s' with types %s and %s", e.Op, e.LeftType, e.RightType)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// UnaryTypeMismatchError reports a type error in a unary expression.
type UnaryTypeMismatchError struct {
	Op          string
	OperandType *Type
	Expected    string // e.g., "Bool", "numeric"
	Line        int
	Column      int
}

func (e *UnaryTypeMismatchError) Error() string {
	msg := fmt.Sprintf("cannot use '%s' with type %s (expected %s)", e.Op, e.OperandType, e.Expected)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}
