package codegen

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
)

// getStatementLine returns the source line number for a statement, or 0 if unknown.
func getStatementLine(stmt ast.Statement) int {
	switch s := stmt.(type) {
	// Declarations
	case *ast.FuncDecl:
		return s.Line
	case *ast.ClassDecl:
		return s.Line
	case *ast.InterfaceDecl:
		return s.Line
	case *ast.ModuleDecl:
		return s.Line
	case *ast.TypeAliasDecl:
		return s.Line
	case *ast.ConstDecl:
		return s.Line
	case *ast.EnumDecl:
		return s.Line
	case *ast.StructDecl:
		return s.Line
	case *ast.MethodDecl:
		return s.Line
	case *ast.AccessorDecl:
		return s.Line
	case *ast.IncludeStmt:
		return s.Line

	// Assignments
	case *ast.AssignStmt:
		return s.Line
	case *ast.OrAssignStmt:
		return s.Line
	case *ast.CompoundAssignStmt:
		return s.Line
	case *ast.MultiAssignStmt:
		return s.Line
	case *ast.MapDestructuringStmt:
		return s.Line
	case *ast.SelectorAssignStmt:
		return s.Line
	case *ast.InstanceVarCompoundAssign:
		return s.Line
	case *ast.ClassVarAssign:
		return s.Line
	case *ast.ClassVarCompoundAssign:
		return s.Line

	// Expressions
	case *ast.ExprStmt:
		return s.Line

	// Control flow
	case *ast.IfStmt:
		return s.Line
	case *ast.CaseStmt:
		return s.Line
	case *ast.CaseTypeStmt:
		return s.Line
	case *ast.WhileStmt:
		return s.Line
	case *ast.UntilStmt:
		return s.Line
	case *ast.ForStmt:
		return s.Line

	// Jump statements
	case *ast.ReturnStmt:
		return s.Line
	case *ast.BreakStmt:
		return s.Line
	case *ast.NextStmt:
		return s.Line
	case *ast.PanicStmt:
		return s.Line
	case *ast.DeferStmt:
		return s.Line

	// Concurrency
	case *ast.GoStmt:
		return s.Line
	case *ast.ChanSendStmt:
		return s.Line
	case *ast.SelectStmt:
		return s.Line
	case *ast.ConcurrentlyStmt:
		return s.Line

	// Testing
	case *ast.DescribeStmt:
		return s.Line
	case *ast.ItStmt:
		return s.Line
	case *ast.TestStmt:
		return s.Line
	case *ast.TableStmt:
		return s.Line
	case *ast.BeforeStmt:
		return s.Line
	case *ast.AfterStmt:
		return s.Line

	default:
		return 0
	}
}

// genStatement dispatches to the appropriate generator for each statement type.
// Statement types are organized into logical groups:
//   - Declarations: functions, classes, interfaces, modules
//   - Assignments: simple, compound, multi-value, instance variables
//   - Control flow: conditionals, loops
//   - Jump statements: break, next, return, panic, defer
//   - Concurrency: go, select, channels, structured concurrency
//   - Testing: describe, it, test, table
func (g *Generator) genStatement(stmt ast.Statement) {
	// Emit line directive to map errors back to Rugby source
	if line := getStatementLine(stmt); line > 0 {
		g.emitLineDirective(line)
	}

	switch s := stmt.(type) {
	// Declarations
	case *ast.FuncDecl:
		g.genFuncDecl(s)
	case *ast.ClassDecl:
		g.genClassDecl(s)
	case *ast.InterfaceDecl:
		g.genInterfaceDecl(s)
	case *ast.ModuleDecl:
		g.genModuleDecl(s)
	case *ast.TypeAliasDecl:
		g.genTypeAliasDecl(s)
	case *ast.ConstDecl:
		g.genConstDecl(s)
	case *ast.EnumDecl:
		g.genEnumDecl(s)
	case *ast.StructDecl:
		g.genStructDecl(s)

	// Assignments
	case *ast.AssignStmt:
		g.genAssignStmt(s)
	case *ast.OrAssignStmt:
		g.genOrAssignStmt(s)
	case *ast.CompoundAssignStmt:
		g.genCompoundAssignStmt(s)
	case *ast.MultiAssignStmt:
		g.genMultiAssignStmt(s)
	case *ast.MapDestructuringStmt:
		g.genMapDestructuringStmt(s)
	case *ast.InstanceVarAssign:
		g.genInstanceVarAssign(s)
	case *ast.InstanceVarOrAssign:
		g.genInstanceVarOrAssign(s)
	case *ast.InstanceVarCompoundAssign:
		g.genInstanceVarCompoundAssign(s)
	case *ast.ClassVarAssign:
		g.genClassVarAssign(s)
	case *ast.ClassVarCompoundAssign:
		g.genClassVarCompoundAssign(s)
	case *ast.SelectorAssignStmt:
		g.genSelectorAssignStmt(s)

	// Expression statements
	case *ast.ExprStmt:
		g.genExprStmt(s)

	// Control flow: conditionals
	case *ast.IfStmt:
		g.genIfStmt(s)
	case *ast.CaseStmt:
		g.genCaseStmt(s)
	case *ast.CaseTypeStmt:
		g.genCaseTypeStmt(s)

	// Control flow: loops
	case *ast.WhileStmt:
		g.genWhileStmt(s)
	case *ast.UntilStmt:
		g.genUntilStmt(s)
	case *ast.ForStmt:
		g.genForStmt(s)
	case *ast.LoopStmt:
		g.genLoopStmt(s)

	// Jump statements
	case *ast.BreakStmt:
		g.genBreakStmt(s)
	case *ast.NextStmt:
		g.genNextStmt(s)
	case *ast.ReturnStmt:
		g.genReturnStmt(s)
	case *ast.PanicStmt:
		g.genPanicStmt(s)
	case *ast.DeferStmt:
		g.genDeferStmt(s)

	// Concurrency
	case *ast.GoStmt:
		g.genGoStmt(s)
	case *ast.SelectStmt:
		g.genSelectStmt(s)
	case *ast.ChanSendStmt:
		g.genChanSendStmt(s)
	case *ast.ConcurrentlyStmt:
		g.genConcurrentlyStmt(s)

	// Testing
	case *ast.DescribeStmt:
		g.genDescribeStmt(s)
	case *ast.ItStmt:
		g.genItStmt(s)
	case *ast.TestStmt:
		g.genTestStmt(s)
	case *ast.TableStmt:
		g.genTableStmt(s)

	default:
		g.addError(fmt.Errorf("unhandled statement type: %T", stmt))
	}
}

func (g *Generator) genInstanceVarAssign(s *ast.InstanceVarAssign) {
	g.writeIndent()
	if g.currentClass != "" {
		recv := receiverName(g.currentClass)
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
			baseType := strings.TrimSuffix(declaredType, "?")
			g.buf.WriteString(fmt.Sprintf("runtime.Some%s(", baseType))
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
	g.buf.WriteString(" = ")
	g.buf.WriteString(s.Name)
	g.buf.WriteString(" ")
	g.buf.WriteString(s.Op)
	g.buf.WriteString(" ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
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

	g.writeIndent()

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

	// Generate names
	for i, name := range effectiveNames {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(name)
	}

	// Use := if any variable is new, = if all are already declared
	if isDeclaration {
		g.buf.WriteString(" := ")
	} else {
		g.buf.WriteString(" = ")
	}

	g.genExpr(s.Value)
	g.buf.WriteString("\n")
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
		g.buf.WriteString(" := ")
		g.buf.WriteString(fmt.Sprintf("%s[%d]", tempVar, i))
		g.buf.WriteString("\n")
	}

	// Generate splat assignment: rest := _arr[numBefore:len(_arr)-numAfter]
	splatName := s.Names[splatIdx]
	if splatName != "_" {
		g.writeIndent()
		g.buf.WriteString(splatName)
		g.buf.WriteString(" := ")
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
		g.buf.WriteString(" := ")
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
		g.writeIndent()
		g.buf.WriteString(pair.Variable)
		g.buf.WriteString(" := ")
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
	if g.currentClass == "" {
		g.addError(fmt.Errorf("instance variable '@%s' ||= used outside of class context", s.Name))
		return
	}

	recv := receiverName(g.currentClass)
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
// This expands to: recv.field = recv.field + value
func (g *Generator) genInstanceVarCompoundAssign(s *ast.InstanceVarCompoundAssign) {
	if g.currentClass == "" {
		g.addError(fmt.Errorf("instance variable '@%s' %s= used outside of class context", s.Name, s.Op))
		return
	}

	recv := receiverName(g.currentClass)
	// Use underscore prefix for accessor fields to match struct definition
	goFieldName := s.Name
	if g.isAccessorField(s.Name) {
		goFieldName = "_" + s.Name
	}
	field := fmt.Sprintf("%s.%s", recv, goFieldName)

	// Generate: field = field op value
	g.writeIndent()
	g.buf.WriteString(field)
	g.buf.WriteString(" = ")
	g.buf.WriteString(field)
	g.buf.WriteString(" ")
	g.buf.WriteString(s.Op)
	g.buf.WriteString(" ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genClassVarAssign(s *ast.ClassVarAssign) {
	if g.currentClass == "" {
		g.addError(fmt.Errorf("class variable '@@%s' assignment used outside of class context", s.Name))
		return
	}

	// Class variables are package-level vars named _ClassName_varname
	varName := fmt.Sprintf("_%s_%s", g.currentClass, s.Name)

	g.writeIndent()
	g.buf.WriteString(varName)
	g.buf.WriteString(" = ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genClassVarCompoundAssign(s *ast.ClassVarCompoundAssign) {
	if g.currentClass == "" {
		g.addError(fmt.Errorf("class variable '@@%s' %s= used outside of class context", s.Name, s.Op))
		return
	}

	// Class variables are package-level vars named _ClassName_varname
	varName := fmt.Sprintf("_%s_%s", g.currentClass, s.Name)

	// Generate: varName = varName op value
	g.writeIndent()
	g.buf.WriteString(varName)
	g.buf.WriteString(" = ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" ")
	g.buf.WriteString(s.Op)
	g.buf.WriteString(" ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

// genSelectorAssignStmt generates code for setter assignment: obj.field = value
// This generates a setter method call: obj.setField(value) or obj.SetField(value) for pub classes/accessors
func (g *Generator) genSelectorAssignStmt(s *ast.SelectorAssignStmt) {
	g.writeIndent()

	// Check if receiver is an instance of a pub class
	receiverClassName := g.getReceiverClassName(s.Object)
	isPubClass := g.isPublicClass(receiverClassName)

	// Check if this specific accessor is pub (for non-pub classes with pub accessors)
	isPubAccessor := false
	if classAccessors := g.pubAccessors[receiverClassName]; classAccessors != nil {
		isPubAccessor = classAccessors[s.Field]
	}

	// Generate: obj.setField(value) or obj.SetField(value)
	g.genExpr(s.Object)
	g.buf.WriteString(".")

	// Generate setter method name
	fieldPascal := snakeToPascalWithAcronyms(s.Field)
	var setterName string
	if isPubClass || isPubAccessor {
		// Pub class/accessor setters use PascalCase: SetField
		setterName = "Set" + fieldPascal
	} else {
		// Private class setters use camelCase: setField
		setterName = "set" + fieldPascal
	}

	g.buf.WriteString(setterName)
	g.buf.WriteString("(")
	g.genExpr(s.Value)
	g.buf.WriteString(")\n")
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
		baseType := strings.TrimSuffix(targetType, "?")

		if s.Type != "" && g.shouldDeclare(s) {
			// Typed declaration: var x Int? = nil
			g.buf.WriteString(fmt.Sprintf("var %s %s = ", s.Name, mapType(s.Type)))
			g.vars[s.Name] = s.Type
		} else if !g.shouldDeclare(s) {
			// Reassignment: x = nil
			g.buf.WriteString(s.Name)
			g.buf.WriteString(" = ")
		} else {
			// Untyped declaration (x = nil)
			g.buf.WriteString(s.Name)
			g.buf.WriteString(" := ")
			g.vars[s.Name] = ""
		}

		g.buf.WriteString(fmt.Sprintf("runtime.None%s()", baseType))
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
		g.buf.WriteString(fmt.Sprintf("var %s %s = ", s.Name, mapType(s.Type)))
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
	if g.isGoInteropTypeConstructor(s.Value) || g.isGoInteropCall(s.Value) {
		g.goInteropVars[s.Name] = true
	} else if g.goInteropVars[s.Name] {
		// Variable was reassigned to non-Go-interop value, remove tracking
		delete(g.goInteropVars, s.Name)
	}

	if needsWrap {
		baseType := strings.TrimSuffix(targetType, "?")
		g.buf.WriteString(fmt.Sprintf("runtime.Some%s(", baseType))
		g.genExpr(s.Value)
		g.buf.WriteString(")")
	} else if goType := g.getShiftLeftTargetType(s.Value); goType != "" {
		// ShiftLeft returns any, so we need a type assertion for typed slice assignment
		g.genExpr(s.Value)
		g.buf.WriteString(".(")
		g.buf.WriteString(goType)
		g.buf.WriteString(")")
	} else {
		g.genExpr(s.Value)
	}
	g.buf.WriteString("\n")
}

// getShiftLeftTargetType returns the Go type for ShiftLeft type assertion if needed.
// Returns empty string if no assertion is needed.
func (g *Generator) getShiftLeftTargetType(expr ast.Expression) string {
	binExpr, ok := expr.(*ast.BinaryExpr)
	if !ok || binExpr.Op != "<<" {
		return ""
	}

	// Get the Go type of the left operand (the array/channel being appended to)
	goType := g.typeInfo.GetGoType(binExpr.Left)
	if goType != "" {
		// Handle slices
		if strings.HasPrefix(goType, "[]") {
			return goType
		}
		// Handle channels
		if strings.HasPrefix(goType, "chan ") {
			return goType
		}
	}

	// Fall back to inferring element type from left operand
	elemType := g.inferArrayElementGoType(binExpr.Left)
	if elemType != "any" && elemType != "" {
		return "[]" + elemType
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
		for i := range len(g.currentReturnTypes) - 1 {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.buf.WriteString(g.zeroValue(g.currentReturnTypes[i]))
		}
		if len(g.currentReturnTypes) > 1 {
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
func (g *Generator) genCondition(cond ast.Expression) error {
	condType := g.inferTypeFromExpr(cond)

	// Allow Bool type explicitly
	if condType == "Bool" {
		g.genExpr(cond)
		return nil
	}

	// Allow nil comparisons (err != nil, x == nil) - these return Bool
	if binary, ok := cond.(*ast.BinaryExpr); ok {
		if binary.Op == "==" || binary.Op == "!=" {
			if _, isNil := binary.Right.(*ast.NilLit); isNil {
				g.genExpr(cond)
				return nil
			}
			if _, isNil := binary.Left.(*ast.NilLit); isNil {
				g.genExpr(cond)
				return nil
			}
		}
	}

	// If type is unknown, generate as-is (may be external function returning bool)
	if condType == "" {
		g.genExpr(cond)
		return nil
	}

	// Optional types require explicit unwrapping
	if isOptionalType(condType) {
		return fmt.Errorf("condition must be Bool, got %s (use 'if let x = ...' or 'x != nil' for optionals)", condType)
	}

	// Non-Bool types are errors
	return fmt.Errorf("condition must be Bool, got %s", condType)
}

func (g *Generator) genIfStmt(s *ast.IfStmt) {
	g.writeIndent()
	g.buf.WriteString("if ")

	// Handle assignment-in-condition pattern: if let n = expr
	// For pointer-based optionals (T?): if _tmp := expr; _tmp != nil { n := *_tmp; ... }
	// For tuple-returning functions: if n, ok := expr; ok { ... }
	if s.AssignName != "" {
		exprType := g.inferTypeFromExpr(s.AssignExpr)

		if isOptionalType(exprType) {
			// Pointer-based optional: generate temp var and nil check
			tempVar := fmt.Sprintf("_iflet%d", g.tempVarCounter)
			g.tempVarCounter++
			g.buf.WriteString(tempVar)
			g.buf.WriteString(" := ")
			g.genExpr(s.AssignExpr)
			g.buf.WriteString("; ")
			g.buf.WriteString(tempVar)
			g.buf.WriteString(" != nil {\n")

			// Unwrap the pointer into the user's variable
			g.indent++
			g.writeIndent()
			g.buf.WriteString(s.AssignName)
			g.buf.WriteString(" := *")
			g.buf.WriteString(tempVar)
			g.buf.WriteString("\n")
			// Silence "declared and not used" error if variable isn't referenced
			g.writeIndent()
			g.buf.WriteString("_ = ")
			g.buf.WriteString(s.AssignName)
			g.buf.WriteString("\n")
			g.indent--

			// Track the variable with its unwrapped type
			unwrappedType := strings.TrimSuffix(exprType, "?")
			g.vars[s.AssignName] = unwrappedType
		} else {
			// Tuple-returning function: if n, ok := expr; ok { ... }
			g.buf.WriteString(s.AssignName)
			g.buf.WriteString(", ok := ")
			g.genExpr(s.AssignExpr)
			g.buf.WriteString("; ok {\n")
			// Track the variable with its inferred type
			g.vars[s.AssignName] = exprType
		}
	} else {
		// For unless, negate the condition
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Cond); err != nil {
			g.addError(fmt.Errorf("line %d: %w", s.Line, err))
			// Generate a placeholder to keep the output somewhat valid
			g.buf.WriteString("false")
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
	}

	g.indent++
	for _, stmt := range s.Then {
		g.genStatement(stmt)
	}
	g.indent--

	for _, elsif := range s.ElseIfs {
		g.writeIndent()
		g.buf.WriteString("} else if ")
		if err := g.genCondition(elsif.Cond); err != nil {
			g.addError(fmt.Errorf("line %d: %w", s.Line, err))
			g.buf.WriteString("false")
		}
		g.buf.WriteString(" {\n")

		g.indent++
		for _, stmt := range elsif.Body {
			g.genStatement(stmt)
		}
		g.indent--
	}

	if len(s.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("} else {\n")

		g.indent++
		for _, stmt := range s.Else {
			g.genStatement(stmt)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genCaseStmt(s *ast.CaseStmt) {
	g.writeIndent()

	// Check if any when clause contains a range (requires special handling)
	hasRanges := false
	for _, when := range s.WhenClauses {
		for _, val := range when.Values {
			if _, ok := val.(*ast.RangeLit); ok {
				hasRanges = true
				break
			}
		}
		if hasRanges {
			break
		}
	}

	// Handle case with subject vs case without subject
	if s.Subject != nil && !hasRanges {
		// Simple switch: switch subject { case value: ... }
		g.buf.WriteString("switch ")
		g.genExpr(s.Subject)
		g.buf.WriteString(" {\n")
	} else if s.Subject != nil && hasRanges {
		// Has ranges: need to store subject and use switch true pattern
		g.needsRuntime = true
		g.buf.WriteString("switch _subj := ")
		g.genExpr(s.Subject)
		g.buf.WriteString("; true {\n")
	} else {
		// Case without subject - use switch true
		g.buf.WriteString("switch {\n")
	}

	// Generate when clauses
	for _, whenClause := range s.WhenClauses {
		g.writeIndent()

		if s.Subject != nil && !hasRanges {
			// With subject (no ranges): case value1, value2:
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.genExpr(val)
			}
			g.buf.WriteString(":\n")
		} else if s.Subject != nil && hasRanges {
			// With subject and ranges: case _subj == value || runtime.RangeContains(range, _subj):
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(" || ")
				}
				if rangeLit, ok := val.(*ast.RangeLit); ok {
					// Generate runtime.RangeContains(range, _subj)
					g.buf.WriteString("runtime.RangeContains(")
					g.genExpr(rangeLit)
					g.buf.WriteString(", _subj)")
				} else {
					// Generate _subj == value
					g.buf.WriteString("_subj == ")
					g.genExpr(val)
				}
			}
			g.buf.WriteString(":\n")
		} else {
			// Without subject: case condition1 || condition2:
			g.buf.WriteString("case ")
			for i, val := range whenClause.Values {
				if i > 0 {
					g.buf.WriteString(" || ")
				}
				if err := g.genCondition(val); err != nil {
					g.addError(err)
					g.buf.WriteString("false")
				}
			}
			g.buf.WriteString(":\n")
		}

		g.indent++
		for _, stmt := range whenClause.Body {
			g.genStatement(stmt)
		}
		g.indent--
	}

	// Generate default (else) clause
	if len(s.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("default:\n")

		g.indent++
		for _, stmt := range s.Else {
			g.genStatement(stmt)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genCaseTypeStmt(s *ast.CaseTypeStmt) {
	g.writeIndent()

	// Check if any when clause has a binding variable
	hasBindings := false
	for _, whenClause := range s.WhenClauses {
		if whenClause.BindingVar != "" {
			hasBindings = true
			break
		}
	}

	// Generate a temp variable for the type switch (only if we have bindings)
	tempVar := "_"
	if hasBindings {
		tempVar = fmt.Sprintf("_ts%d", g.tempVarCounter)
		g.tempVarCounter++
	}

	// If subject is a simple identifier, we can use it directly
	// Otherwise, evaluate and store in temp
	var subjectExpr string
	if ident, ok := s.Subject.(*ast.Ident); ok {
		subjectExpr = ident.Name
	} else {
		// For complex expressions, always generate a temp var
		evalVar := fmt.Sprintf("_ts%d", g.tempVarCounter)
		g.tempVarCounter++
		g.buf.WriteString(fmt.Sprintf("%s := ", evalVar))
		g.genExpr(s.Subject)
		g.buf.WriteString("\n")
		g.writeIndent()
		subjectExpr = evalVar
	}

	// Generate type switch: switch _ts := subject.(type) { or switch subject.(type) {
	if hasBindings {
		g.buf.WriteString(fmt.Sprintf("switch %s := %s.(type) {\n", tempVar, subjectExpr))
	} else {
		g.buf.WriteString(fmt.Sprintf("switch %s.(type) {\n", subjectExpr))
	}

	// Generate when clauses for each type
	for _, whenClause := range s.WhenClauses {
		g.writeIndent()
		g.buf.WriteString("case ")
		g.buf.WriteString(mapType(whenClause.Type))
		g.buf.WriteString(":\n")

		g.indent++

		// If binding variable is specified, create the binding
		var prevType string
		var wasDefinedBefore bool
		if whenClause.BindingVar != "" {
			// Save previous state for cleanup
			prevType, wasDefinedBefore = g.vars[whenClause.BindingVar]

			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("%s := %s\n", whenClause.BindingVar, tempVar))
			// Suppress "declared and not used" if variable isn't referenced
			g.writeIndent()
			g.buf.WriteString(fmt.Sprintf("_ = %s\n", whenClause.BindingVar))
			// Track the variable so it's not re-declared
			g.vars[whenClause.BindingVar] = whenClause.Type
		}

		for _, stmt := range whenClause.Body {
			g.genStatement(stmt)
		}

		// Clean up variable scope
		if whenClause.BindingVar != "" {
			if !wasDefinedBefore {
				delete(g.vars, whenClause.BindingVar)
			} else {
				g.vars[whenClause.BindingVar] = prevType
			}
		}

		g.indent--
	}

	// Generate default (else) clause
	if len(s.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("default:\n")

		g.indent++
		for _, stmt := range s.Else {
			g.genStatement(stmt)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genWhileStmt(s *ast.WhileStmt) {
	g.writeIndent()
	g.buf.WriteString("for ")
	if err := g.genCondition(s.Cond); err != nil {
		g.addError(fmt.Errorf("line %d: %w", s.Line, err))
		g.buf.WriteString("false")
	}
	g.buf.WriteString(" {\n")

	g.enterLoop()
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genUntilStmt(s *ast.UntilStmt) {
	g.writeIndent()
	g.buf.WriteString("for !")
	// Wrap condition in parentheses unless it's a simple expression.
	// Simple expressions that don't need parens: identifiers, booleans, calls, selectors.
	// All other expressions (binary, nil-coalesce, etc.) need parens for correct precedence.
	needsParens := true
	switch s.Cond.(type) {
	case *ast.Ident, *ast.BoolLit, *ast.CallExpr, *ast.SelectorExpr:
		needsParens = false
	}
	if needsParens {
		g.buf.WriteString("(")
	}
	if err := g.genCondition(s.Cond); err != nil {
		g.addError(fmt.Errorf("line %d: %w", s.Line, err))
		g.buf.WriteString("false")
	}
	if needsParens {
		g.buf.WriteString(")")
	}
	g.buf.WriteString(" {\n")

	g.enterLoop()
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genLoopStmt(s *ast.LoopStmt) {
	g.writeIndent()
	g.buf.WriteString("for {\n")

	g.enterLoop()
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genForStmt(s *ast.ForStmt) {
	// Save variable state
	prevType, wasDefinedBefore := g.vars[s.Var]
	var prevType2 string
	var wasDefinedBefore2 bool
	if s.Var2 != "" {
		prevType2, wasDefinedBefore2 = g.vars[s.Var2]
	}

	// Check if iterable is a range literal - optimize to C-style for loop
	if rangeLit, ok := s.Iterable.(*ast.RangeLit); ok {
		g.genForRangeLoop(s.Var, rangeLit, s.Body, wasDefinedBefore, prevType)
		return
	}

	// Check if iterable is a variable of type Range
	if ident, ok := s.Iterable.(*ast.Ident); ok {
		if g.typeInfo.GetRugbyType(ident) == "Range" {
			g.genForRangeVarLoop(s.Var, ident.Name, s.Body, wasDefinedBefore, prevType)
			return
		}
	}

	g.writeIndent()
	// Check if iterating over a channel - channels only allow one iteration variable
	// Note: type might be in Rugby format (Chan<T>) or Go format (chan T)
	iterType := g.inferTypeFromExpr(s.Iterable)
	if strings.HasPrefix(iterType, "chan ") || strings.HasPrefix(iterType, "Chan<") {
		g.buf.WriteString("for ")
		g.buf.WriteString(s.Var)
		g.buf.WriteString(" := range ")
		g.genExpr(s.Iterable)
		g.buf.WriteString(" {\n")
	} else if s.Var2 != "" {
		// Two-variable form: for key, value in map/array
		g.buf.WriteString("for ")
		g.buf.WriteString(s.Var)
		g.buf.WriteString(", ")
		g.buf.WriteString(s.Var2)
		g.buf.WriteString(" := range ")
		g.genExpr(s.Iterable)
		g.buf.WriteString(" {\n")
	} else if strings.HasPrefix(iterType, "map[") || strings.HasPrefix(iterType, "Map<") {
		// Single-variable map iteration: get keys only
		g.buf.WriteString("for ")
		g.buf.WriteString(s.Var)
		g.buf.WriteString(" := range ")
		g.genExpr(s.Iterable)
		g.buf.WriteString(" {\n")
	} else {
		// Single-variable array iteration: skip index, get values
		g.buf.WriteString("for _, ")
		g.buf.WriteString(s.Var)
		g.buf.WriteString(" := range ")
		g.genExpr(s.Iterable)
		g.buf.WriteString(" {\n")
	}

	g.vars[s.Var] = "" // loop variable, type unknown
	if s.Var2 != "" {
		g.vars[s.Var2] = "" // second loop variable
	}

	g.enterLoop()
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	// Restore variable state
	if !wasDefinedBefore {
		delete(g.vars, s.Var)
	} else {
		g.vars[s.Var] = prevType
	}
	if s.Var2 != "" {
		if !wasDefinedBefore2 {
			delete(g.vars, s.Var2)
		} else {
			g.vars[s.Var2] = prevType2
		}
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genForRangeVarLoop(varName string, rangeVar string, body []ast.Statement, wasDefinedBefore bool, prevType string) {
	g.writeIndent()
	g.buf.WriteString("for ")
	g.buf.WriteString(varName)
	g.buf.WriteString(" := ")
	g.buf.WriteString(rangeVar)
	g.buf.WriteString(".Start; ")

	// Condition: (r.Exclusive && i < r.End) || (!r.Exclusive && i <= r.End)
	g.buf.WriteString(fmt.Sprintf("(%s.Exclusive && %s < %s.End) || (!%s.Exclusive && %s <= %s.End)",
		rangeVar, varName, rangeVar, rangeVar, varName, rangeVar))

	g.buf.WriteString("; ")
	g.buf.WriteString(varName)
	g.buf.WriteString("++ {\n")

	g.vars[varName] = "Int"

	g.enterLoop()
	g.indent++
	for _, stmt := range body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	if !wasDefinedBefore {
		delete(g.vars, varName)
	} else {
		g.vars[varName] = prevType
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genForRangeLoop(varName string, r *ast.RangeLit, body []ast.Statement, wasDefinedBefore bool, prevType string) {
	// Optimization: use Go 1.22 range-over-int for 0...n loops
	// Check if start is 0
	isStartZero := false
	if intLit, ok := r.Start.(*ast.IntLit); ok && intLit.Value == 0 {
		isStartZero = true
	}

	g.writeIndent()
	if isStartZero && r.Exclusive {
		// Generate: for i := range end { ... }
		g.buf.WriteString("for ")
		g.buf.WriteString(varName)
		g.buf.WriteString(" := range ")
		g.genExpr(r.End)
		g.buf.WriteString(" {\n")
	} else {
		// Generate: for i := start; i <= end; i++ { ... }
		// or:       for i := start; i < end; i++ { ... } (exclusive)
		g.buf.WriteString("for ")
		g.buf.WriteString(varName)
		g.buf.WriteString(" := ")
		g.genExpr(r.Start)
		g.buf.WriteString("; ")
		g.buf.WriteString(varName)
		if r.Exclusive {
			g.buf.WriteString(" < ")
		} else {
			g.buf.WriteString(" <= ")
		}
		g.genExpr(r.End)
		g.buf.WriteString("; ")
		g.buf.WriteString(varName)
		g.buf.WriteString("++ {\n")
	}

	g.vars[varName] = "Int" // range loop variable is always Int

	g.enterLoop()
	g.indent++
	for _, stmt := range body {
		g.genStatement(stmt)
	}
	g.indent--
	g.exitLoop()

	// Restore variable state
	if !wasDefinedBefore {
		delete(g.vars, varName)
	} else {
		g.vars[varName] = prevType
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genExprStmt(s *ast.ExprStmt) {
	// Handle BangExpr: call()! as a statement
	if bangExpr, ok := s.Expr.(*ast.BangExpr); ok {
		g.genBangStmt(bangExpr)
		return
	}

	// Handle RescueExpr: call() rescue do ... end as a statement
	if rescueExpr, ok := s.Expr.(*ast.RescueExpr); ok {
		g.genRescueStmt(rescueExpr)
		return
	}

	// Handle BinaryExpr with << on Ident: arr << value becomes arr = arr << value
	// This provides Ruby-like mutation semantics for array append
	// For channels, just generate: ch <- value (no assignment wrapper needed)
	if binExpr, ok := s.Expr.(*ast.BinaryExpr); ok {
		if binExpr.Op == "<<" {
			// Check if left operand is a channel - generate native send without assignment
			if g.typeInfo.GetTypeKind(binExpr.Left) == TypeChannel {
				g.writeIndent()
				g.genExpr(s.Expr) // genExpr generates "ch <- value" for channels
				g.buf.WriteString("\n")
				return
			}
			if ident, ok := binExpr.Left.(*ast.Ident); ok {
				// Generate: arr = runtime.ShiftLeft(arr, value).([]T)
				g.writeIndent()
				g.buf.WriteString(ident.Name)
				g.buf.WriteString(" = ")
				g.genExpr(s.Expr)
				// Add type assertion if needed
				if goType := g.getShiftLeftTargetType(binExpr); goType != "" {
					g.buf.WriteString(".(")
					g.buf.WriteString(goType)
					g.buf.WriteString(")")
				}
				g.buf.WriteString("\n")
				return
			}
		}
	}

	// Convert SelectorExpr to CallExpr when used as statement (Ruby-style method call)
	// e.g., "obj.foo" as statement becomes "obj.foo()"
	expr := s.Expr
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		expr = &ast.CallExpr{Func: sel, Args: nil}
	}

	// Convert Ident to CallExpr when it's a no-arg function (Ruby-style method call)
	// e.g., "helper" as statement becomes "helper()"
	if ident, ok := expr.(*ast.Ident); ok {
		if g.isNoArgFunction(ident.Name) {
			expr = &ast.CallExpr{Func: ident, Args: nil}
		}
	}

	if s.Condition != nil {
		// Wrap in if/unless block
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Condition); err != nil {
			g.addError(err)
			g.buf.WriteString("false")
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
		g.writeIndent()
		g.genExpr(expr)
		g.buf.WriteString("\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	} else {
		g.writeIndent()
		g.genExpr(expr)
		g.buf.WriteString("\n")
	}
}

func (g *Generator) genBreakStmt(s *ast.BreakStmt) {
	if s.Condition != nil {
		// Wrap in if/unless block
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Condition); err != nil {
			g.addError(err)
			g.buf.WriteString("false")
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
	}

	g.writeIndent()
	g.buf.WriteString("break\n")

	if s.Condition != nil {
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

func (g *Generator) genNextStmt(s *ast.NextStmt) {
	if s.Condition != nil {
		// Wrap in if/unless block
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Condition); err != nil {
			g.addError(err)
			g.buf.WriteString("false")
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
	}

	g.writeIndent()
	g.buf.WriteString("continue\n")

	if s.Condition != nil {
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

func (g *Generator) genReturnStmt(s *ast.ReturnStmt) {
	if s.Condition != nil {
		// Wrap in if/unless block
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
		}
		if err := g.genCondition(s.Condition); err != nil {
			g.addError(err)
			g.buf.WriteString("false")
		}
		if s.IsUnless {
			g.buf.WriteString(")")
		}
		g.buf.WriteString(" {\n")
		g.indent++
	}

	g.writeIndent()
	// In inlined lambdas (each/times loops), return acts like continue
	// This matches Ruby semantics where return in a block skips to the next iteration
	// The return value (if any) is discarded since each/times ignore block return values
	if g.inInlinedLambda {
		g.buf.WriteString("continue")
		// Close the condition block if present
		if s.Condition != nil {
			g.buf.WriteString("\n")
			g.indent--
			g.writeIndent()
			g.buf.WriteString("}\n")
		} else {
			g.buf.WriteString("\n")
		}
		return
	}
	g.buf.WriteString("return")
	if len(s.Values) > 0 {
		g.buf.WriteString(" ")
		for i, val := range s.Values {
			if i > 0 {
				g.buf.WriteString(", ")
			}

			// Check if we need to wrap for Optional types
			needsWrap := false
			baseType := ""
			handled := false

			if i < len(g.currentReturnTypes) {
				targetType := g.currentReturnTypes[i]
				if isValueTypeOptional(targetType) {
					baseType = strings.TrimSuffix(targetType, "?")

					// If returning nil, use NoneT()
					if _, isNil := val.(*ast.NilLit); isNil {
						g.buf.WriteString(fmt.Sprintf("runtime.None%s()", baseType))
						handled = true
					} else {
						// Infer type of val to see if it needs wrapping
						inferred := g.inferTypeFromExpr(val)
						if inferred == baseType {
							needsWrap = true
						}
					}
				}
			}

			if !handled {
				if needsWrap {
					g.buf.WriteString(fmt.Sprintf("runtime.Some%s(", baseType))
					g.genExpr(val)
					g.buf.WriteString(")")
				} else {
					g.genExpr(val)
				}
			}
		}
	}
	g.buf.WriteString("\n")

	if s.Condition != nil {
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

func (g *Generator) genPanicStmt(s *ast.PanicStmt) {
	if s.Condition != nil {
		// panic "msg" if cond  ->  if cond { panic("msg") }
		// panic "msg" unless cond  ->  if !cond { panic("msg") }
		g.writeIndent()
		g.buf.WriteString("if ")
		if s.IsUnless {
			g.buf.WriteString("!(")
			g.genExpr(s.Condition)
			g.buf.WriteString(")")
		} else {
			g.genExpr(s.Condition)
		}
		g.buf.WriteString(" {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString("panic(")
		g.genExpr(s.Message)
		g.buf.WriteString(")\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	} else {
		g.writeIndent()
		g.buf.WriteString("panic(")
		g.genExpr(s.Message)
		g.buf.WriteString(")\n")
	}
}

func (g *Generator) genDeferStmt(s *ast.DeferStmt) {
	g.writeIndent()
	g.buf.WriteString("defer ")
	g.genCallExpr(s.Call)
	g.buf.WriteString("\n")
}

func (g *Generator) genGoStmt(s *ast.GoStmt) {
	g.writeIndent()
	// Check if it's a block (go do ... end) or a simple call (go func())
	if s.Block != nil {
		// go func() { ... }()
		g.buf.WriteString("go func() {\n")
		g.indent++
		for _, stmt := range s.Block.Body {
			g.genStatement(stmt)
		}
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}()\n")
	} else if s.Call != nil {
		// go func()
		g.buf.WriteString("go ")
		if call, ok := s.Call.(*ast.CallExpr); ok {
			g.genCallExpr(call)
		} else {
			g.addError(fmt.Errorf("line %d: go statement must be a function call", s.Line))
			g.buf.WriteString("func(){}()")
		}
		g.buf.WriteString("\n")
	}
}

func (g *Generator) genSelectStmt(s *ast.SelectStmt) {
	g.writeIndent()
	g.buf.WriteString("select {\n")

	for _, when := range s.Cases {
		g.writeIndent()
		g.buf.WriteString("case ")

		if !when.IsSend {
			// case val := <-ch:
			if when.AssignName != "" {
				g.buf.WriteString(when.AssignName)
				g.buf.WriteString(" := ")
				// Infer type of received value (not implemented fully without type info)
				// For now assume "any" or rely on runtime
				g.vars[when.AssignName] = "any"
			}

			// Unwrap .receive or .try_receive to get the underlying channel
			expr := when.Chan
			if sel, ok := expr.(*ast.SelectorExpr); ok {
				if sel.Sel == "receive" || sel.Sel == "try_receive" {
					expr = sel.X
				}
			}

			g.buf.WriteString("<-")
			g.genExpr(expr)
		} else {
			// case ch <- val:
			g.genExpr(when.Chan)
			g.buf.WriteString(" <- ")
			g.genExpr(when.Value)
		}

		g.buf.WriteString(":\n")
		g.indent++
		for _, stmt := range when.Body {
			g.genStatement(stmt)
		}
		g.indent--
	}

	if len(s.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("default:\n")
		g.indent++
		for _, stmt := range s.Else {
			g.genStatement(stmt)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

func (g *Generator) genChanSendStmt(s *ast.ChanSendStmt) {
	g.writeIndent()
	g.genExpr(s.Chan)
	g.buf.WriteString(" <- ")
	g.genExpr(s.Value)
	g.buf.WriteString("\n")
}

func (g *Generator) genConcurrentlyStmt(s *ast.ConcurrentlyStmt) {
	// Generate:
	// func() (T, T, error) {
	//   ctx, cancel := context.WithCancel(context.Background())
	//   defer cancel()
	//   var wg sync.WaitGroup
	//   scope := &runtime.Scope{Ctx: ctx, Wg: &wg}
	//   ...
	//   wg.Wait()
	//   return ...
	// }()

	g.needsRuntime = true
	g.buf.WriteString("func() ")

	// Return types are currently just "any" because we can't easily infer block return type
	// This might need refinement for typed returns
	g.buf.WriteString("any")

	g.buf.WriteString(" {\n")
	g.indent++

	// Create scope variable (use fallback if not specified)
	scopeVar := s.ScopeVar
	if scopeVar == "" {
		scopeVar = "_scope"
	}
	g.vars[scopeVar] = "Scope"

	// Create scope object
	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("%s := runtime.NewScope()\n", scopeVar))

	// Generate: defer scope.Wait()
	g.writeIndent()
	g.buf.WriteString(fmt.Sprintf("defer %s.Wait()\n", scopeVar))

	// Generate body
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}

	g.indent--
	g.writeIndent()
	g.buf.WriteString("}()\n") // Call the closure immediately
}
