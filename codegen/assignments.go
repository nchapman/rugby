package codegen

// Assignment statement code generation.
// This file handles all forms of assignment:
//   - Simple assignment (x = expr)
//   - Compound assignment (x += expr, x -= expr, etc.)
//   - Or-assignment (x ||= expr)
//   - Multi-assignment and destructuring (a, b = expr)
//   - Instance variable assignment (@x = expr)
//   - Class variable assignment (@@x = expr)
//   - Selector assignment (obj.field = expr)
//   - Index assignment (arr[i] = expr)
//   - Bang and rescue assignment patterns

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
)

func (g *Generator) genInstanceVarAssign(s *ast.InstanceVarAssign) {
	g.writeIndent()
	if g.currentClass() != "" {
		recv := receiverName(g.currentClass())
		// Use underscore prefix for accessor fields to match struct definition
		goFieldName := s.Name
		if g.isAccessorField(s.Name) {
			goFieldName = "_" + s.Name
		}

		// Check if this field is from a parent class (not defined in current class)
		// If so, navigate through the embedded struct hierarchy
		fieldPath := g.getInheritedFieldPath(s.Name)
		if fieldPath != "" {
			g.buf.WriteString(fmt.Sprintf("%s.%s = ", recv, fieldPath))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s.%s = ", recv, goFieldName))
		}

		// Check for optional wrapping - need to take address for value type optionals
		// (e.g., @count : Int? with @count = 5 -> c.count = func() *int { v := 5; return &v }())
		fieldType := g.getFieldType(s.Name)
		sourceType := g.inferTypeFromExpr(s.Value)
		needsWrap := false

		if isValueTypeOptional(fieldType) {
			baseType := strings.TrimSuffix(fieldType, "?")
			if sourceType == baseType || sourceType == "" {
				needsWrap = true
			}
		}

		if needsWrap {
			// For value type optionals (Int?, Bool?, etc.), we use pointer types (*int, *bool)
			// Take the address of the value using an IIFE to handle non-addressable expressions
			goType := mapType(strings.TrimSuffix(fieldType, "?"))
			g.buf.WriteString(fmt.Sprintf("func() *%s { _v := ", goType))
			g.genExpr(s.Value)
			g.buf.WriteString("; return &_v }()")
		} else {
			g.genExpr(s.Value)
		}
	} else {
		g.addError(fmt.Errorf("instance variable '@%s' assigned outside of class context", s.Name))
		// Emit value expression to keep output somewhat valid
		g.genExpr(s.Value)
	}
	g.buf.WriteString("\n")
}

func (g *Generator) genOrAssignStmt(s *ast.OrAssignStmt) {
	if g.shouldDeclare(s) {
		// First declaration: just use :=
		g.writeIndent()
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" := ")
		g.genExpr(s.Value)
		g.buf.WriteString("\n")
		g.vars[s.Name] = "" // type unknown
	} else {
		// Variable exists: generate check
		g.writeIndent()
		g.buf.WriteString("if ")

		declaredType := g.vars[s.Name]
		// Check if nil
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" == nil")

		g.buf.WriteString(" {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" = ")

		// Check for optional wrapping
		sourceType := g.inferTypeFromExpr(s.Value)
		needsWrap := false
		if isValueTypeOptional(declaredType) {
			baseType := strings.TrimSuffix(declaredType, "?")
			if sourceType == baseType {
				needsWrap = true
			}
		}

		if needsWrap {
			g.buf.WriteString("runtime.Some(")
			g.genExpr(s.Value)
			g.buf.WriteString(")")
		} else {
			g.genExpr(s.Value)
		}

		g.buf.WriteString("\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

func (g *Generator) genCompoundAssignStmt(s *ast.CompoundAssignStmt) {
	g.writeIndent()
	g.buf.WriteString(s.Name)
	g.buf.WriteString(" ")
	g.buf.WriteString(s.Op)
	g.buf.WriteString("= ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")

	// Subtraction and modulo can make a value negative, invalidate tracking
	if s.Op == "-" || s.Op == "%" {
		delete(g.nonNegativeVars, s.Name)
	}
}

// genMultiAssignStmt generates code for tuple unpacking: val, ok = expr
// Also handles splat patterns: first, *rest = items or *head, last = items
func (g *Generator) genMultiAssignStmt(s *ast.MultiAssignStmt) {
	// Check for splat destructuring
	if s.SplatIndex >= 0 {
		g.genSplatDestructuring(s)
		return
	}

	// Check if the value expression returns an optional type (T?)
	// If so, we need special handling for the comma-ok pattern
	valueType := g.inferTypeFromExpr(s.Value)
	if len(s.Names) == 2 && isOptionalType(valueType) {
		g.genMultiAssignFromOptional(s, valueType)
		return
	}

	// Determine effective names (replace unused variables with _)
	effectiveNames := make([]string, len(s.Names))
	for i, name := range s.Names {
		if name == "_" {
			effectiveNames[i] = "_"
		} else if !g.typeInfo.IsVariableUsedAt(s, name) {
			// Variable is declared but never used - replace with _
			effectiveNames[i] = "_"
		} else {
			effectiveNames[i] = name
		}
	}

	// Use semantic analysis to determine if this is a declaration
	isDeclaration := g.shouldDeclare(s)

	// Check if the value is a tuple type (stored as a struct)
	// If so, we need to destructure it using field access

	// Check if value is an identifier with known tuple type
	if ident, ok := s.Value.(*ast.Ident); ok {
		if varType, exists := g.vars[ident.Name]; exists && isTupleType(varType) {
			g.genTupleDestructure(effectiveNames, ident.Name, isDeclaration, len(s.Names))
			return
		}
	}

	// Check if value expression has a tuple type (e.g., arr[i] where arr is Array<(T1, T2)>)
	// Only use tuple destructuring for non-call expressions - function calls return
	// multiple values directly, while array indexing returns a tuple struct.
	if isTupleType(valueType) {
		// For identifiers, only use tuple destructuring if they're known variables.
		// Unknown identifiers might be function calls without parentheses, which
		// return multiple values and should use standard multi-assignment.
		if ident, ok := s.Value.(*ast.Ident); ok {
			if _, isVar := g.vars[ident.Name]; isVar {
				g.genTupleDestructure(effectiveNames, ident.Name, isDeclaration, len(s.Names))
				return
			}
			// Not a known variable - might be a function call, fall through to
			// standard multi-assignment
		} else if _, isCall := s.Value.(*ast.CallExpr); isCall {
			// Function calls return multiple values directly in Go, not structs.
			// Fall through to standard multi-assignment.
		} else {
			// For complex expressions (like arr[i]), use a temporary variable
			// to avoid evaluating the expression multiple times
			g.genTupleDestructureFromExpr(s, effectiveNames, isDeclaration)
			return
		}
	}

	// Standard multi-assignment (e.g., from function returning multiple values)
	g.writeIndent()
	for i, name := range effectiveNames {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(name)
	}

	if isDeclaration {
		g.buf.WriteString(" := ")
	} else {
		g.buf.WriteString(" = ")
	}

	g.genExpr(s.Value)
	g.buf.WriteString("\n")

	// Track Go interop for multi-value assignments
	// If the value is a Go interop call (e.g., os.open(filename)), the first
	// variable receives a Go type and should be tracked as a Go interop var.
	if g.isGoInteropCall(s.Value) && len(effectiveNames) > 0 && effectiveNames[0] != "_" {
		g.goInteropVars[effectiveNames[0]] = true
	}
}

// genTupleDestructure generates code to destructure a tuple variable
// Example: a, b := tuple._0, tuple._1
func (g *Generator) genTupleDestructure(names []string, varName string, isDeclaration bool, count int) {
	g.writeIndent()
	for i, name := range names {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(name)
	}

	// For tuple destructuring, check if any non-blank variable is new
	// If so, we need to use :=
	if !isDeclaration {
		for _, name := range names {
			if name != "_" {
				if _, exists := g.vars[name]; !exists {
					isDeclaration = true
					break
				}
			}
		}
	}

	if isDeclaration {
		g.buf.WriteString(" := ")
		// Track the new variables
		for _, name := range names {
			if name != "_" {
				g.vars[name] = "" // type will be inferred
			}
		}
	} else {
		g.buf.WriteString(" = ")
	}

	for i := range count {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(fmt.Sprintf("%s._%d", varName, i))
	}
	g.buf.WriteString("\n")
}

// genTupleDestructureFromExpr generates code to destructure a tuple from an expression
// Example: _tmp := expr; a, b := _tmp._0, _tmp._1
func (g *Generator) genTupleDestructureFromExpr(s *ast.MultiAssignStmt, names []string, isDeclaration bool) {
	// First, store the expression in a temporary variable
	tempVar := fmt.Sprintf("_tuple%d", g.tempVarCounter)
	g.tempVarCounter++

	g.writeIndent()
	g.buf.WriteString(tempVar)
	g.buf.WriteString(" := ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")

	// Then destructure from the temporary variable
	g.genTupleDestructure(names, tempVar, isDeclaration, len(s.Names))
}

// genSplatDestructuring handles splat patterns: first, *rest = items or *head, last = items
// Generates individual assignments using slicing.
func (g *Generator) genSplatDestructuring(s *ast.MultiAssignStmt) {
	splatIdx := s.SplatIndex
	numBefore := splatIdx                   // elements before splat
	numAfter := len(s.Names) - splatIdx - 1 // elements after splat

	// Generate: _arr := expr (store in temp to avoid multiple evaluations)
	tempVar := fmt.Sprintf("_splat%d", g.tempVarCounter)
	g.tempVarCounter++
	g.writeIndent()
	g.buf.WriteString(tempVar)
	g.buf.WriteString(" := ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")

	// Generate assignments for elements before splat: first := _arr[0]
	for i := range numBefore {
		name := s.Names[i]
		if name == "_" {
			continue // skip unused
		}
		g.writeIndent()
		g.buf.WriteString(name)
		// Use = for existing variables, := for new ones
		if _, exists := g.vars[name]; exists {
			g.buf.WriteString(" = ")
		} else {
			g.buf.WriteString(" := ")
			g.vars[name] = "any"
		}
		g.buf.WriteString(fmt.Sprintf("%s[%d]", tempVar, i))
		g.buf.WriteString("\n")
	}

	// Generate splat assignment: rest := _arr[numBefore:len(_arr)-numAfter]
	splatName := s.Names[splatIdx]
	if splatName != "_" {
		g.writeIndent()
		g.buf.WriteString(splatName)
		// Use = for existing variables, := for new ones
		if _, exists := g.vars[splatName]; exists {
			g.buf.WriteString(" = ")
		} else {
			g.buf.WriteString(" := ")
			g.vars[splatName] = "any"
		}
		if numAfter == 0 {
			// Simple case: rest := _arr[numBefore:]
			g.buf.WriteString(fmt.Sprintf("%s[%d:]", tempVar, numBefore))
		} else {
			// Complex case: middle := _arr[numBefore:len(_arr)-numAfter]
			g.buf.WriteString(fmt.Sprintf("%s[%d:len(%s)-%d]", tempVar, numBefore, tempVar, numAfter))
		}
		g.buf.WriteString("\n")
	}

	// Generate assignments for elements after splat: last := _arr[len(_arr)-numAfter+i]
	for i := range numAfter {
		name := s.Names[splatIdx+1+i]
		if name == "_" {
			continue // skip unused
		}
		g.writeIndent()
		g.buf.WriteString(name)
		// Use = for existing variables, := for new ones
		if _, exists := g.vars[name]; exists {
			g.buf.WriteString(" = ")
		} else {
			g.buf.WriteString(" := ")
			g.vars[name] = "any"
		}
		if numAfter == 1 {
			// Simple case: last := _arr[len(_arr)-1]
			g.buf.WriteString(fmt.Sprintf("%s[len(%s)-1]", tempVar, tempVar))
		} else {
			// General case: elem := _arr[len(_arr)-numAfter+i]
			offset := numAfter - 1 - i
			if offset == 0 {
				g.buf.WriteString(fmt.Sprintf("%s[len(%s)-1]", tempVar, tempVar))
			} else {
				g.buf.WriteString(fmt.Sprintf("%s[len(%s)-%d]", tempVar, tempVar, offset+1))
			}
		}
		g.buf.WriteString("\n")
	}
}

// genMapDestructuringStmt generates code for map destructuring patterns:
//
//	{name:, age:} = user_data         -> name := user_data["name"]; age := user_data["age"]
//	{name: n, age: a} = user_data     -> n := user_data["name"]; a := user_data["age"]
func (g *Generator) genMapDestructuringStmt(s *ast.MapDestructuringStmt) {
	// Generate temp variable to hold the map
	tempVar := fmt.Sprintf("_map%d", g.tempVarCounter)
	g.tempVarCounter++

	g.writeIndent()
	g.buf.WriteString(tempVar)
	g.buf.WriteString(" := ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")

	// Generate assignment for each key-variable pair
	for _, pair := range s.Pairs {
		if pair.Variable == "_" {
			continue // skip blank identifier
		}
		g.writeIndent()
		g.buf.WriteString(pair.Variable)
		// Use = for existing variables, := for new ones
		if _, exists := g.vars[pair.Variable]; exists {
			g.buf.WriteString(" = ")
		} else {
			g.buf.WriteString(" := ")
			g.vars[pair.Variable] = "any"
		}
		g.buf.WriteString(tempVar)
		g.buf.WriteString("[\"")
		g.buf.WriteString(pair.Key)
		g.buf.WriteString("\"]\n")
	}
}

// genMultiAssignFromOptional handles tuple unpacking from optionals: val, ok = optional_expr
// Generates: _tmp := expr; ok := _tmp != nil; var val T; if ok { val = *_tmp }
func (g *Generator) genMultiAssignFromOptional(s *ast.MultiAssignStmt, optType string) {
	valName := s.Names[0]
	okName := s.Names[1]

	// Get the unwrapped type (remove the ?)
	unwrappedType := strings.TrimSuffix(optType, "?")
	goType := mapType(unwrappedType)

	// Use semantic analysis to determine if this is a declaration
	isDeclaration := g.shouldDeclare(s)

	// Generate temp variable assignment
	tempVar := fmt.Sprintf("_opt%d", g.tempVarCounter)
	g.tempVarCounter++

	g.writeIndent()
	g.buf.WriteString(tempVar)
	g.buf.WriteString(" := ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")

	// Generate ok assignment
	g.writeIndent()
	if isDeclaration && okName != "_" {
		g.buf.WriteString(okName)
		g.buf.WriteString(" := ")
	} else {
		g.buf.WriteString(okName)
		g.buf.WriteString(" = ")
	}
	g.buf.WriteString(tempVar)
	g.buf.WriteString(" != nil\n")

	// Generate val declaration with zero value
	if valName != "_" {
		g.writeIndent()
		if isDeclaration {
			g.buf.WriteString("var ")
			g.buf.WriteString(valName)
			g.buf.WriteString(" ")
			g.buf.WriteString(goType)
			g.buf.WriteString("\n")
		}

		// Generate conditional assignment
		g.writeIndent()
		g.buf.WriteString("if ")
		g.buf.WriteString(okName)
		g.buf.WriteString(" {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString(valName)
		g.buf.WriteString(" = *")
		g.buf.WriteString(tempVar)
		g.buf.WriteString("\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

func (g *Generator) genInstanceVarOrAssign(s *ast.InstanceVarOrAssign) {
	if g.currentClass() == "" {
		g.addError(fmt.Errorf("instance variable '@%s' ||= used outside of class context", s.Name))
		return
	}

	recv := receiverName(g.currentClass())
	// Use underscore prefix for accessor fields to match struct definition
	goFieldName := s.Name
	if g.isAccessorField(s.Name) {
		goFieldName = "_" + s.Name
	}
	field := fmt.Sprintf("%s.%s", recv, goFieldName)
	fieldType := g.getFieldType(s.Name)

	// Generate check
	g.writeIndent()
	g.buf.WriteString("if ")

	// Check if nil
	g.buf.WriteString(field)
	g.buf.WriteString(" == nil")

	g.buf.WriteString(" {\n")
	g.indent++
	g.writeIndent()
	g.buf.WriteString(field)
	g.buf.WriteString(" = ")

	// Check for optional wrapping - need to take address for value type optionals
	sourceType := g.inferTypeFromExpr(s.Value)
	needsWrap := false
	if isValueTypeOptional(fieldType) {
		baseType := strings.TrimSuffix(fieldType, "?")
		if sourceType == baseType || sourceType == "" {
			needsWrap = true
		}
	}

	if needsWrap {
		// For value type optionals (Int?, Bool?, etc.), we use pointer types (*int, *bool)
		// Take the address of the value using an IIFE to handle non-addressable expressions
		goType := mapType(strings.TrimSuffix(fieldType, "?"))
		g.buf.WriteString(fmt.Sprintf("func() *%s { _v := ", goType))
		g.genExpr(s.Value)
		g.buf.WriteString("; return &_v }()")
	} else {
		g.genExpr(s.Value)
	}

	g.buf.WriteString("\n")
	g.indent--
	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genInstanceVarCompoundAssign generates code for @name += value, @name -= value, etc.
// This generates: recv.field += value (using Go's native compound operators)
func (g *Generator) genInstanceVarCompoundAssign(s *ast.InstanceVarCompoundAssign) {
	if g.currentClass() == "" {
		g.addError(fmt.Errorf("instance variable '@%s' %s= used outside of class context", s.Name, s.Op))
		return
	}

	recv := receiverName(g.currentClass())
	// Use underscore prefix for accessor fields to match struct definition
	goFieldName := s.Name
	if g.isAccessorField(s.Name) {
		goFieldName = "_" + s.Name
	}
	field := fmt.Sprintf("%s.%s", recv, goFieldName)

	// Generate: field += value (using native compound operator)
	g.writeIndent()
	g.buf.WriteString(field)
	g.buf.WriteString(" ")
	g.buf.WriteString(s.Op)
	g.buf.WriteString("= ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genClassVarAssign(s *ast.ClassVarAssign) {
	if g.currentClass() == "" {
		g.addError(fmt.Errorf("class variable '@@%s' assignment used outside of class context", s.Name))
		return
	}

	// Class variables are package-level vars named _ClassName_varname
	varName := fmt.Sprintf("_%s_%s", g.currentClass(), s.Name)

	g.writeIndent()
	g.buf.WriteString(varName)
	g.buf.WriteString(" = ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genClassVarCompoundAssign(s *ast.ClassVarCompoundAssign) {
	if g.currentClass() == "" {
		g.addError(fmt.Errorf("class variable '@@%s' %s= used outside of class context", s.Name, s.Op))
		return
	}

	// Class variables are package-level vars named _ClassName_varname
	varName := fmt.Sprintf("_%s_%s", g.currentClass(), s.Name)

	// Generate: varName += value (using native compound operator)
	g.writeIndent()
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(s.Op)
	g.buf.WriteString("= ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

// genSelectorAssignStmt generates code for setter assignment: obj.field = value
// For declared accessors (setter/property), generates direct field access: obj._field = value
// For custom setters (def field=), generates method call: obj.setField(value)
func (g *Generator) genSelectorAssignStmt(s *ast.SelectorAssignStmt) {
	g.writeIndent()

	// Check if this is a declared accessor (setter/property) with a backing field
	receiverClassName := g.getReceiverClassName(s.Object)
	hasAccessor := g.typeInfo != nil && g.typeInfo.HasAccessor(receiverClassName, s.Field)

	if hasAccessor {
		// Declared accessor - use direct field access (inlined for performance)
		g.genExpr(s.Object)
		g.buf.WriteString(".")
		g.buf.WriteString(privateFieldName(s.Field))
		g.buf.WriteString(" = ")
		g.genExpr(s.Value)
		g.buf.WriteString("\n")
	} else {
		// Custom setter method - generate setter method call
		isPubClass := g.isPublicClass(receiverClassName)
		isPubAccessor := false
		if classAccessors := g.pubAccessors[receiverClassName]; classAccessors != nil {
			isPubAccessor = classAccessors[s.Field]
		}

		g.genExpr(s.Object)
		g.buf.WriteString(".")
		fieldPascal := snakeToPascalWithAcronyms(s.Field)
		if isPubClass || isPubAccessor {
			g.buf.WriteString("Set" + fieldPascal)
		} else {
			g.buf.WriteString("set" + fieldPascal)
		}
		g.buf.WriteString("(")
		g.genExpr(s.Value)
		g.buf.WriteString(")\n")
	}
}

// genSelectorCompoundAssign generates code for compound setter assignment: obj.field += value
// For declared accessors, generates direct field access: obj._field += value
// For custom setters, generates method call: obj.setField(obj.field() + value)
func (g *Generator) genSelectorCompoundAssign(s *ast.SelectorCompoundAssign) {
	g.writeIndent()

	// Check if this is a declared accessor (setter/property) with a backing field
	receiverClassName := g.getReceiverClassName(s.Object)
	hasAccessor := g.typeInfo != nil && g.typeInfo.HasAccessor(receiverClassName, s.Field)

	if hasAccessor {
		// Declared accessor - use direct field access (inlined for performance)
		g.genExpr(s.Object)
		g.buf.WriteString(".")
		g.buf.WriteString(privateFieldName(s.Field))
		g.buf.WriteString(" ")
		g.buf.WriteString(s.Op)
		g.buf.WriteString("= ")
		g.genExpr(s.Value)
		g.buf.WriteString("\n")
	} else {
		// Custom accessor - generate setter/getter method calls
		isPubClass := g.isPublicClass(receiverClassName)
		isPubAccessor := false
		if classAccessors := g.pubAccessors[receiverClassName]; classAccessors != nil {
			isPubAccessor = classAccessors[s.Field]
		}

		fieldPascal := snakeToPascalWithAcronyms(s.Field)
		var setterName, getterName string
		if isPubClass || isPubAccessor {
			setterName = "Set" + fieldPascal
			getterName = fieldPascal
		} else {
			setterName = "set" + fieldPascal
			getterName = snakeToCamelWithAcronyms(s.Field)
		}

		// Check if object expression needs to be stored in a temp variable
		// to avoid double evaluation (e.g., get_user().score += 1)
		objVar := "_obj"
		needsTempVar := !isSimpleExpr(s.Object)

		if needsTempVar {
			// Generate: _obj := expr
			g.buf.WriteString(objVar)
			g.buf.WriteString(" := ")
			g.genExpr(s.Object)
			g.buf.WriteString("\n")
			g.writeIndent()
		}

		// Generate: obj.setField(obj.field() op value)
		if needsTempVar {
			g.buf.WriteString(objVar)
		} else {
			g.genExpr(s.Object)
		}
		g.buf.WriteString(".")
		g.buf.WriteString(setterName)
		g.buf.WriteString("(")
		if needsTempVar {
			g.buf.WriteString(objVar)
		} else {
			g.genExpr(s.Object)
		}
		g.buf.WriteString(".")
		g.buf.WriteString(getterName)
		g.buf.WriteString("() ")
		g.buf.WriteString(s.Op)
		g.buf.WriteString(" ")
		g.genExpr(s.Value)
		g.buf.WriteString(")\n")
	}
}

// genIndexAssignStmt generates code for index assignment: arr[idx] = value or map[key] = value
func (g *Generator) genIndexAssignStmt(s *ast.IndexAssignStmt) {
	g.writeIndent()
	g.genExpr(s.Left)
	g.buf.WriteString("[")
	g.genExpr(s.Index)
	g.buf.WriteString("] = ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

// genIndexCompoundAssignStmt generates code for index compound assignment: arr[idx] += value
// This generates: arr[idx] += value (using Go's native compound operators)
func (g *Generator) genIndexCompoundAssignStmt(s *ast.IndexCompoundAssignStmt) {
	g.writeIndent()
	g.genExpr(s.Left)
	g.buf.WriteString("[")
	g.genExpr(s.Index)
	g.buf.WriteString("] ")
	g.buf.WriteString(s.Op)
	g.buf.WriteString("= ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genAssignStmt(s *ast.AssignStmt) {
	// Handle BangExpr: x = call()! -> x, _err := call(); if _err != nil { ... }
	if bangExpr, ok := s.Value.(*ast.BangExpr); ok {
		g.genBangAssign(s, bangExpr)
		return
	}

	// Handle RescueExpr: x = call() rescue default
	if rescueExpr, ok := s.Value.(*ast.RescueExpr); ok {
		g.genRescueAssign(s, rescueExpr)
		return
	}

	g.writeIndent()

	// Determine target type
	var targetType string
	if s.Type != "" {
		targetType = s.Type
	} else if declaredType, ok := g.vars[s.Name]; ok {
		targetType = declaredType
	}

	// Check if we are assigning nil to a value type optional
	isNilAssignment := false
	if _, ok := s.Value.(*ast.NilLit); ok {
		isNilAssignment = true
	}

	if isNilAssignment && isValueTypeOptional(targetType) {
		goType := mapType(targetType)

		if s.Type != "" && g.shouldDeclare(s) {
			// Typed declaration: var x *int = nil
			g.buf.WriteString(fmt.Sprintf("var %s %s = nil", s.Name, goType))
			g.vars[s.Name] = s.Type
		} else if !g.shouldDeclare(s) {
			// Reassignment: x = nil
			g.buf.WriteString(s.Name)
			g.buf.WriteString(" = nil")
		} else {
			// Untyped declaration - need explicit type since Go can't infer nil's type
			g.buf.WriteString(fmt.Sprintf("var %s %s = nil", s.Name, goType))
			g.vars[s.Name] = targetType
		}

		g.buf.WriteString("\n")
		return
	}

	// Check if we need to wrap value type -> Optional (e.g. x : Int? = 5)
	sourceType := g.inferTypeFromExpr(s.Value)
	needsWrap := false
	if isValueTypeOptional(targetType) {
		baseType := strings.TrimSuffix(targetType, "?")
		if sourceType == baseType {
			needsWrap = true
		}
	}

	if s.Type != "" && g.shouldDeclare(s) {
		// Typed declaration: var x int = value
		g.buf.WriteString(fmt.Sprintf("var %s %s = ", s.Name, g.mapTypeClassAware(s.Type)))
		g.vars[s.Name] = s.Type // store the declared type
	} else if s.Name == "_" || !g.shouldDeclare(s) {
		// Blank identifier or reassignment: x = value
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" = ")
	} else {
		// Untyped declaration: x := value
		g.buf.WriteString(s.Name)
		g.buf.WriteString(" := ")
		g.vars[s.Name] = sourceType // infer type from value if variable is new
	}

	// Track Go interop type variables for method call snake_case -> PascalCase conversion
	isGoInteropCtor := g.isGoInteropTypeConstructor(s.Value)
	isGoInteropCallVal := g.isGoInteropCall(s.Value)
	if isGoInteropCtor || isGoInteropCallVal {
		g.goInteropVars[s.Name] = true
	} else if g.goInteropVars[s.Name] {
		// Variable was reassigned to non-Go-interop value, remove tracking
		delete(g.goInteropVars, s.Name)
	}

	if needsWrap {
		g.buf.WriteString("runtime.Some(")
		g.genExpr(s.Value)
		g.buf.WriteString(")")
	} else if goType := g.getShiftLeftTargetType(s.Value); goType != "" {
		// ShiftLeft returns any, so we need a type assertion for typed slice assignment
		g.genExpr(s.Value)
		g.buf.WriteString(".(")
		g.buf.WriteString(goType)
		g.buf.WriteString(")")
	} else if tupleLit, ok := s.Value.(*ast.TupleLit); ok && isTupleType(targetType) {
		// Tuple literal assigned to tuple type: generate struct literal
		g.genTupleLitAsStruct(tupleLit, targetType)
	} else {
		g.genExpr(s.Value)
	}
	g.buf.WriteString("\n")

	// Track non-negative variables for native array indexing optimization
	g.trackNonNegativeAssign(s.Name, s.Value)
}

// trackNonNegativeAssign marks a variable as non-negative if the assigned value is provably >= 0.
func (g *Generator) trackNonNegativeAssign(name string, value ast.Expression) {
	if g.isNonNegativeValue(value) {
		g.nonNegativeVars[name] = true
	} else {
		// Variable might be negative, remove from tracking
		delete(g.nonNegativeVars, name)
	}
}

// isNonNegativeValue checks if an expression is known to produce a non-negative value.
func (g *Generator) isNonNegativeValue(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.IntLit:
		return e.Value >= 0
	case *ast.Ident:
		// Check if this variable is tracked as non-negative
		return g.nonNegativeVars[e.Name]
	case *ast.BinaryExpr:
		// Addition and multiplication of non-negative values are non-negative
		if e.Op == "+" || e.Op == "*" {
			return g.isNonNegativeValue(e.Left) && g.isNonNegativeValue(e.Right)
		}
		return false
	case *ast.CallExpr:
		// len(x) is always non-negative
		if fn, ok := e.Func.(*ast.Ident); ok && fn.Name == "len" {
			return true
		}
		// .length is also always non-negative
		if sel, ok := e.Func.(*ast.SelectorExpr); ok && sel.Sel == "length" {
			return true
		}
		return false
	default:
		return false
	}
}

// getShiftLeftTargetType returns the Go type for ShiftLeft type assertion if needed.
// Returns empty string if no assertion is needed.
// Note: When the slice type is known, genBinaryExpr uses runtime.Append which is generic,
// so no type assertion is needed. This function only returns a type for the fallback
// ShiftLeft case with unknown types.
func (g *Generator) getShiftLeftTargetType(expr ast.Expression) string {
	binExpr, ok := expr.(*ast.BinaryExpr)
	if !ok || binExpr.Op != "<<" {
		return ""
	}

	// Check if this is a channel - channels use native Go send syntax, no assertion needed
	if g.typeInfo.GetTypeKind(binExpr.Left) == TypeChannel {
		return ""
	}

	// Check if we can infer the slice type - if so, genBinaryExpr uses Append (generic)
	if elemType := g.inferArrayElementGoType(binExpr.Left); elemType != "" && elemType != "any" {
		// Append is generic and returns the correct type, no assertion needed
		return ""
	}

	// Fallback: ShiftLeft returns any, so we might need a type assertion
	// Try to get the type from typeInfo
	goType := g.typeInfo.GetGoType(binExpr.Left)
	if goType != "" {
		if strings.HasPrefix(goType, "[]") {
			return goType
		}
		if strings.HasPrefix(goType, "chan ") {
			return goType
		}
	}

	return ""
}

// genBangAssign generates code for: x = call()!
// Produces: x, _err0 := call(); if _err0 != nil { return/Fatal }
func (g *Generator) genBangAssign(s *ast.AssignStmt, bang *ast.BangExpr) {
	call, ok := bang.Expr.(*ast.CallExpr)
	if !ok {
		g.addError(fmt.Errorf("bang expression must be a call expression"))
		return
	}

	name := s.Name

	// Use unique error variable name to avoid conflicts
	errVar := fmt.Sprintf("_err%d", g.tempVarCounter)
	g.tempVarCounter++

	// Generate: name, _errN := call()
	g.writeIndent()
	if !g.shouldDeclare(s) {
		// Reassignment needs a temp for error
		g.buf.WriteString(fmt.Sprintf("%s, %s = ", name, errVar))
	} else {
		g.buf.WriteString(fmt.Sprintf("%s, %s := ", name, errVar))
		g.vars[name] = "" // mark as declared (type unknown for now)
	}
	g.genCallExpr(call)
	g.buf.WriteString("\n")

	// Track Go interop type variables for method call snake_case -> PascalCase conversion
	if g.isGoInteropCall(bang) {
		g.goInteropVars[name] = true
	}

	// Generate error check
	g.genBangErrorCheckWithVar(errVar)
}

// genBangStmt generates code for a standalone: call()!
// Produces: if _err := call(); _err != nil { return/Fatal }
func (g *Generator) genBangStmt(bang *ast.BangExpr) {
	call, ok := bang.Expr.(*ast.CallExpr)
	if !ok {
		g.addError(fmt.Errorf("bang expression must be a call expression"))
		return
	}

	// Generate: if _err := call(); _err != nil { ... }
	// Using a scoped variable in the if statement, so no unique name needed
	g.writeIndent()
	g.buf.WriteString("if _err := ")
	g.genCallExpr(call)
	g.buf.WriteString("; _err != nil {\n")
	g.indent++
	g.genBangErrorBodyWithVar("_err")
	g.indent--
	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genRescueAssign generates code for: x = call() rescue default
// Produces: x, _err0 := call(); if _err0 != nil { x = default }
func (g *Generator) genRescueAssign(s *ast.AssignStmt, rescue *ast.RescueExpr) {
	call, ok := rescue.Expr.(*ast.CallExpr)
	if !ok {
		g.addError(fmt.Errorf("rescue expression must be a call expression"))
		return
	}

	name := s.Name

	// Use unique error variable name to avoid conflicts
	errVar := fmt.Sprintf("_err%d", g.tempVarCounter)
	g.tempVarCounter++

	// Generate: name, _errN := call()
	g.writeIndent()
	if !g.shouldDeclare(s) {
		g.buf.WriteString(fmt.Sprintf("%s, %s = ", name, errVar))
	} else {
		g.buf.WriteString(fmt.Sprintf("%s, %s := ", name, errVar))
		g.vars[name] = "" // mark as declared
	}
	g.genCallExpr(call)
	g.buf.WriteString("\n")

	// Generate error handling
	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("if %s != nil {\n", errVar))
	g.indent++

	if rescue.Block != nil {
		// Block form
		if rescue.ErrName != "" {
			// Bind error to variable
			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("%s := %s\n", rescue.ErrName, errVar))
		}
		// Generate block body
		for i, stmt := range rescue.Block.Body {
			isLast := i == len(rescue.Block.Body)-1
			if isLast {
				// Last expression becomes the value
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					g.writeIndent()
					g.buf.WriteString(fmt.Sprintf("%s = ", name))
					g.genExpr(exprStmt.Expr)
					g.buf.WriteString("\n")
					continue
				}
			}
			g.genStatement(stmt)
		}
	} else {
		// Inline form: name = default
		g.writeIndent()
		g.buf.WriteString(fmt.Sprintf("%s = ", name))
		g.genExpr(rescue.Default)
		g.buf.WriteString("\n")
	}

	g.indent--
	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genRescueStmt generates code for a standalone: call() rescue do ... end
// Used when rescue appears in statement context
func (g *Generator) genRescueStmt(rescue *ast.RescueExpr) {
	call, ok := rescue.Expr.(*ast.CallExpr)
	if !ok {
		g.addError(fmt.Errorf("rescue expression must be a call expression"))
		return
	}

	// For statement context, we just need to handle the error
	// Using a scoped variable in the if statement, so no unique name needed
	g.writeIndent()
	g.buf.WriteString("if _err := ")
	g.genCallExpr(call)
	g.buf.WriteString("; _err != nil {\n")
	g.indent++

	if rescue.Block != nil {
		if rescue.ErrName != "" {
			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("%s := _err\n", rescue.ErrName))
		}
		for _, stmt := range rescue.Block.Body {
			g.genStatement(stmt)
		}
	} else if rescue.Default != nil {
		// Inline default in statement context - just evaluate it
		g.writeIndent()
		g.genExpr(rescue.Default)
		g.buf.WriteString("\n")
	}

	g.indent--
	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genBangErrorCheckWithVar generates the if errVar != nil check after an assignment
func (g *Generator) genBangErrorCheckWithVar(errVar string) {
	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("if %s != nil {\n", errVar))
	g.indent++
	g.genBangErrorBodyWithVar(errVar)
	g.indent--
	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genBangErrorBodyWithVar generates the body of the error check using the specified error variable
func (g *Generator) genBangErrorBodyWithVar(errVar string) {
	g.writeIndent()
	if g.inMainFunc {
		// In main: call runtime.Fatal
		g.needsRuntime = true
		g.buf.WriteString(fmt.Sprintf("runtime.Fatal(%s)\n", errVar))
	} else if g.returnsError() {
		// In error-returning function: propagate the error
		g.buf.WriteString("return ")
		// Generate zero values for all return types except error
		for i := range len(g.currentReturnTypes()) - 1 {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.buf.WriteString(g.zeroValue(g.currentReturnTypes()[i]))
		}
		if len(g.currentReturnTypes()) > 1 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(errVar + "\n")
	} else {
		// Not in main and doesn't return error - emit panic as fallback
		g.addError(fmt.Errorf("'!' operator used in function that doesn't return error"))
		g.buf.WriteString(fmt.Sprintf("panic(%s)\n", errVar))
	}
}

// genCondition generates a condition expression with strict Bool type checking.
// Rugby requires conditions to be Bool type - no implicit truthiness.
// Use 'if let x = expr' for optional unwrapping, or explicit checks like 'x != nil'.
