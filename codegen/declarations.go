package codegen

import (
	"fmt"
	"strings"

	"github.com/nchapman/rugby/ast"
)

// mapParamType converts Rugby types to Go types, making class types pointers.
// This handles nested generic types like Chan[Job] -> chan *Job and Array[Job] -> []*Job.
func (g *Generator) mapParamType(rubyType string) string {
	// Handle Array[T] - recursively convert element type
	if strings.HasPrefix(rubyType, "Array[") && strings.HasSuffix(rubyType, "]") {
		inner := rubyType[6 : len(rubyType)-1]
		return "[]" + g.mapParamType(inner)
	}
	// Handle Chan[T] - recursively convert element type
	if strings.HasPrefix(rubyType, "Chan[") && strings.HasSuffix(rubyType, "]") {
		inner := rubyType[5 : len(rubyType)-1]
		return "chan " + g.mapParamType(inner)
	}
	// Handle Map[K, V] - recursively convert key and value types
	if strings.HasPrefix(rubyType, "Map[") && strings.HasSuffix(rubyType, "]") {
		content := rubyType[4 : len(rubyType)-1]
		parts := strings.Split(content, ",")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			return fmt.Sprintf("map[%s]%s", g.mapParamType(key), g.mapParamType(val))
		}
	}

	// Use mapType for the base mapping
	goType := mapType(rubyType)

	// If it's a known class type (not already a pointer), make it a pointer
	// Note: mapType returns class names unchanged, so we check for them here
	if g.classes[rubyType] && !strings.HasPrefix(goType, "*") && !strings.HasPrefix(goType, "[]") && !strings.HasPrefix(goType, "map[") {
		return "*" + goType
	}

	return goType
}

func (g *Generator) genFuncDecl(fn *ast.FuncDecl) {
	clear(g.vars)          // reset vars for each function
	clear(g.goInteropVars) // reset Go interop tracking for each function
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
			g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, g.mapParamType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s any", param.Name))
		}
	}
	g.buf.WriteString(")")

	// Generate return type(s) if specified
	// Use mapParamType to properly convert class types to pointers
	if len(fn.ReturnTypes) == 1 {
		g.buf.WriteString(" ")
		g.buf.WriteString(g.mapParamType(fn.ReturnTypes[0]))
	} else if len(fn.ReturnTypes) > 1 {
		g.buf.WriteString(" (")
		for i, rt := range fn.ReturnTypes {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.buf.WriteString(g.mapParamType(rt))
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
	g.classes[className] = true // Track all class names for pointer type mapping
	clear(g.classFields)
	clear(g.currentClassModuleMethods)    // Reset for new class
	clear(g.currentClassInterfaceMethods) // Reset for new class

	// For structural typing support: export methods that match ANY interface method
	// name in the program. In Go, interface methods must be exported (PascalCase),
	// so any class method that could satisfy an interface must also be exported.
	// This enables Go-style duck typing where classes can satisfy interfaces
	// without explicit 'implements' declarations.
	for _, methods := range g.interfaceMethods {
		for methodName := range methods {
			g.currentClassInterfaceMethods[methodName] = true
		}
	}

	// Collect all fields: explicit, inferred, from accessors, and from included modules
	// Track field names to avoid duplicates
	// accessorFields tracks which field names come from accessors (need underscore prefix)
	fieldNames := make(map[string]bool)
	clear(g.accessorFields) // Reset for new class
	allFields := make([]*ast.FieldDecl, 0, len(cls.Fields)+len(cls.Accessors))
	for _, f := range cls.Fields {
		if !fieldNames[f.Name] {
			allFields = append(allFields, f)
			fieldNames[f.Name] = true
		}
	}
	for _, acc := range cls.Accessors {
		// Always track accessor fields for underscore prefix, even if field already exists
		// (field might come from parameter promotion in initialize)
		g.accessorFields[acc.Name] = true
		if !fieldNames[acc.Name] {
			allFields = append(allFields, &ast.FieldDecl{Name: acc.Name, Type: acc.Type})
			fieldNames[acc.Name] = true
		} else if acc.Type != "" {
			// If field already exists but has no type, use the accessor's type
			for _, f := range allFields {
				if f.Name == acc.Name && f.Type == "" {
					f.Type = acc.Type
					break
				}
			}
		}
	}

	// Track accessor and method origins for conflict detection
	// Maps name -> source (class name or "ClassName (from ModuleName)")
	accessorSources := make(map[string]string)
	methodSources := make(map[string]string)

	// Collect all accessors including from modules (with conflict detection)
	allAccessors := make([]*ast.AccessorDecl, 0, len(cls.Accessors))
	for _, acc := range cls.Accessors {
		accessorSources[acc.Name] = className
		allAccessors = append(allAccessors, acc)
	}

	// Collect all methods including from modules (with conflict detection)
	allMethods := make([]*ast.MethodDecl, 0, len(cls.Methods))
	for _, m := range cls.Methods {
		methodSources[m.Name] = className
		allMethods = append(allMethods, m)
	}

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
			// Always track accessor fields for underscore prefix
			g.accessorFields[acc.Name] = true
			if !fieldNames[acc.Name] {
				allFields = append(allFields, &ast.FieldDecl{Name: acc.Name, Type: acc.Type})
				fieldNames[acc.Name] = true
			}
		}
		// Add module accessors with conflict detection
		for _, acc := range mod.Accessors {
			if source, exists := accessorSources[acc.Name]; exists {
				// Conflict: accessor already defined
				// If class defined it, class wins (intentional override)
				if source == className {
					continue
				}
				g.addError(fmt.Errorf("accessor '%s' from module '%s' conflicts with %s", acc.Name, modName, source))
			} else {
				accessorSources[acc.Name] = fmt.Sprintf("module '%s'", modName)
				allAccessors = append(allAccessors, acc)
			}
		}
		// Add module methods with conflict detection
		for _, m := range mod.Methods {
			if source, exists := methodSources[m.Name]; exists {
				// Conflict: method already defined
				// If class defined it, class wins (intentional override)
				// If another module defined it, that's a conflict
				if source == className {
					// Class intentionally overrides module method - skip module version
					continue
				}
				g.addError(fmt.Errorf("method '%s' from module '%s' conflicts with %s", m.Name, modName, source))
			} else {
				methodSources[m.Name] = fmt.Sprintf("module '%s'", modName)
				allMethods = append(allMethods, m)
				// Track as module method for bare identifier resolution in class methods
				g.currentClassModuleMethods[m.Name] = true
			}
		}
	}

	// Also add underscore prefix for fields that have methods with the same name
	// This handles cases like: field @name + method def name -> ... (without accessor declarations)
	for _, m := range allMethods {
		if fieldNames[m.Name] && !g.accessorFields[m.Name] {
			g.accessorFields[m.Name] = true
		}
	}

	// Store accessor fields globally for subclass access
	g.classAccessorFields[className] = make(map[string]bool)
	for name := range g.accessorFields {
		g.classAccessorFields[className][name] = true
	}

	// Emit struct definition
	// Class names are already PascalCase by convention; pub affects field/method visibility
	// Accessor fields use underscore prefix (e.g., _name) to avoid Go field/method name conflict
	hasContent := len(cls.Embeds) > 0 || len(allFields) > 0
	if hasContent {
		g.buf.WriteString(fmt.Sprintf("type %s struct {\n", className))
		for _, embed := range cls.Embeds {
			g.buf.WriteString("\t")
			g.buf.WriteString(embed)
			g.buf.WriteString("\n")
		}
		for _, field := range allFields {
			g.classFields[field.Name] = field.Type // Store field type (original name for lookup)
			// Use underscore prefix for accessor fields to avoid conflict with getter/setter methods
			goFieldName := field.Name
			if g.accessorFields[field.Name] {
				goFieldName = "_" + field.Name
			}
			if field.Type != "" {
				g.buf.WriteString(fmt.Sprintf("\t%s %s\n", goFieldName, mapType(field.Type)))
			} else {
				g.buf.WriteString(fmt.Sprintf("\t%s any\n", goFieldName))
			}
		}
		g.buf.WriteString("}\n\n")
	} else {
		g.buf.WriteString(fmt.Sprintf("type %s struct{}\n\n", className))
	}

	// Emit constructor if initialize exists
	hasInitialize := false
	for _, method := range cls.Methods {
		if method.Name == "initialize" {
			g.genConstructor(className, method, cls.Pub)
			hasInitialize = true
			break
		}
	}

	// If no initialize but has parent class, generate delegating constructor
	if !hasInitialize && len(cls.Embeds) > 0 {
		g.genSubclassConstructor(className, cls.Embeds[0], cls.Pub)
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
	// Internal field has underscore prefix to avoid conflict with method name
	internalField := "_" + acc.Name

	// Generate getter for "getter" and "property"
	if acc.Kind == "getter" || acc.Kind == "property" {
		// getter/property generates: func (r *T) name() T { return r._name }
		// Method name matches Rugby accessor name (camelCase for consistency)
		// Field has underscore prefix to avoid Go field/method name conflict
		var methodName string
		if pub {
			methodName = snakeToPascalWithAcronyms(acc.Name)
		} else {
			methodName = snakeToCamelWithAcronyms(acc.Name)
		}
		g.buf.WriteString(fmt.Sprintf("func (%s *%s) %s() %s {\n", recv, className, methodName, goType))
		g.buf.WriteString(fmt.Sprintf("\treturn %s.%s\n", recv, internalField))
		g.buf.WriteString("}\n\n")
	}

	// Generate setter for "setter" and "property"
	if acc.Kind == "setter" || acc.Kind == "property" {
		// setter/property generates: func (r *T) SetName(v T) { r._name = v }
		var methodName string
		if pub {
			methodName = "Set" + snakeToPascalWithAcronyms(acc.Name)
		} else {
			methodName = "set" + snakeToPascalWithAcronyms(acc.Name)
		}
		g.buf.WriteString(fmt.Sprintf("func (%s *%s) %s(v %s) {\n", recv, className, methodName, goType))
		g.buf.WriteString(fmt.Sprintf("\t%s.%s = v\n", recv, internalField))
		g.buf.WriteString("}\n\n")
	}
}

func (g *Generator) genInterfaceDecl(iface *ast.InterfaceDecl) {
	// Track this interface for zero-value generation
	g.interfaces[iface.Name] = true

	// Track interface methods for structural typing
	if g.interfaceMethods[iface.Name] == nil {
		g.interfaceMethods[iface.Name] = make(map[string]bool)
	}
	for _, method := range iface.Methods {
		g.interfaceMethods[iface.Name][method.Name] = true
	}

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

		// Return types - use mapParamType to convert class types to pointers
		if len(method.ReturnTypes) == 1 {
			g.buf.WriteString(" ")
			g.buf.WriteString(g.mapParamType(method.ReturnTypes[0]))
		} else if len(method.ReturnTypes) > 1 {
			g.buf.WriteString(" (")
			for i, rt := range method.ReturnTypes {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.buf.WriteString(g.mapParamType(rt))
			}
			g.buf.WriteString(")")
		}

		g.buf.WriteString("\n")
	}

	g.buf.WriteString("}\n")
}

func (g *Generator) genConstructor(className string, method *ast.MethodDecl, pub bool) {
	clear(g.vars)

	// Store constructor parameters for subclass delegation
	g.classConstructorParams[className] = method.Params

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
		// Use underscore prefix for accessor fields
		goFieldName := promoted.fieldName
		if g.accessorFields[promoted.fieldName] {
			goFieldName = "_" + promoted.fieldName
		}
		g.buf.WriteString(fmt.Sprintf("\t%s.%s = %s\n", recv, goFieldName, promoted.paramName))
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

// genSubclassConstructor generates a constructor for a subclass that doesn't
// define its own initialize method. It delegates to the parent class constructor.
func (g *Generator) genSubclassConstructor(className, parentClass string, pub bool) {
	recv := receiverName(className)

	// Only generate delegating constructor if parent is a known Rugby class
	// If parent is external (Go type), skip constructor generation - the struct embedding still works
	if !g.classes[parentClass] {
		return
	}

	// Look up parent constructor parameters (collected in pre-pass)
	parentParams := g.classConstructorParams[parentClass]

	// Constructor name: newClassName (private) or NewClassName (pub)
	constructorName := "new" + className
	if pub {
		constructorName = "New" + className
	}

	// Parent constructor name
	parentConstructorName := "new" + parentClass
	if g.pubClasses[parentClass] {
		parentConstructorName = "New" + parentClass
	}

	// Generate: func newSubclass(params...) *Subclass {
	g.buf.WriteString(fmt.Sprintf("func %s(", constructorName))

	// Use same parameters as parent constructor
	for i, param := range parentParams {
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

	// Initialize embedded parent using parent constructor
	// Dereference the parent constructor result to embed by value
	g.buf.WriteString(fmt.Sprintf("\t%s.%s = *%s(", recv, parentClass, parentConstructorName))
	for i, param := range parentParams {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		paramName := param.Name
		if len(paramName) > 0 && paramName[0] == '@' {
			paramName = paramName[1:]
		}
		g.buf.WriteString(paramName)
	}
	g.buf.WriteString(")\n")

	// Return instance
	g.buf.WriteString(fmt.Sprintf("\treturn %s\n", recv))
	g.buf.WriteString("}\n\n")
}

func (g *Generator) genMethodDecl(className string, method *ast.MethodDecl) {
	clear(g.vars)          // reset vars for each method
	clear(g.goInteropVars) // reset Go interop tracking for each method
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

	// message -> Error() string (satisfies error interface)
	// Only applies when message has no parameters
	if method.Name == "message" && len(method.Params) == 0 {
		g.buf.WriteString(fmt.Sprintf("func (%s *%s) Error() string {\n", recv, className))
		g.indent++
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
	// Methods required by interfaces must be exported (PascalCase) for Go interface satisfaction
	var methodName string
	if method.Pub || g.currentClassInterfaceMethods[method.Name] {
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
			g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, g.mapParamType(param.Type)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s any", param.Name))
		}
	}
	g.buf.WriteString(")")

	// Generate return type(s) if specified
	// Use mapParamType to properly convert class types to pointers
	if len(method.ReturnTypes) == 1 {
		g.buf.WriteString(" ")
		g.buf.WriteString(g.mapParamType(method.ReturnTypes[0]))
	} else if len(method.ReturnTypes) > 1 {
		g.buf.WriteString(" (")
		for i, rt := range method.ReturnTypes {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.buf.WriteString(g.mapParamType(rt))
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
