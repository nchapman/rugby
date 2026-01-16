package semantic

import (
	"fmt"
	"strings"
)

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

// CircularInheritanceError is returned when class inheritance creates a cycle.
type CircularInheritanceError struct {
	Cycle []string // class names in the cycle (e.g., ["A", "B", "A"])
	Line  int
}

func (e *CircularInheritanceError) Error() string {
	msg := fmt.Sprintf("circular inheritance detected: %s", formatCycle(e.Cycle))
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// formatCycle formats a cycle path like "A -> B -> A"
func formatCycle(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}
	return strings.Join(cycle, " -> ")
}

// CircularTypeAliasError is returned when type aliases form a cycle.
type CircularTypeAliasError struct {
	Name string
	Line int
}

func (e *CircularTypeAliasError) Error() string {
	msg := fmt.Sprintf("circular type alias: '%s' references itself", e.Name)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// DuplicateTypeAliasError is returned when a type alias is defined twice.
type DuplicateTypeAliasError struct {
	Name string
	Line int
}

func (e *DuplicateTypeAliasError) Error() string {
	msg := fmt.Sprintf("type alias '%s' already defined", e.Name)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// TypeAliasConflictError is returned when a type alias conflicts with an existing type.
type TypeAliasConflictError struct {
	Name     string
	Conflict string // "class", "interface", "module"
	Line     int
}

func (e *TypeAliasConflictError) Error() string {
	msg := fmt.Sprintf("type alias '%s' conflicts with existing %s", e.Name, e.Conflict)
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

// ReturnInsideBlockError is returned for return statements inside iterator blocks.
type ReturnInsideBlockError struct {
	Line   int
	Column int
}

func (e *ReturnInsideBlockError) Error() string {
	msg := "return statement not allowed inside block (blocks are functional - use find or select instead)"
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

// BreakInsideBlockError is returned for break statements inside iterator blocks.
type BreakInsideBlockError struct {
	Line   int
	Column int
}

func (e *BreakInsideBlockError) Error() string {
	msg := "break statement not allowed inside block (use find or take_while for early exit)"
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// NextInsideBlockError is returned for next statements inside iterator blocks.
type NextInsideBlockError struct {
	Line   int
	Column int
}

func (e *NextInsideBlockError) Error() string {
	msg := "next statement not allowed inside block (use select or reject for filtering)"
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// NextOutsideLoopError is returned for next statements outside loops.
type NextOutsideLoopError struct {
	Line   int
	Column int
}

func (e *NextOutsideLoopError) Error() string {
	msg := "next statement outside loop"
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

// StructFieldMutationError is returned when trying to mutate a struct field (structs are immutable).
type StructFieldMutationError struct {
	Field      string
	StructName string
	Line       int
	Column     int
}

func (e *StructFieldMutationError) Error() string {
	msg := fmt.Sprintf("cannot modify struct field '@%s' (structs are immutable)", e.Field)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// InstanceStateInClassMethodError is returned when instance state (self, @var) is accessed in a class method.
type InstanceStateInClassMethodError struct {
	What   string // "self" or "@varname"
	Line   int
	Column int
}

func (e *InstanceStateInClassMethodError) Error() string {
	msg := fmt.Sprintf("cannot access %s in class method (class methods have no instance)", e.What)
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

// ReturnTypeMismatchError reports when a returned value doesn't match the declared return type.
type ReturnTypeMismatchError struct {
	Expected *Type
	Got      *Type
	Line     int
}

func (e *ReturnTypeMismatchError) Error() string {
	msg := fmt.Sprintf("cannot return %s as %s", e.Got, e.Expected)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// ArgumentTypeMismatchError reports when an argument type doesn't match the parameter type.
type ArgumentTypeMismatchError struct {
	FuncName  string
	ParamName string
	Expected  *Type
	Got       *Type
	ArgIndex  int
	Line      int
}

func (e *ArgumentTypeMismatchError) Error() string {
	msg := fmt.Sprintf("cannot use %s as %s in argument to %s", e.Got, e.Expected, e.FuncName)
	if e.ParamName != "" {
		msg = fmt.Sprintf("cannot use %s as %s for parameter '%s' in %s", e.Got, e.Expected, e.ParamName, e.FuncName)
	}
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// ConditionTypeMismatchError reports when a condition expression is not Bool.
type ConditionTypeMismatchError struct {
	Got     *Type
	Context string // "if", "while", "until"
	Line    int
}

func (e *ConditionTypeMismatchError) Error() string {
	msg := fmt.Sprintf("%s condition must be Bool, got %s", e.Context, e.Got)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// IndexTypeMismatchError reports when an array/string index is not Int.
type IndexTypeMismatchError struct {
	Got  *Type
	Line int
}

func (e *IndexTypeMismatchError) Error() string {
	msg := fmt.Sprintf("array/string index must be Int, got %s", e.Got)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// InterfaceNotImplementedError reports when a class doesn't implement an interface.
type InterfaceNotImplementedError struct {
	ClassName     string
	InterfaceName string
	MissingMethod string
	Line          int
}

func (e *InterfaceNotImplementedError) Error() string {
	msg := fmt.Sprintf("class %s does not implement interface %s: missing method '%s'",
		e.ClassName, e.InterfaceName, e.MissingMethod)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// MethodSignatureMismatchError reports when a method signature doesn't match interface.
type MethodSignatureMismatchError struct {
	ClassName     string
	InterfaceName string
	MethodName    string
	Expected      string
	Got           string
	Line          int
}

func (e *MethodSignatureMismatchError) Error() string {
	msg := fmt.Sprintf("class %s method '%s' has wrong signature for interface %s: expected %s, got %s",
		e.ClassName, e.MethodName, e.InterfaceName, e.Expected, e.Got)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// BangOnNonErrorError reports when ! is used on a non-error expression.
type BangOnNonErrorError struct {
	Got  *Type
	Line int
}

func (e *BangOnNonErrorError) Error() string {
	msg := fmt.Sprintf("cannot use '!' on expression of type %s (expected (T, error) tuple)", e.Got)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// BangOutsideErrorFunctionError reports when ! is used in a function that doesn't return error.
type BangOutsideErrorFunctionError struct {
	Line int
}

func (e *BangOutsideErrorFunctionError) Error() string {
	msg := "'!' can only be used in functions that return error"
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// MethodRequiresArgumentsError reports when a method requiring arguments is called without them.
type MethodRequiresArgumentsError struct {
	MethodName string
	ParamCount int
	Line       int
}

func (e *MethodRequiresArgumentsError) Error() string {
	msg := fmt.Sprintf("method '%s' requires %d argument(s)", e.MethodName, e.ParamCount)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// ConflictingVisibilityError reports when a method has both pub and private modifiers.
type ConflictingVisibilityError struct {
	MethodName string
	Line       int
}

func (e *ConflictingVisibilityError) Error() string {
	msg := fmt.Sprintf("method '%s' cannot be both 'pub' and 'private'", e.MethodName)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}

// PrivateMethodAccessError reports when a private method is called from outside its class.
type PrivateMethodAccessError struct {
	MethodName string
	ClassName  string
	Line       int
}

func (e *PrivateMethodAccessError) Error() string {
	msg := fmt.Sprintf("cannot access private method '%s' from outside class '%s'", e.MethodName, e.ClassName)
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, msg)
	}
	return msg
}
