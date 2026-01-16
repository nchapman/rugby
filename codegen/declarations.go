package codegen

import (
	"fmt"
	"maps"
	"strings"

	"github.com/nchapman/rugby/ast"
)

// mapParamType converts Rugby types to Go types, making class types pointers.
// This handles nested generic types like Chan<Job> -> chan *Job and Array<Job> -> []*Job.
func (g *Generator) mapParamType(rubyType string) string {
	// Handle Array<T> - recursively convert element type
	if strings.HasPrefix(rubyType, "Array<") && strings.HasSuffix(rubyType, ">") {
		inner := rubyType[6 : len(rubyType)-1]
		return "[]" + g.mapParamType(inner)
	}
	// Handle Chan<T> - recursively convert element type
	if strings.HasPrefix(rubyType, "Chan<") && strings.HasSuffix(rubyType, ">") {
		inner := rubyType[5 : len(rubyType)-1]
		return "chan " + g.mapParamType(inner)
	}
	// Handle Map<K, V> - recursively convert key and value types
	if strings.HasPrefix(rubyType, "Map<") && strings.HasSuffix(rubyType, ">") {
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
	// Also check for module-scoped classes (e.g., Http_Response)
	// But NOT for structs - structs are value types and don't need pointers
	if (g.isClass(rubyType) || g.isModuleScopedClass(rubyType)) && !g.isStruct(rubyType) && !strings.HasPrefix(goType, "*") && !strings.HasPrefix(goType, "[]") && !strings.HasPrefix(goType, "map[") {
		return "*" + goType
	}

	return goType
}

// mapConstraint converts Rugby type constraints to Go constraint interfaces.
// Rugby constraints map to Go's comparable or custom interfaces.
func (g *Generator) mapConstraint(constraint string) string {
	// Handle multiple constraints: "A & B" -> "interface{ A; B }" (Go doesn't support this directly)
	// For now, just use the first constraint
	if strings.Contains(constraint, " & ") {
		parts := strings.Split(constraint, " & ")
		constraint = parts[0] // Use first constraint for now
	}

	switch constraint {
	case "Comparable", "Equatable":
		return "comparable"
	case "Hashable":
		return "comparable" // Go uses comparable for map keys
	case "Ordered":
		// Go's cmp.Ordered requires import - use constraints.Ordered
		g.needsConstraints = true
		return "constraints.Ordered"
	case "Numeric":
		// Custom interface for numeric types
		g.needsConstraints = true
		return "constraints.Integer | constraints.Float"
	default:
		// Assume it's a custom interface name
		return constraint
	}
}

func (g *Generator) genFuncDecl(fn *ast.FuncDecl) {
	clear(g.vars)          // reset vars for each function
	clear(g.goInteropVars) // reset Go interop tracking for each function
	g.currentReturnTypes = fn.ReturnTypes
	g.inMainFunc = fn.Name == "main"

	// Track type parameters for this function (for T.zero, etc.)
	g.currentFuncTypeParams = make(map[string]string)
	for _, tp := range fn.TypeParams {
		g.currentFuncTypeParams[tp.Name] = tp.Constraint
	}
	defer func() { g.currentFuncTypeParams = nil }()

	// Store the function declaration for default parameter lookup at call sites
	g.functions[fn.Name] = fn

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
	g.buf.WriteString(fmt.Sprintf("func %s", funcName))

	// Generate type parameters if present: [T any, U Constraint]
	if len(fn.TypeParams) > 0 {
		g.buf.WriteString("[")
		for i, tp := range fn.TypeParams {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.buf.WriteString(tp.Name)
			g.buf.WriteString(" ")
			if tp.Constraint != "" {
				g.buf.WriteString(g.mapConstraint(tp.Constraint))
			} else {
				g.buf.WriteString("any")
			}
		}
		g.buf.WriteString("]")
	}

	// Track destructuring params for body generation
	type destructureInfo struct {
		paramName string
		paramType string
		pairs     []ast.MapDestructurePair
	}
	var destructureParams []destructureInfo

	g.buf.WriteString("(")
	for i, param := range fn.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}

		// Handle destructuring parameters: {name:, age:} : Type
		if len(param.DestructurePairs) > 0 {
			paramName := fmt.Sprintf("_arg%d", i)
			destructureParams = append(destructureParams, destructureInfo{
				paramName: paramName,
				paramType: param.Type,
				pairs:     param.DestructurePairs,
			})
			g.buf.WriteString(fmt.Sprintf("%s %s", paramName, g.mapParamType(param.Type)))
			continue
		}

		if param.Type != "" {
			if param.Variadic {
				// Variadic parameter: *args : String -> ...string
				g.buf.WriteString(fmt.Sprintf("%s ...%s", param.Name, g.mapParamType(param.Type)))
			} else {
				g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, g.mapParamType(param.Type)))
			}
		} else {
			if param.Variadic {
				g.buf.WriteString(fmt.Sprintf("%s ...any", param.Name))
			} else {
				g.buf.WriteString(fmt.Sprintf("%s any", param.Name))
			}
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

	// Generate destructuring assignments at function start
	for _, dp := range destructureParams {
		for _, pair := range dp.pairs {
			g.writeIndent()
			// For structs: varname := _arg0.FieldName
			fieldName := snakeToPascalWithAcronyms(pair.Key)
			g.buf.WriteString(fmt.Sprintf("%s := %s.%s\n", pair.Variable, dp.paramName, fieldName))
			// Track the variable as declared
			g.vars[pair.Variable] = "" // Type unknown without deeper analysis
		}
	}

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
// It generates:
// 1. Nested classes as top-level types with namespaced names (Module_Class)
// 2. Module methods (def self.method) as package-level functions (Module_method)
// 3. A Go interface with the module's instance method signatures
func (g *Generator) genModuleDecl(mod *ast.ModuleDecl) {
	g.modules[mod.Name] = mod

	// Track module context for nested class naming
	g.currentModule = mod.Name

	// Generate nested classes with namespaced names (Module_Class)
	for _, cls := range mod.Classes {
		// Temporarily rename the class for code generation
		originalName := cls.Name
		cls.Name = mod.Name + "_" + cls.Name

		// Track the original class name for semantic analysis lookups
		g.currentOriginalClass = originalName

		// Track instance methods for the namespaced class
		if g.instanceMethods[cls.Name] == nil {
			g.instanceMethods[cls.Name] = make(map[string]string)
		}

		g.genClassDecl(cls)

		// Restore original name and clear tracking
		cls.Name = originalName
		g.currentOriginalClass = ""
	}

	// Generate class methods (def self.method) as package-level functions
	for _, method := range mod.Methods {
		if method.IsClassMethod {
			g.genModuleMethod(mod.Name, method)
		}
	}

	g.currentModule = ""

	// Generate an interface for the module's instance methods (for include)
	hasInstanceMethods := false
	for _, method := range mod.Methods {
		if !method.IsClassMethod {
			hasInstanceMethods = true
			break
		}
	}

	if !hasInstanceMethods {
		return
	}

	// Interface name is the module name (they're in separate namespaces in Rugby)
	g.buf.WriteString(fmt.Sprintf("type %s interface {\n", mod.Name))

	for _, method := range mod.Methods {
		if method.IsClassMethod {
			continue // Skip class methods in interface
		}
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

	g.buf.WriteString("}\n\n")
}

// genModuleMethod generates a module-level function (def self.method)
func (g *Generator) genModuleMethod(moduleName string, method *ast.MethodDecl) {
	clear(g.vars)
	clear(g.goInteropVars)
	g.currentReturnTypes = method.ReturnTypes

	g.emitLineDirective(method.Line)

	// Function name is Module_Method
	funcName := moduleName + "_" + snakeToPascalWithAcronyms(method.Name)
	g.buf.WriteString(fmt.Sprintf("func %s(", funcName))

	// Generate parameters
	for i, param := range method.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		// Handle @param promotion (strip @)
		paramName := param.Name
		if len(paramName) > 0 && paramName[0] == '@' {
			paramName = paramName[1:]
		}
		g.buf.WriteString(paramName)
		if param.Type != "" {
			g.buf.WriteString(" ")
			// Handle module-scoped types (e.g., Response -> Module_Response)
			paramType := g.resolveModuleScopedType(moduleName, param.Type)
			g.buf.WriteString(g.mapParamType(paramType))
		} else {
			g.buf.WriteString(" any")
		}
		g.vars[paramName] = param.Type
	}
	g.buf.WriteString(")")

	// Return type
	if len(method.ReturnTypes) == 1 {
		g.buf.WriteString(" ")
		retType := g.resolveModuleScopedType(moduleName, method.ReturnTypes[0])
		g.buf.WriteString(g.mapParamType(retType))
	} else if len(method.ReturnTypes) > 1 {
		g.buf.WriteString(" (")
		for i, rt := range method.ReturnTypes {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			retType := g.resolveModuleScopedType(moduleName, rt)
			g.buf.WriteString(g.mapParamType(retType))
		}
		g.buf.WriteString(")")
	}

	g.buf.WriteString(" {\n")

	// Generate function body
	g.indent++
	for i, stmt := range method.Body {
		isLast := i == len(method.Body)-1
		if isLast && len(method.ReturnTypes) > 0 {
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

	g.buf.WriteString("}\n\n")
	g.currentReturnTypes = nil
}

// resolveModuleScopedType resolves a type name that might be a module-scoped class
// (e.g., "Response" inside module "Http" becomes "Http_Response")
func (g *Generator) resolveModuleScopedType(moduleName, typeName string) string {
	if g.modules[moduleName] == nil {
		return typeName
	}

	mod := g.modules[moduleName]
	for _, cls := range mod.Classes {
		if cls.Name == typeName {
			return moduleName + "_" + typeName
		}
	}
	return typeName
}

// resolveClassName resolves a class name, handling module-scoped classes.
// If we're in a module context and the class is defined in that module,
// returns the namespaced name (e.g., "Response" -> "Http_Response").
// Otherwise returns the class name unchanged.
func (g *Generator) resolveClassName(className string) string {
	if g.currentModule == "" {
		return className
	}
	return g.resolveModuleScopedType(g.currentModule, className)
}

func (g *Generator) genClassDecl(cls *ast.ClassDecl) {
	className := cls.Name
	g.currentClass = className
	g.currentClassEmbeds = cls.Embeds
	g.currentClassTypeParamClause = ""
	g.currentClassTypeParamNames = ""
	clear(g.currentClassInterfaceMethods) // Reset for new class

	// For structural typing support: export methods that match ANY interface method
	// name in the program. In Go, interface methods must be exported (PascalCase),
	// so any class method that could satisfy an interface must also be exported.
	// This enables Go-style duck typing where classes can satisfy interfaces
	// without explicit 'implements' declarations.
	for _, ifaceName := range g.getAllInterfaceNames() {
		for _, methodName := range g.getInterfaceMethodNames(ifaceName) {
			g.currentClassInterfaceMethods[methodName] = true
		}
	}
	// Also include module methods (modules generate Go interfaces)
	for _, modName := range g.getAllModuleNames() {
		for _, methodName := range g.getModuleMethodNames(modName) {
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
		// Add module accessors with conflict resolution (last include wins)
		for _, acc := range mod.Accessors {
			if source, exists := accessorSources[acc.Name]; exists {
				// Conflict: accessor already defined
				// If class defined it, class wins (intentional override)
				if source == className {
					continue
				}
				// If another module defined it, replace with this version (last include wins)
				for i, existing := range allAccessors {
					if existing.Name == acc.Name {
						allAccessors = append(allAccessors[:i], allAccessors[i+1:]...)
						break
					}
				}
			}
			accessorSources[acc.Name] = fmt.Sprintf("module '%s'", modName)
			allAccessors = append(allAccessors, acc)
		}
		// Add module methods with conflict resolution (last include wins)
		for _, m := range mod.Methods {
			if source, exists := methodSources[m.Name]; exists {
				// Conflict: method already defined
				// If class defined it, class wins (intentional override)
				if source == className {
					// Class intentionally overrides module method - skip module version
					continue
				}
				// If another module defined it, replace with this version (last include wins)
				// Remove old method and add new one
				for i, existing := range allMethods {
					if existing.Name == m.Name {
						allMethods = append(allMethods[:i], allMethods[i+1:]...)
						break
					}
				}
			}
			methodSources[m.Name] = fmt.Sprintf("module '%s'", modName)
			allMethods = append(allMethods, m)
			// Mark as interface method so it's exported (PascalCase) to satisfy module interface
			g.currentClassInterfaceMethods[m.Name] = true
		}
	}

	// Also add underscore prefix for fields that have methods with the same name
	// This handles cases like: field @name + method def name -> ... (without accessor declarations)
	for _, m := range allMethods {
		if fieldNames[m.Name] && !g.accessorFields[m.Name] {
			g.accessorFields[m.Name] = true
		}
	}

	// Emit class variable declarations (package-level vars)
	for _, cv := range cls.ClassVars {
		varName := fmt.Sprintf("_%s_%s", className, cv.Name)
		g.buf.WriteString("var ")
		g.buf.WriteString(varName)
		g.buf.WriteString(" = ")
		g.genExpr(cv.Value)
		g.buf.WriteString("\n")
		// Track this class variable for later reference
		g.classVars[className+"@@"+cv.Name] = varName
	}
	if len(cls.ClassVars) > 0 {
		g.buf.WriteString("\n")
	}

	// Emit struct definition
	// Class names are already PascalCase by convention; pub affects field/method visibility
	// Accessor fields use underscore prefix (e.g., _name) to avoid Go field/method name conflict

	// Build type parameter names set for mapping field types
	var classTypeParamNames map[string]bool
	if len(cls.TypeParams) > 0 {
		classTypeParamNames = make(map[string]bool, len(cls.TypeParams))
		for _, tp := range cls.TypeParams {
			classTypeParamNames[tp.Name] = true
		}
	}

	// Generate type parameter clause if present
	typeParamClause := ""
	if len(cls.TypeParams) > 0 {
		var params []string
		for _, tp := range cls.TypeParams {
			if tp.Constraint != "" {
				params = append(params, fmt.Sprintf("%s %s", tp.Name, g.mapConstraint(tp.Constraint)))
			} else {
				params = append(params, tp.Name+" any")
			}
		}
		typeParamClause = "[" + strings.Join(params, ", ") + "]"
	}

	hasContent := len(cls.Embeds) > 0 || len(allFields) > 0
	if hasContent {
		g.buf.WriteString(fmt.Sprintf("type %s%s struct {\n", className, typeParamClause))
		for _, embed := range cls.Embeds {
			g.buf.WriteString("\t")
			g.buf.WriteString(embed)
			g.buf.WriteString("\n")
		}
		for _, field := range allFields {
			// Skip fields that are already accessors in parent classes
			// Child class should use inherited getter/setter instead of creating new field
			if g.isParentAccessor(cls.Embeds, field.Name) {
				continue
			}

			// Use underscore prefix for accessor fields to avoid conflict with getter/setter methods
			goFieldName := field.Name
			if g.isAccessorField(field.Name) {
				goFieldName = "_" + field.Name
			}
			// Determine field type: explicit annotation takes precedence,
			// then inferred type from semantic analysis, then fallback to any
			var goType string
			if field.Type != "" {
				goType = mapType(field.Type)
			} else if inferredType := g.typeInfo.GetFieldType(className, field.Name); inferredType != "" && inferredType != "unknown" {
				goType = mapType(inferredType)
			} else {
				goType = "any"
			}
			g.buf.WriteString(fmt.Sprintf("\t%s %s\n", goFieldName, goType))
		}
		g.buf.WriteString("}\n\n")
	} else {
		g.buf.WriteString(fmt.Sprintf("type %s%s struct{}\n\n", className, typeParamClause))
	}

	// Build type param names string for generic instantiation (e.g., "[T, U]")
	typeParamNames := ""
	if len(cls.TypeParams) > 0 {
		var names []string
		for _, tp := range cls.TypeParams {
			names = append(names, tp.Name)
		}
		typeParamNames = "[" + strings.Join(names, ", ") + "]"
	}

	// Store type params in generator fields for use by method generation
	g.currentClassTypeParamClause = typeParamClause
	g.currentClassTypeParamNames = typeParamNames

	// Emit constructor if initialize exists
	hasInitialize := false
	for _, method := range cls.Methods {
		if method.Name == "initialize" {
			g.genConstructor(className, method, cls.Pub, typeParamClause, typeParamNames)
			hasInitialize = true
			break
		}
	}

	// If no initialize but has parent class, generate delegating constructor
	if !hasInitialize && len(cls.Embeds) > 0 {
		g.genSubclassConstructor(className, cls.Embeds[0], cls.Pub, typeParamClause, typeParamNames)
	}

	// If no initialize and no parent, generate a simple default constructor
	// This enables instantiation of classes that only have methods (no fields)
	if !hasInitialize && len(cls.Embeds) == 0 {
		g.genDefaultConstructor(className, cls.Pub, typeParamClause, typeParamNames)
	}

	// Emit accessor methods (including from modules)
	// An accessor is exported if the class is pub OR the accessor is marked pub
	for _, acc := range allAccessors {
		g.genAccessorMethods(className, acc, cls.Pub || acc.Pub)
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

	// Emit module interface conformance checks
	// This suppresses "unused method" lint warnings for module methods
	// by asserting that the class satisfies each included module's interface
	for _, modName := range cls.Includes {
		mod := g.modules[modName]
		if mod != nil && len(mod.Methods) > 0 {
			g.buf.WriteString(fmt.Sprintf("var _ %s = (*%s)(nil)\n", modName, className))
		}
	}
	if len(cls.Includes) > 0 {
		g.buf.WriteString("\n")
	}

	// Emit method reference assertions for base classes
	// This suppresses "unused method" lint warnings for base class methods
	// that are shadowed (overridden) by subclasses. Format: var _ = (*ClassName).methodName
	if g.baseClasses[className] {
		// Reference methods
		for _, methodName := range g.baseClassMethods[className] {
			var goMethodName string
			if cls.Pub {
				goMethodName = snakeToPascalWithAcronyms(methodName)
			} else {
				goMethodName = snakeToCamelWithAcronyms(methodName)
			}
			g.buf.WriteString(fmt.Sprintf("var _ = (*%s).%s\n", className, goMethodName))
		}
		// Reference accessor methods (getters)
		for _, accName := range g.baseClassAccessors[className] {
			var goMethodName string
			if cls.Pub {
				goMethodName = snakeToPascalWithAcronyms(accName)
			} else {
				goMethodName = snakeToCamelWithAcronyms(accName)
			}
			g.buf.WriteString(fmt.Sprintf("var _ = (*%s).%s\n", className, goMethodName))
		}
		if len(g.baseClassMethods[className]) > 0 || len(g.baseClassAccessors[className]) > 0 {
			g.buf.WriteString("\n")
		}
	}

	g.currentClass = ""
	g.currentClassEmbeds = nil
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
		g.buf.WriteString(fmt.Sprintf("func (%s *%s%s) %s() %s {\n", recv, className, g.currentClassTypeParamNames, methodName, goType))
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
		g.buf.WriteString(fmt.Sprintf("func (%s *%s%s) %s(v %s) {\n", recv, className, g.currentClassTypeParamNames, methodName, goType))
		g.buf.WriteString(fmt.Sprintf("\t%s.%s = v\n", recv, internalField))
		g.buf.WriteString("}\n\n")
	}
}

func (g *Generator) genInterfaceDecl(iface *ast.InterfaceDecl) {
	// Generate type parameter clause if present (e.g., "[T any]")
	typeParamClause := ""
	if len(iface.TypeParams) > 0 {
		var params []string
		for _, tp := range iface.TypeParams {
			if tp.Constraint != "" {
				params = append(params, fmt.Sprintf("%s %s", tp.Name, g.mapConstraint(tp.Constraint)))
			} else {
				params = append(params, tp.Name+" any")
			}
		}
		typeParamClause = "[" + strings.Join(params, ", ") + "]"
	}

	// Generate: type InterfaceName[T any] interface { ... }
	g.buf.WriteString(fmt.Sprintf("type %s%s interface {\n", iface.Name, typeParamClause))

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

func (g *Generator) genConstructor(className string, method *ast.MethodDecl, pub bool, typeParamClause string, typeParamNames string) {
	clear(g.vars)

	// Set currentMethod so super calls are recognized as being inside a method
	previousMethod := g.currentMethod
	g.currentMethod = "initialize"
	defer func() { g.currentMethod = previousMethod }()

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

	// Constructor: func NewClassName[T any](params) *ClassName[T] (pub) or newClassName (non-pub)
	constructorName := "new" + className
	if pub {
		constructorName = "New" + className
	}
	g.buf.WriteString(fmt.Sprintf("func %s%s(", constructorName, typeParamClause))
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
	g.buf.WriteString(fmt.Sprintf(") *%s%s {\n", className, typeParamNames))

	// Create instance
	g.buf.WriteString(fmt.Sprintf("\t%s := &%s%s{}\n", recv, className, typeParamNames))

	// Auto-assign promoted parameters
	for _, promoted := range promotedParams {
		// Use underscore prefix for accessor fields
		goFieldName := promoted.fieldName
		if g.isAccessorField(promoted.fieldName) {
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
func (g *Generator) genSubclassConstructor(className, parentClass string, pub bool, typeParamClause string, typeParamNames string) {
	recv := receiverName(className)

	// Only generate delegating constructor if parent is a known Rugby class
	// If parent is external (Go type), skip constructor generation - the struct embedding still works
	if !g.isClass(parentClass) {
		return
	}

	// Look up parent constructor parameters via TypeInfo
	parentParams := g.getConstructorParams(parentClass)

	// Constructor name: newClassName (private) or NewClassName (pub)
	constructorName := "new" + className
	if pub {
		constructorName = "New" + className
	}

	// Parent constructor name
	parentConstructorName := "new" + parentClass
	if g.isPublicClass(parentClass) {
		parentConstructorName = "New" + parentClass
	}

	// Generate: func newSubclass[T any](params...) *Subclass[T] {
	g.buf.WriteString(fmt.Sprintf("func %s%s(", constructorName, typeParamClause))

	// Use same parameters as parent constructor
	for i, param := range parentParams {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		// Use field name (without @) for promoted parameters
		paramName := param[0]
		if len(paramName) > 0 && paramName[0] == '@' {
			paramName = paramName[1:]
		}
		paramType := param[1]
		if paramType != "" {
			g.buf.WriteString(fmt.Sprintf("%s %s", paramName, mapType(paramType)))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s any", paramName))
		}
	}
	g.buf.WriteString(fmt.Sprintf(") *%s%s {\n", className, typeParamNames))

	// Create instance
	g.buf.WriteString(fmt.Sprintf("\t%s := &%s%s{}\n", recv, className, typeParamNames))

	// Initialize embedded parent using parent constructor
	// Dereference the parent constructor result to embed by value
	g.buf.WriteString(fmt.Sprintf("\t%s.%s = *%s(", recv, parentClass, parentConstructorName))
	for i, param := range parentParams {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		paramName := param[0]
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

// genDefaultConstructor generates a zero-argument constructor for a class that has fields
// but no initialize method (e.g., classes that only include modules with fields).
func (g *Generator) genDefaultConstructor(className string, pub bool, typeParamClause string, typeParamNames string) {
	// Constructor name: newClassName (private) or NewClassName (pub)
	constructorName := "new" + className
	if pub {
		constructorName = "New" + className
	}

	// Generate: func newClassName[T any]() *ClassName[T] { return &ClassName[T]{} }
	g.buf.WriteString(fmt.Sprintf("func %s%s() *%s%s {\n", constructorName, typeParamClause, className, typeParamNames))
	g.buf.WriteString(fmt.Sprintf("\treturn &%s%s{}\n", className, typeParamNames))
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

	// Handle class methods (def self.method_name)
	if method.IsClassMethod {
		g.genClassMethod(className, method)
		return
	}

	// Receiver name: first letter of class, lowercase
	recv := receiverName(className)

	// Special method handling
	// to_s -> String() string (satisfies fmt.Stringer)
	// Only applies when to_s has no parameters
	if method.Name == "to_s" && len(method.Params) == 0 {
		g.buf.WriteString(fmt.Sprintf("func (%s *%s%s) String() string {\n", recv, className, g.currentClassTypeParamNames))
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

	// error or message -> Error() string (satisfies Go error interface)
	// Only applies when method has no parameters
	if (method.Name == "error" || method.Name == "message") && len(method.Params) == 0 {
		g.buf.WriteString(fmt.Sprintf("func (%s *%s%s) Error() string {\n", recv, className, g.currentClassTypeParamNames))
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
		g.buf.WriteString(fmt.Sprintf("func (%s *%s%s) Equal(other any) bool {\n", recv, className, g.currentClassTypeParamNames))
		g.indent++
		// Create type assertion for the parameter
		g.writeIndent()
		g.buf.WriteString(fmt.Sprintf("%s, ok := other.(*%s%s)\n", param.Name, className, g.currentClassTypeParamNames))
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
	// pub def -> PascalCase (exported)
	// private def -> _camelCase (private, underscore prefix)
	// def (default) -> camelCase (package-private)
	// Methods required by interfaces must be exported (PascalCase) for Go interface satisfaction
	// Special: setter methods (name=) -> SetName or setName
	var methodName string

	// Handle setter methods (ending in =)
	isSetter := strings.HasSuffix(method.Name, "=")
	baseName := strings.TrimSuffix(method.Name, "=")

	if method.Pub || g.currentClassInterfaceMethods[method.Name] {
		if isSetter {
			methodName = "Set" + snakeToPascalWithAcronyms(baseName)
		} else {
			methodName = snakeToPascalWithAcronyms(method.Name)
		}
	} else if method.Private {
		if isSetter {
			methodName = "_set" + snakeToPascalWithAcronyms(baseName)
		} else {
			methodName = "_" + snakeToCamelWithAcronyms(method.Name)
		}
	} else {
		if isSetter {
			methodName = "set" + snakeToPascalWithAcronyms(baseName)
		} else {
			methodName = snakeToCamelWithAcronyms(method.Name)
		}
	}

	// Check if method implicitly returns self (for method chaining)
	// Handles both bare `self` as last statement and explicit `return self`
	returnsSelf := false
	if len(method.ReturnTypes) == 0 && len(method.Body) > 0 {
		lastStmt := method.Body[len(method.Body)-1]

		// Check for bare `self` as last expression
		if exprStmt, ok := lastStmt.(*ast.ExprStmt); ok {
			if ident, ok := exprStmt.Expr.(*ast.Ident); ok && ident.Name == "self" {
				returnsSelf = true
			}
		}
		// Check for explicit `return self`
		if returnStmt, ok := lastStmt.(*ast.ReturnStmt); ok {
			if len(returnStmt.Values) == 1 {
				if ident, ok := returnStmt.Values[0].(*ast.Ident); ok && ident.Name == "self" {
					returnsSelf = true
				}
			}
		}
	}

	// Generate method signature with pointer receiver: func (r *ClassName[T]) MethodName(params) returns
	g.buf.WriteString(fmt.Sprintf("func (%s *%s%s) %s(", recv, className, g.currentClassTypeParamNames, methodName))
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
	if returnsSelf {
		// Method returns self for chaining - add *ClassName return type
		g.buf.WriteString(fmt.Sprintf(" *%s", className))
	} else if len(method.ReturnTypes) == 1 {
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
		if isLast && (len(method.ReturnTypes) > 0 || returnsSelf) {
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				// Implicit return (either declared return type or self for chaining)
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

// genClassMethod generates a class method (def self.method_name)
// Class methods are generated as package-level functions: func ClassName_MethodName(...)
func (g *Generator) genClassMethod(className string, method *ast.MethodDecl) {
	// Validate: methods ending in ? must return Bool
	if strings.HasSuffix(method.Name, "?") {
		if len(method.ReturnTypes) != 1 || method.ReturnTypes[0] != "Bool" {
			g.addError(fmt.Errorf("line %d: method '%s' ending in '?' must return Bool", method.Line, method.Name))
		}
	}

	// Method name: ClassName_methodName for unexported, ClassName_MethodName for exported
	var methodName string
	if method.Pub {
		methodName = className + "_" + snakeToPascalWithAcronyms(method.Name)
	} else {
		methodName = className + "_" + snakeToCamelWithAcronyms(method.Name)
	}

	// Track this class method for call resolution
	if g.classMethods == nil {
		g.classMethods = make(map[string]map[string]string)
	}
	if g.classMethods[className] == nil {
		g.classMethods[className] = make(map[string]string)
	}
	g.classMethods[className][method.Name] = methodName

	// Generate function signature: func ClassName_MethodName(params) returns
	g.buf.WriteString(fmt.Sprintf("func %s(", methodName))
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

	// Generate return type(s)
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

// genTypeAliasDecl generates a Go type alias from a Rugby type alias
// e.g., type UserID = Int64 becomes type UserID = int64
func (g *Generator) genTypeAliasDecl(typeAlias *ast.TypeAliasDecl) {
	goType := mapType(typeAlias.Type)
	g.buf.WriteString(fmt.Sprintf("type %s = %s\n\n", typeAlias.Name, goType))
}

// genEnumDecl generates a Go enum from a Rugby enum declaration
// e.g., enum Status { Pending, Active } becomes:
//
//	type Status int
//	const (
//	    StatusPending Status = iota
//	    StatusActive
//	)
func (g *Generator) genEnumDecl(enumDecl *ast.EnumDecl) {
	enumName := enumDecl.Name

	// Track this as an enum for later reference
	if g.enums == nil {
		g.enums = make(map[string]*ast.EnumDecl)
	}
	g.enums[enumName] = enumDecl

	// Generate type definition
	g.buf.WriteString(fmt.Sprintf("type %s int\n\n", enumName))

	// Check if any values have explicit assignments
	hasExplicitValues := false
	for _, v := range enumDecl.Values {
		if v.Value != nil {
			hasExplicitValues = true
			break
		}
	}

	// Generate constants
	g.buf.WriteString("const (\n")
	for i, v := range enumDecl.Values {
		constName := enumName + v.Name

		if hasExplicitValues {
			// Explicit values - generate each one with type and value
			g.buf.WriteString(fmt.Sprintf("\t%s %s", constName, enumName))
			if v.Value != nil {
				g.buf.WriteString(" = ")
				g.genExpr(v.Value)
			} else {
				// No value specified - use previous + 1 (iota behavior)
				g.buf.WriteString(" = iota")
				if i > 0 {
					g.buf.WriteString(fmt.Sprintf(" + %d", i))
				}
			}
		} else {
			// Auto-incrementing with iota
			// In Go, only the first constant needs the type and = iota
			if i == 0 {
				g.buf.WriteString(fmt.Sprintf("\t%s %s = iota", constName, enumName))
			} else {
				g.buf.WriteString(fmt.Sprintf("\t%s", constName))
			}
		}
		g.buf.WriteString("\n")
	}
	g.buf.WriteString(")\n\n")

	// Generate String() method for .to_s
	g.buf.WriteString(fmt.Sprintf("func (e %s) String() string {\n", enumName))
	g.buf.WriteString("\tswitch e {\n")
	for _, v := range enumDecl.Values {
		constName := enumName + v.Name
		g.buf.WriteString(fmt.Sprintf("\tcase %s:\n", constName))
		g.buf.WriteString(fmt.Sprintf("\t\treturn %q\n", v.Name))
	}
	g.buf.WriteString("\t}\n")
	g.buf.WriteString("\treturn \"\"\n")
	g.buf.WriteString("}\n\n")

	// Generate <EnumName>Values() function for .values
	g.buf.WriteString(fmt.Sprintf("func %sValues() []%s {\n", enumName, enumName))
	g.buf.WriteString(fmt.Sprintf("\treturn []%s{", enumName))
	for i, v := range enumDecl.Values {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(enumName + v.Name)
	}
	g.buf.WriteString("}\n")
	g.buf.WriteString("}\n\n")

	// Generate <EnumName>FromString() function for .from_string
	// Returns *EnumName (nil if not found) for optional semantics
	g.buf.WriteString(fmt.Sprintf("func %sFromString(s string) *%s {\n", enumName, enumName))
	g.buf.WriteString("\tswitch s {\n")
	for _, v := range enumDecl.Values {
		constName := enumName + v.Name
		g.buf.WriteString(fmt.Sprintf("\tcase %q:\n", v.Name))
		g.buf.WriteString(fmt.Sprintf("\t\tv := %s\n", constName))
		g.buf.WriteString("\t\treturn &v\n")
	}
	g.buf.WriteString("\t}\n")
	g.buf.WriteString("\treturn nil\n")
	g.buf.WriteString("}\n\n")
}

// genStructDecl generates a Go struct from a Rugby struct declaration
// Structs are value types with:
// - Auto-generated field getters (via direct field access)
// - Auto-generated String() method for printing
// - Value semantics (copy on assignment)
// - Immutability (no field assignment allowed - enforced by semantic analyzer)
func (g *Generator) genStructDecl(structDecl *ast.StructDecl) {
	structName := structDecl.Name

	// Track this as a struct for later reference
	if g.structs == nil {
		g.structs = make(map[string]*ast.StructDecl)
	}
	g.structs[structName] = structDecl

	// Generate type parameters if generic
	typeParamStr := ""
	if len(structDecl.TypeParams) > 0 {
		var params []string
		for _, tp := range structDecl.TypeParams {
			if tp.Constraint != "" {
				params = append(params, fmt.Sprintf("%s %s", tp.Name, g.mapConstraint(tp.Constraint)))
			} else {
				params = append(params, tp.Name+" any")
			}
		}
		typeParamStr = "[" + strings.Join(params, ", ") + "]"
	}

	// Generate struct type definition
	g.buf.WriteString(fmt.Sprintf("type %s%s struct {\n", structName, typeParamStr))
	for _, field := range structDecl.Fields {
		fieldName := snakeToPascalWithAcronyms(field.Name) // Public field for direct access
		fieldType := mapType(field.Type)
		g.buf.WriteString(fmt.Sprintf("\t%s %s\n", fieldName, fieldType))
	}
	g.buf.WriteString("}\n\n")

	// Generate String() method for printing
	g.needsFmt = true
	g.buf.WriteString(fmt.Sprintf("func (s %s%s) String() string {\n", structName, typeParamStr))
	g.buf.WriteString(fmt.Sprintf("\treturn fmt.Sprintf(\"%s{", structName))
	for i, field := range structDecl.Fields {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		// Use %v for general formatting, %q for strings
		if field.Type == "String" {
			g.buf.WriteString(fmt.Sprintf("%s: %%q", field.Name))
		} else {
			g.buf.WriteString(fmt.Sprintf("%s: %%v", field.Name))
		}
	}
	g.buf.WriteString("}\"")
	for _, field := range structDecl.Fields {
		fieldName := snakeToPascalWithAcronyms(field.Name)
		g.buf.WriteString(fmt.Sprintf(", s.%s", fieldName))
	}
	g.buf.WriteString(")\n")
	g.buf.WriteString("}\n\n")

	// Generate methods if any
	for _, method := range structDecl.Methods {
		g.genStructMethod(structDecl, method)
	}
}

// genStructMethod generates a method for a struct with value receiver
func (g *Generator) genStructMethod(structDecl *ast.StructDecl, method *ast.MethodDecl) {
	structName := structDecl.Name
	methodName := snakeToPascalWithAcronyms(method.Name)

	// Save old state
	oldVars := g.vars
	oldReturnTypes := g.currentReturnTypes
	g.vars = maps.Clone(oldVars)

	// Type parameters - for struct methods, we only need the type names (e.g., "[T]"), not the constraints
	typeParamStr := ""
	if len(structDecl.TypeParams) > 0 {
		var names []string
		for _, tp := range structDecl.TypeParams {
			names = append(names, tp.Name)
		}
		typeParamStr = "[" + strings.Join(names, ", ") + "]"
	}

	// Method receiver (value receiver, not pointer)
	g.buf.WriteString(fmt.Sprintf("func (s %s%s) %s(", structName, typeParamStr, methodName))

	// Parameters - add to vars map
	for i, param := range method.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		paramName := param.Name
		paramType := mapType(param.Type)
		g.buf.WriteString(fmt.Sprintf("%s %s", paramName, paramType))
		g.vars[paramName] = param.Type // Track parameter type (Rugby type, not Go type)
	}
	g.buf.WriteString(")")

	// Return type
	if len(method.ReturnTypes) > 0 {
		g.buf.WriteString(" ")
		if len(method.ReturnTypes) == 1 {
			g.buf.WriteString(mapType(method.ReturnTypes[0]))
		} else {
			g.buf.WriteString("(")
			for i, rt := range method.ReturnTypes {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.buf.WriteString(mapType(rt))
			}
			g.buf.WriteString(")")
		}
	}

	g.buf.WriteString(" {\n")
	g.indent++

	// Set up method context
	g.currentStruct = structDecl
	g.currentReturnTypes = method.ReturnTypes

	// Generate method body with implicit return for last expression
	hasReturnType := len(method.ReturnTypes) > 0
	for i, stmt := range method.Body {
		isLast := i == len(method.Body)-1
		if isLast && hasReturnType {
			// If last statement is an expression, add implicit return
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

	// Restore state
	g.currentStruct = nil
	g.currentReturnTypes = oldReturnTypes
	g.vars = oldVars
	g.indent--
	g.buf.WriteString("}\n\n")
}

// genConstDecl generates a Go const declaration from a Rugby const declaration
// e.g., const MAX_SIZE = 1024 becomes const MAX_SIZE = 1024
func (g *Generator) genConstDecl(constDecl *ast.ConstDecl) {
	g.writeIndent()
	g.buf.WriteString("const ")
	g.buf.WriteString(constDecl.Name)

	// Add type annotation if specified
	if constDecl.Type != "" {
		g.buf.WriteString(" ")
		g.buf.WriteString(mapType(constDecl.Type))
	}

	g.buf.WriteString(" = ")
	g.genExpr(constDecl.Value)
	g.buf.WriteString("\n")
}
