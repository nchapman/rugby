package codegen

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
)

func (g *Generator) genFuncDecl(fn *ast.FuncDecl) {
	clear(g.vars) // reset vars for each function
	g.currentReturnTypes = fn.ReturnTypes
	g.inMainFunc = fn.Name == "main"

	// Validate: functions ending in ? must return Bool
	if strings.HasSuffix(fn.Name, "?") {
		if len(fn.ReturnTypes) != 1 || fn.ReturnTypes[0] != "Bool" {
			g.addError(fmt.Errorf("line %d: function '%s' ending in '?' must return Bool", fn.Line, fn.Name))
		}
	}

	// Mark parameters as declared variables with their types
	for _, param := range fn.Params {
		g.vars[param.Name] = param.Type
	}

	// Generate function name with proper casing
	// pub def parse_json -> ParseJSON (exported)
	// def parse_json -> parseJSON (unexported)
	var funcName string
	if fn.Pub {
		funcName = snakeToPascalWithAcronyms(fn.Name)
	} else {
		funcName = snakeToCamelWithAcronyms(fn.Name)
	}

	// Generate function signature
	g.buf.WriteString(fmt.Sprintf("func %s(", funcName))
	for i, param := range fn.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		if param.Type != "" {
			g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, mapType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s any", param.Name))
		}
	}
	g.buf.WriteString(")")

	// Generate return type(s) if specified
	if len(fn.ReturnTypes) == 1 {
		g.buf.WriteString(" ")
		g.buf.WriteString(mapType(fn.ReturnTypes[0]))
	} else if len(fn.ReturnTypes) > 1 {
		g.buf.WriteString(" (")
		for i, rt := range fn.ReturnTypes {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.buf.WriteString(mapType(rt))
		}
		g.buf.WriteString(")")
	}

	g.buf.WriteString(" {\n")

	g.indent++
	for i, stmt := range fn.Body {
		isLast := i == len(fn.Body)-1
		if isLast && len(fn.ReturnTypes) > 0 {
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				// Implicit return: treat last expression as return value
				retStmt := &ast.ReturnStmt{Values: []ast.Expression{exprStmt.Expr}}
				g.genReturnStmt(retStmt)
				continue
			}
		}
		g.genStatement(stmt)
	}
	g.indent--
	g.buf.WriteString("}\n")
	g.currentReturnTypes = nil
	g.inMainFunc = false
}

// genModuleDecl stores the module definition for later use when classes include it.
// Modules don't generate code directly - their fields and methods are injected
// into classes that include them.
func (g *Generator) genModuleDecl(mod *ast.ModuleDecl) {
	g.modules[mod.Name] = mod
}

func (g *Generator) genClassDecl(cls *ast.ClassDecl) {
	className := cls.Name
	g.currentClass = className
	g.currentClassEmbeds = cls.Embeds
	g.pubClasses[className] = cls.Pub
	clear(g.classFields)

	// Collect all fields: explicit, inferred, from accessors, and from included modules
	// Track field names to avoid duplicates
	fieldNames := make(map[string]bool)
	allFields := make([]*ast.FieldDecl, 0, len(cls.Fields)+len(cls.Accessors))
	for _, f := range cls.Fields {
		if !fieldNames[f.Name] {
			allFields = append(allFields, f)
			fieldNames[f.Name] = true
		}
	}
	for _, acc := range cls.Accessors {
		if !fieldNames[acc.Name] {
			allFields = append(allFields, &ast.FieldDecl{Name: acc.Name, Type: acc.Type})
			fieldNames[acc.Name] = true
		}
	}

	// Collect all accessors including from modules
	allAccessors := append([]*ast.AccessorDecl{}, cls.Accessors...)

	// Collect all methods including from modules
	allMethods := append([]*ast.MethodDecl{}, cls.Methods...)

	// Include modules: add their fields, accessors, and methods
	for _, modName := range cls.Includes {
		mod, ok := g.modules[modName]
		if !ok {
			g.addError(fmt.Errorf("undefined module: %s", modName))
			continue
		}
		// Add module fields (skip duplicates)
		for _, f := range mod.Fields {
			if !fieldNames[f.Name] {
				allFields = append(allFields, f)
				fieldNames[f.Name] = true
			}
		}
		for _, acc := range mod.Accessors {
			if !fieldNames[acc.Name] {
				allFields = append(allFields, &ast.FieldDecl{Name: acc.Name, Type: acc.Type})
				fieldNames[acc.Name] = true
			}
		}
		// Add module accessors
		allAccessors = append(allAccessors, mod.Accessors...)
		// Add module methods (these will be "specialized" by being generated with the class's receiver)
		allMethods = append(allMethods, mod.Methods...)
	}

	// Emit struct definition
	// Class names are already PascalCase by convention; pub affects field/method visibility
	hasContent := len(cls.Embeds) > 0 || len(allFields) > 0
	if hasContent {
		g.buf.WriteString(fmt.Sprintf("type %s struct {\n", className))
		for _, embed := range cls.Embeds {
			g.buf.WriteString("\t")
			g.buf.WriteString(embed)
			g.buf.WriteString("\n")
		}
		for _, field := range allFields {
			g.classFields[field.Name] = field.Type // Store field type
			if field.Type != "" {
				g.buf.WriteString(fmt.Sprintf("\t%s %s\n", field.Name, mapType(field.Type)))
			} else {
				g.buf.WriteString(fmt.Sprintf("\t%s any\n", field.Name))
			}
		}
		g.buf.WriteString("}\n\n")
	} else {
		g.buf.WriteString(fmt.Sprintf("type %s struct{}\n\n", className))
	}

	// Emit constructor if initialize exists
	for _, method := range cls.Methods {
		if method.Name == "initialize" {
			g.genConstructor(className, method, cls.Pub)
			break
		}
	}

	// Emit accessor methods (including from modules)
	for _, acc := range allAccessors {
		g.genAccessorMethods(className, acc, cls.Pub)
	}

	// Emit methods (skip initialize - it's the constructor)
	// This includes module methods which are "specialized" by being generated with the class's receiver
	for _, method := range allMethods {
		if method.Name != "initialize" {
			g.genMethodDecl(className, method)
		}
	}

	// Emit compile-time interface conformance checks
	for _, iface := range cls.Implements {
		g.buf.WriteString(fmt.Sprintf("var _ %s = (*%s)(nil)\n", iface, className))
	}
	if len(cls.Implements) > 0 {
		g.buf.WriteString("\n")
	}

	g.currentClass = ""
	g.currentClassEmbeds = nil
	clear(g.classFields)
}

func (g *Generator) genAccessorMethods(className string, acc *ast.AccessorDecl, pub bool) {
	recv := receiverName(className)
	goType := mapType(acc.Type)

	// Generate getter for "getter" and "property"
	if acc.Kind == "getter" || acc.Kind == "property" {
		// getter/property generates: func (r *T) Name() T { return r.name }
		// Method name is PascalCase if pub class, camelCase otherwise
		var methodName string
		if pub {
			methodName = snakeToPascalWithAcronyms(acc.Name)
		} else {
			methodName = snakeToCamelWithAcronyms(acc.Name)
		}
		g.buf.WriteString(fmt.Sprintf("func (%s *%s) %s() %s {\n", recv, className, methodName, goType))
		g.buf.WriteString(fmt.Sprintf("\treturn %s.%s\n", recv, acc.Name))
		g.buf.WriteString("}\n\n")
	}

	// Generate setter for "setter" and "property"
	if acc.Kind == "setter" || acc.Kind == "property" {
		// setter/property generates: func (r *T) SetName(v T) { r.name = v }
		// Method name is SetPascalCase if pub class, setcamelCase otherwise
		var methodName string
		if pub {
			methodName = "Set" + snakeToPascalWithAcronyms(acc.Name)
		} else {
			methodName = "set" + snakeToPascalWithAcronyms(acc.Name)
		}
		g.buf.WriteString(fmt.Sprintf("func (%s *%s) %s(v %s) {\n", recv, className, methodName, goType))
		g.buf.WriteString(fmt.Sprintf("\t%s.%s = v\n", recv, acc.Name))
		g.buf.WriteString("}\n\n")
	}
}

func (g *Generator) genInterfaceDecl(iface *ast.InterfaceDecl) {
	// Track this interface for zero-value generation
	g.interfaces[iface.Name] = true

	// Generate: type InterfaceName interface { ... }
	g.buf.WriteString(fmt.Sprintf("type %s interface {\n", iface.Name))

	// Embed parent interfaces
	for _, parent := range iface.Parents {
		g.buf.WriteString(fmt.Sprintf("\t%s\n", parent))
	}

	for _, method := range iface.Methods {
		g.buf.WriteString("\t")
		// Interface methods are always exported (PascalCase) per Go interface rules
		methodName := snakeToPascalWithAcronyms(method.Name)
		g.buf.WriteString(methodName)
		g.buf.WriteString("(")

		// Parameters (just types, no names in Go interface definitions)
		for i, param := range method.Params {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			if param.Type != "" {
				g.buf.WriteString(mapType(param.Type))
			} else {
				g.buf.WriteString("any")
			}
		}
		g.buf.WriteString(")")

		// Return types
		if len(method.ReturnTypes) == 1 {
			g.buf.WriteString(" ")
			g.buf.WriteString(mapType(method.ReturnTypes[0]))
		} else if len(method.ReturnTypes) > 1 {
			g.buf.WriteString(" (")
			for i, rt := range method.ReturnTypes {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.buf.WriteString(mapType(rt))
			}
			g.buf.WriteString(")")
		}

		g.buf.WriteString("\n")
	}

	g.buf.WriteString("}\n")
}

func (g *Generator) genConstructor(className string, method *ast.MethodDecl, pub bool) {
	clear(g.vars)

	// Separate promoted parameters (@field : Type) from regular parameters
	var promotedParams []struct {
		fieldName string
		paramName string
		paramType string
	}

	// Mark parameters as declared variables
	for _, param := range method.Params {
		if len(param.Name) > 0 && param.Name[0] == '@' {
			// Parameter promotion: @name : Type
			fieldName := param.Name[1:] // strip @
			promotedParams = append(promotedParams, struct {
				fieldName string
				paramName string
				paramType string
			}{fieldName, fieldName, param.Type})
			g.vars[fieldName] = param.Type
		} else {
			g.vars[param.Name] = param.Type
		}
	}

	// Receiver name for field assignments
	recv := receiverName(className)

	// Constructor: func NewClassName(params) *ClassName (pub) or newClassName (non-pub)
	constructorName := "new" + className
	if pub {
		constructorName = "New" + className
	}
	g.buf.WriteString(fmt.Sprintf("func %s(", constructorName))
	for i, param := range method.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		// Use field name (without @) for promoted parameters
		paramName := param.Name
		if len(paramName) > 0 && paramName[0] == '@' {
			paramName = paramName[1:]
		}
		if param.Type != "" {
			g.buf.WriteString(fmt.Sprintf("%s %s", paramName, mapType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s any", paramName))
		}
	}
	g.buf.WriteString(fmt.Sprintf(") *%s {\n", className))

	// Create instance
	g.buf.WriteString(fmt.Sprintf("\t%s := &%s{}\n", recv, className))

	// Auto-assign promoted parameters
	for _, promoted := range promotedParams {
		g.buf.WriteString(fmt.Sprintf("\t%s.%s = %s\n", recv, promoted.fieldName, promoted.paramName))
	}

	// Generate body (instance var assignments become field sets)
	g.indent++
	for _, stmt := range method.Body {
		g.genStatement(stmt)
	}
	g.indent--

	// Return instance
	g.buf.WriteString(fmt.Sprintf("\treturn %s\n", recv))
	g.buf.WriteString("}\n\n")
}

func (g *Generator) genMethodDecl(className string, method *ast.MethodDecl) {
	clear(g.vars) // reset vars for each method
	g.currentReturnTypes = method.ReturnTypes
	g.currentMethod = method.Name
	g.currentMethodPub = method.Pub

	// Validate: methods ending in ? must return Bool
	if strings.HasSuffix(method.Name, "?") {
		if len(method.ReturnTypes) != 1 || method.ReturnTypes[0] != "Bool" {
			g.addError(fmt.Errorf("line %d: method '%s' ending in '?' must return Bool", method.Line, method.Name))
		}
	}

	// Mark parameters as declared variables with their types
	for _, param := range method.Params {
		// Handle promoted parameters (shouldn't appear in regular methods, but be safe)
		paramName := param.Name
		if len(paramName) > 0 && paramName[0] == '@' {
			paramName = paramName[1:]
		}
		g.vars[paramName] = param.Type
	}

	// Receiver name: first letter of class, lowercase
	recv := receiverName(className)

	// Special method handling
	// to_s -> String() string (satisfies fmt.Stringer)
	// Only applies when to_s has no parameters
	if method.Name == "to_s" && len(method.Params) == 0 {
		g.buf.WriteString(fmt.Sprintf("func (%s *%s) String() string {\n", recv, className))
		g.indent++
		for _, stmt := range method.Body {
			g.genStatement(stmt)
		}
		g.indent--
		g.buf.WriteString("}\n\n")
		g.currentReturnTypes = nil
		g.currentMethod = ""
		g.currentMethodPub = false
		return
	}

	// message -> Error() string (satisfies error interface)
	// Only applies when message has no parameters
	if method.Name == "message" && len(method.Params) == 0 {
		g.buf.WriteString(fmt.Sprintf("func (%s *%s) Error() string {\n", recv, className))
		g.indent++
		for _, stmt := range method.Body {
			g.genStatement(stmt)
		}
		g.indent--
		g.buf.WriteString("}\n\n")
		g.currentReturnTypes = nil
		g.currentMethod = ""
		g.currentMethodPub = false
		return
	}

	// == -> Equal(other any) bool (satisfies runtime.Equaler)
	// Only applies when == has exactly one parameter
	if method.Name == "==" && len(method.Params) == 1 {
		param := method.Params[0]
		g.buf.WriteString(fmt.Sprintf("func (%s *%s) Equal(other any) bool {\n", recv, className))
		g.indent++
		// Create type assertion for the parameter
		g.writeIndent()
		g.buf.WriteString(fmt.Sprintf("%s, ok := other.(*%s)\n", param.Name, className))
		g.writeIndent()
		g.buf.WriteString("if !ok {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString("return false\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
		// Generate body statements, with implicit return for last expression
		for i, stmt := range method.Body {
			isLast := i == len(method.Body)-1
			if isLast {
				// If last statement is an expression, add return
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					g.writeIndent()
					g.buf.WriteString("return ")
					g.genExpr(exprStmt.Expr)
					g.buf.WriteString("\n")
					continue
				}
			}
			g.genStatement(stmt)
		}
		g.indent--
		g.buf.WriteString("}\n\n")
		g.currentReturnTypes = nil
		g.currentMethod = ""
		g.currentMethodPub = false
		return
	}

	// Method name: convert snake_case to proper casing
	// pub def in a pub class -> PascalCase (exported)
	// def in any class -> camelCase (unexported)
	var methodName string
	if method.Pub {
		methodName = snakeToPascalWithAcronyms(method.Name)
	} else {
		methodName = snakeToCamelWithAcronyms(method.Name)
	}

	// Generate method signature with pointer receiver: func (r *ClassName) MethodName(params) returns
	g.buf.WriteString(fmt.Sprintf("func (%s *%s) %s(", recv, className, methodName))
	for i, param := range method.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		if param.Type != "" {
			g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, mapType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s any", param.Name))
		}
	}
	g.buf.WriteString(")")

	// Generate return type(s) if specified
	if len(method.ReturnTypes) == 1 {
		g.buf.WriteString(" ")
		g.buf.WriteString(mapType(method.ReturnTypes[0]))
	} else if len(method.ReturnTypes) > 1 {
		g.buf.WriteString(" (")
		for i, rt := range method.ReturnTypes {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.buf.WriteString(mapType(rt))
		}
		g.buf.WriteString(")")
	}

	g.buf.WriteString(" {\n")

	g.indent++
	for i, stmt := range method.Body {
		isLast := i == len(method.Body)-1
		if isLast && len(method.ReturnTypes) > 0 {
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				// Implicit return
				retStmt := &ast.ReturnStmt{Values: []ast.Expression{exprStmt.Expr}}
				g.genReturnStmt(retStmt)
				continue
			}
		}
		g.genStatement(stmt)
	}
	g.indent--
	g.buf.WriteString("}\n\n")
	g.currentReturnTypes = nil
	g.currentMethod = ""
	g.currentMethodPub = false
}
