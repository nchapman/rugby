package codegen

// Call expression and selector expression code generation.
// This file handles:
//   - Function and method calls (genCallExpr)
//   - Selector expressions (obj.field, obj.method)
//   - Go interop detection and handling
//   - Standard library property methods (x.abs, s.upcase, etc.)
//   - Array property methods (arr.first, arr.last, etc.)

import (
	"fmt"
	"slices"
	"strings"

	"github.com/nchapman/rugby/ast"
)

func (g *Generator) genCallExpr(call *ast.CallExpr) {
	if call.Block != nil {
		g.genBlockCall(call)
		return
	}

	// Check if receiver is an optional type with lambda argument - needs special handling
	if sel, ok := call.Func.(*ast.SelectorExpr); ok {
		receiverType := g.inferTypeFromExpr(sel.X)
		if isOptionalType(receiverType) && len(call.Args) >= 1 {
			if lambda, ok := call.Args[0].(*ast.LambdaExpr); ok {
				switch sel.Sel {
				case "map":
					g.genOptionalMapLambda(sel.X, lambda, receiverType)
					return
				case "flat_map":
					g.genOptionalFlatMapLambda(sel.X, lambda, receiverType)
					return
				case "filter":
					g.genOptionalFilterLambda(sel.X, lambda, receiverType)
					return
				case "each":
					g.genOptionalEachLambda(sel.X, lambda, receiverType)
					return
				}
			}
		}
	}

	if sel, ok := call.Func.(*ast.SelectorExpr); ok {
		if bm, isBlockMethod := blockMethods[sel.Sel]; isBlockMethod {
			if len(call.Args) >= 1 {
				if stp, ok := call.Args[0].(*ast.SymbolToProcExpr); ok {
					g.genSymbolToProcBlockCall(sel.X, stp, bm)
					return
				}
				// Handle lambda argument for iteration methods
				if lambda, ok := call.Args[0].(*ast.LambdaExpr); ok {
					g.genLambdaIterationCall(sel.X, sel.Sel, lambda, call.Args[1:])
					return
				}
				// Handle reduce with initial value + lambda
				if bm.hasAccumulator && len(call.Args) >= 2 {
					if lambda, ok := call.Args[1].(*ast.LambdaExpr); ok {
						g.genLambdaIterationCall(sel.X, sel.Sel, lambda, call.Args[:1])
						return
					}
				}
			}
		}
		// Handle each/each_with_index with lambda argument
		if sel.Sel == "each" || sel.Sel == "each_with_index" {
			if len(call.Args) >= 1 {
				if lambda, ok := call.Args[0].(*ast.LambdaExpr); ok {
					g.genLambdaIterationCall(sel.X, sel.Sel, lambda, call.Args[1:])
					return
				}
			}
		}
		// Handle times with lambda argument (n.times -> { |i| ... })
		if sel.Sel == "times" {
			if len(call.Args) >= 1 {
				if lambda, ok := call.Args[0].(*ast.LambdaExpr); ok {
					g.genLambdaIterationCall(sel.X, sel.Sel, lambda, call.Args[1:])
					return
				}
			}
		}
		// Handle upto/downto with lambda argument (n.upto(m) -> { |i| ... })
		if sel.Sel == "upto" || sel.Sel == "downto" {
			if len(call.Args) >= 2 {
				if lambda, ok := call.Args[1].(*ast.LambdaExpr); ok {
					g.genLambdaIterationCall(sel.X, sel.Sel, lambda, call.Args[:1])
					return
				}
			}
		}
		// Handle scope.spawn(lambda) -> runtime.ScopeSpawn(scope, lambda)
		if sel.Sel == "spawn" {
			receiverType := g.inferTypeFromExpr(sel.X)
			if receiverType == "Scope" || receiverType == "*Scope" {
				if len(call.Args) >= 1 {
					if lambda, ok := call.Args[0].(*ast.LambdaExpr); ok {
						g.genScopedSpawnLambda(sel.X, lambda)
						return
					}
				}
			}
		}
	}

	if sel, ok := call.Func.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			if ident.Name == "assert" || ident.Name == "require" {
				g.genTestAssertion(call, ident.Name, sel.Sel)
				return
			}
			// Handle enum type methods: Color.from_string("Red"), Color.values
			if g.enums != nil {
				if _, isEnum := g.enums[ident.Name]; isEnum {
					switch sel.Sel {
					case "from_string":
						// Color.from_string("Red") -> ColorFromString("Red")
						g.buf.WriteString(ident.Name)
						g.buf.WriteString("FromString(")
						for i, arg := range call.Args {
							if i > 0 {
								g.buf.WriteString(", ")
							}
							g.genExpr(arg)
						}
						g.buf.WriteString(")")
						return
					case "values":
						// Color.values -> ColorValues()
						g.buf.WriteString(ident.Name)
						g.buf.WriteString("Values()")
						return
					}
				}
			}
		}
	}

	switch fn := call.Func.(type) {
	case *ast.Ident:
		funcName := fn.Name

		if funcName == "error_is?" {
			if len(call.Args) != 2 {
				g.addError(fmt.Errorf("error_is? requires 2 arguments (err, target), got %d", len(call.Args)))
				g.buf.WriteString("false")
				return
			}
			g.needsErrors = true
			g.buf.WriteString("errors.Is(")
			g.genExpr(call.Args[0])
			g.buf.WriteString(", ")
			g.genExpr(call.Args[1])
			g.buf.WriteString(")")
			return
		}

		if funcName == "error_as" {
			if len(call.Args) != 2 {
				g.addError(fmt.Errorf("error_as requires 2 arguments (err, Type), got %d", len(call.Args)))
				g.buf.WriteString("nil")
				return
			}
			g.needsErrors = true
			g.buf.WriteString("func() *")
			g.genExpr(call.Args[1])
			g.buf.WriteString(" { var _target *")
			g.genExpr(call.Args[1])
			g.buf.WriteString("; if errors.As(")
			g.genExpr(call.Args[0])
			g.buf.WriteString(", &_target) { return _target }; return nil }()")
			return
		}

		// Check for module method call within class context
		// (e.g., touch(now) in a class that includes a module with touch method)
		if g.currentClass() != "" && g.isModuleMethod(funcName) {
			recv := receiverName(g.currentClass())
			methodName := snakeToPascalWithAcronyms(funcName)
			g.buf.WriteString(fmt.Sprintf("%s.%s(", recv, methodName))
			for i, arg := range call.Args {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.genExpr(arg)
			}
			g.buf.WriteString(")")
			return
		}

		// Check for private method call within class context
		// (e.g., encrypt_password() in a class that has private def encrypt_password)
		if g.currentClass() != "" && g.privateMethods[g.currentClass()] != nil && g.privateMethods[g.currentClass()][funcName] {
			recv := receiverName(g.currentClass())
			methodName := "_" + snakeToCamelWithAcronyms(funcName)
			g.buf.WriteString(fmt.Sprintf("%s.%s(", recv, methodName))
			for i, arg := range call.Args {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				g.genExpr(arg)
			}
			g.buf.WriteString(")")
			return
		}

		// Check for keyword arguments: func(name: value, ...)
		if fnDecl, ok := g.functions[funcName]; ok && g.hasKeywordArgs(call.Args) {
			if g.pubFuncs[funcName] {
				funcName = snakeToPascalWithAcronyms(funcName)
			} else {
				funcName = snakeToCamelWithAcronyms(funcName)
			}
			g.buf.WriteString(funcName)
			g.genKeywordCallArgs(fnDecl, call.Args)
			return
		}

		// Check for user-defined function with default parameters
		if fnDecl, ok := g.functions[funcName]; ok && len(call.Args) < len(fnDecl.Params) {
			// We need to fill in default values for missing arguments
			if g.pubFuncs[funcName] {
				funcName = snakeToPascalWithAcronyms(funcName)
			} else {
				funcName = snakeToCamelWithAcronyms(funcName)
			}
			g.buf.WriteString(funcName)
			g.buf.WriteString("(")
			for i, param := range fnDecl.Params {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				if i < len(call.Args) {
					// Use provided argument
					g.genExpr(call.Args[i])
				} else if param.DefaultValue != nil {
					// Use default value
					g.genExpr(param.DefaultValue)
				} else {
					// No default - this would be an error, but semantic analysis should catch it
					g.buf.WriteString("nil")
				}
			}
			g.buf.WriteString(")")
			return
		}

		if kf, ok := kernelFuncs[funcName]; ok {
			g.needsRuntime = true
			if kf.transform != nil {
				funcName = kf.transform(call.Args)
			} else {
				funcName = kf.runtimeFunc
			}
		} else if g.pubFuncs[funcName] {
			// Function declared with 'pub' -> PascalCase
			funcName = snakeToPascalWithAcronyms(funcName)
		} else {
			// Regular function -> camelCase
			funcName = snakeToCamelWithAcronyms(funcName)
		}
		g.buf.WriteString(funcName)
	case *ast.SelectorExpr:
		if fn.Sel == "new" {
			if indexExpr, ok := fn.X.(*ast.IndexExpr); ok {
				if chanIdent, ok := indexExpr.Left.(*ast.Ident); ok && chanIdent.Name == "Chan" {
					var chanType string
					if typeIdent, ok := indexExpr.Index.(*ast.Ident); ok {
						chanType = typeIdent.Name
					} else {
						chanType = "any"
					}
					// Use mapParamType to properly convert class types to pointers
					goType := g.mapParamType(chanType)

					g.buf.WriteString("make(chan ")
					g.buf.WriteString(goType)
					if len(call.Args) > 0 {
						g.buf.WriteString(", ")
						g.genExpr(call.Args[0])
					}
					g.buf.WriteString(")")
					return
				}
				// Array<T>.new(size, default) -> runtime.Fill[T](size, default)
				// Array<T>.new(size) -> make([]T, size)
				if arrIdent, ok := indexExpr.Left.(*ast.Ident); ok && arrIdent.Name == "Array" {
					var elemType string
					if typeIdent, ok := indexExpr.Index.(*ast.Ident); ok {
						elemType = typeIdent.Name
					} else {
						elemType = "any"
					}
					goType := mapType(elemType)

					if len(call.Args) >= 2 {
						// Check if default is a zero value - can use make() instead of Fill
						if isZeroValue(call.Args[1]) {
							// Array<T>.new(size, zero) -> make([]T, size)
							g.buf.WriteString("make([]")
							g.buf.WriteString(goType)
							g.buf.WriteString(", ")
							g.genExpr(call.Args[0]) // size
							g.buf.WriteString(")")
						} else {
							// Array<T>.new(size, default) -> runtime.Fill[T](default, size)
							g.needsRuntime = true
							g.buf.WriteString("runtime.Fill[")
							g.buf.WriteString(goType)
							g.buf.WriteString("](")
							g.genExpr(call.Args[1]) // default value
							g.buf.WriteString(", ")
							g.genExpr(call.Args[0]) // size
							g.buf.WriteString(")")
						}
					} else if len(call.Args) == 1 {
						// Array<T>.new(size) -> make([]T, size)
						g.buf.WriteString("make([]")
						g.buf.WriteString(goType)
						g.buf.WriteString(", ")
						g.genExpr(call.Args[0])
						g.buf.WriteString(")")
					} else {
						// Array<T>.new() -> []T{}
						g.buf.WriteString("[]")
						g.buf.WriteString(goType)
						g.buf.WriteString("{}")
					}
					return
				}
			}
		}

		if fn.Sel == "new" {
			// Handle bare Array.new(size, default) or Array.new(size) with inferred type
			if ident, ok := fn.X.(*ast.Ident); ok && ident.Name == "Array" {
				if len(call.Args) >= 2 {
					// Get inferred element type from semantic analysis
					elemType := g.typeInfo.GetElementType(call)
					if elemType != "" && elemType != "unknown" {
						// Use mapParamType to handle class types as pointers
						goType := g.mapParamType(elemType)
						// Check if default is a zero value - can use make() instead of Fill
						if isZeroValue(call.Args[1]) {
							// Array.new(size, zero) -> make([]T, size)
							g.buf.WriteString("make([]")
							g.buf.WriteString(goType)
							g.buf.WriteString(", ")
							g.genExpr(call.Args[0]) // size
							g.buf.WriteString(")")
						} else {
							g.needsRuntime = true
							g.buf.WriteString("runtime.Fill[")
							g.buf.WriteString(goType)
							g.buf.WriteString("](")
							g.genExpr(call.Args[1]) // default value
							g.buf.WriteString(", ")
							g.genExpr(call.Args[0]) // size
							g.buf.WriteString(")")
						}
						return
					}
					// Fallback: use []any when type inference fails
					if isZeroValue(call.Args[1]) {
						g.buf.WriteString("make([]any, ")
						g.genExpr(call.Args[0])
						g.buf.WriteString(")")
					} else {
						g.needsRuntime = true
						g.buf.WriteString("runtime.Fill[any](")
						g.genExpr(call.Args[1])
						g.buf.WriteString(", ")
						g.genExpr(call.Args[0])
						g.buf.WriteString(")")
					}
					return
				} else if len(call.Args) == 1 {
					// Array.new(size) -> make([]any, size)
					g.buf.WriteString("make([]any, ")
					g.genExpr(call.Args[0])
					g.buf.WriteString(")")
					return
				}
			}
			// Only treat as Rugby class constructor if NOT a Go import
			// (errors.new should become errors.New, not newerrors)
			if ident, ok := fn.X.(*ast.Ident); ok && !g.imports[ident.Name] {
				// Resolve module-scoped class name (e.g., Response -> Http_Response when inside module Http)
				className := g.resolveClassName(ident.Name)
				var ctorName string
				if g.isPublicClass(ident.Name) {
					ctorName = fmt.Sprintf("New%s", className)
				} else {
					ctorName = fmt.Sprintf("new%s", className)
				}
				g.buf.WriteString(ctorName)
				g.buf.WriteString("(")
				for i, arg := range call.Args {
					if i > 0 {
						g.buf.WriteString(", ")
					}
					g.genExpr(arg)
				}
				g.buf.WriteString(")")
				return
			}
			// Handle scoped class constructors like Http::Response.new(...)
			// Generate: newHttp_Response(...)
			if scope, ok := fn.X.(*ast.ScopeExpr); ok {
				className := g.scopeExprToClassName(scope)
				var ctorName string
				// For scoped classes, check if the module is public
				if g.isPublicClass(className) {
					ctorName = "New" + className
				} else {
					ctorName = "new" + className
				}
				g.buf.WriteString(ctorName)
				g.buf.WriteString("(")
				for i, arg := range call.Args {
					if i > 0 {
						g.buf.WriteString(", ")
					}
					g.genExpr(arg)
				}
				g.buf.WriteString(")")
				return
			}
			// Handle Go package type .new calls like sync.WaitGroup.new
			// Generate zero-value initialization: &pkg.Type{}
			if sel, ok := fn.X.(*ast.SelectorExpr); ok {
				if pkgIdent, ok := sel.X.(*ast.Ident); ok && g.imports[pkgIdent.Name] {
					g.buf.WriteString("&")
					g.genExpr(fn.X)
					g.buf.WriteString("{}")
					return
				}
			}
		}

		// Handle class method calls (ClassName.method_name)
		if ident, ok := fn.X.(*ast.Ident); ok && g.isClass(ident.Name) {
			if methodMap, hasClass := g.classMethods[ident.Name]; hasClass {
				if funcName, hasMethod := methodMap[fn.Sel]; hasMethod {
					g.buf.WriteString(funcName)
					g.buf.WriteString("(")
					for i, arg := range call.Args {
						if i > 0 {
							g.buf.WriteString(", ")
						}
						g.genExpr(arg)
					}
					g.buf.WriteString(")")
					return
				}
			}
		}

		// Handle module method calls (ModuleName.method_name -> ModuleName_Method)
		if ident, ok := fn.X.(*ast.Ident); ok {
			if mod, hasModule := g.modules[ident.Name]; hasModule {
				// Check if the module has this class method
				for _, method := range mod.Methods {
					if method.IsClassMethod && method.Name == fn.Sel {
						// Generate: ModuleName_MethodName(args)
						funcName := ident.Name + "_" + snakeToPascalWithAcronyms(fn.Sel)
						g.buf.WriteString(funcName)
						g.buf.WriteString("(")
						for i, arg := range call.Args {
							if i > 0 {
								g.buf.WriteString(", ")
							}
							g.genExpr(arg)
						}
						g.buf.WriteString(")")
						return
					}
				}
			}
		}

		if fn.Sel == "close" {
			// Check if receiver is a channel - use close() builtin
			// For Go interop types (like *os.File), use .Close() method
			recvType := g.inferTypeFromExpr(fn.X)
			isChannel := strings.HasPrefix(recvType, "Chan<") || strings.HasPrefix(recvType, "chan ")
			// Also check TypeInfo for channel detection
			if !isChannel && g.typeInfo != nil {
				isChannel = g.typeInfo.GetTypeKind(fn.X) == TypeChannel
			}
			// Fallback: if type is unknown and not a Go interop var, treat as channel
			if !isChannel && recvType == "" {
				if ident, ok := fn.X.(*ast.Ident); ok && !g.goInteropVars[ident.Name] && !g.imports[ident.Name] {
					isChannel = true
				}
			}
			if isChannel {
				g.buf.WriteString("close(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			}
			// For Go interop types, fall through to normal method handling
			// which will convert to .Close()
		}
		if fn.Sel == "receive" {
			g.buf.WriteString("<- ")
			g.genExpr(fn.X)
			return
		}
		if fn.Sel == "try_receive" {
			g.needsRuntime = true
			g.buf.WriteString("runtime.TryReceivePtr(")
			g.genExpr(fn.X)
			g.buf.WriteString(")")
			return
		}

		// Handle array/slice method calls FIRST, before stdLib lookup
		recvTypeForArray := g.inferTypeFromExpr(fn.X)
		if strings.HasPrefix(recvTypeForArray, "[]") || recvTypeForArray == "Array" || strings.HasPrefix(recvTypeForArray, "Array<") {
			elemType := g.inferArrayElementGoType(fn.X)
			switch fn.Sel {
			case "any?":
				// any? without a block returns true if array has any elements
				g.buf.WriteString("(len(")
				g.genExpr(fn.X)
				g.buf.WriteString(") > 0)")
				return
			case "empty?":
				// empty? returns true if array has no elements
				g.buf.WriteString("(len(")
				g.genExpr(fn.X)
				g.buf.WriteString(") == 0)")
				return
			case "shift":
				// shift removes and returns the first element, needs pointer
				g.needsRuntime = true
				switch elemType {
				case "int":
					g.buf.WriteString("runtime.ShiftInt(&")
				case "string":
					g.buf.WriteString("runtime.ShiftString(&")
				case "float64":
					g.buf.WriteString("runtime.ShiftFloat(&")
				default:
					g.buf.WriteString("runtime.ShiftAny(&")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "pop":
				// pop removes and returns the last element, needs pointer
				g.needsRuntime = true
				switch elemType {
				case "int":
					g.buf.WriteString("runtime.PopInt(&")
				case "string":
					g.buf.WriteString("runtime.PopString(&")
				case "float64":
					g.buf.WriteString("runtime.PopFloat(&")
				default:
					g.buf.WriteString("runtime.PopAny(&")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "push":
				// push appends an element to the end, needs pointer
				if len(call.Args) != 1 {
					g.addError(fmt.Errorf("push requires exactly 1 argument"))
					return
				}
				g.needsRuntime = true
				switch elemType {
				case "int":
					g.buf.WriteString("runtime.PushInt(&")
				case "string":
					g.buf.WriteString("runtime.PushString(&")
				case "float64":
					g.buf.WriteString("runtime.PushFloat(&")
				default:
					g.buf.WriteString("runtime.PushAny(&")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")")
				return
			case "unshift":
				// unshift prepends an element to the beginning, needs pointer
				if len(call.Args) != 1 {
					g.addError(fmt.Errorf("unshift requires exactly 1 argument"))
					return
				}
				g.needsRuntime = true
				switch elemType {
				case "int":
					g.buf.WriteString("runtime.UnshiftInt(&")
				case "string":
					g.buf.WriteString("runtime.UnshiftString(&")
				case "float64":
					g.buf.WriteString("runtime.UnshiftFloat(&")
				default:
					g.buf.WriteString("runtime.UnshiftAny(&")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")")
				return
			case "sum":
				g.needsRuntime = true
				if elemType == "float64" {
					g.buf.WriteString("runtime.SumFloat(")
				} else {
					g.buf.WriteString("runtime.SumInt(")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "min":
				g.needsRuntime = true
				if elemType == "float64" {
					g.buf.WriteString("runtime.MinFloatPtr(")
				} else {
					g.buf.WriteString("runtime.MinIntPtr(")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "max":
				g.needsRuntime = true
				if elemType == "float64" {
					g.buf.WriteString("runtime.MaxFloatPtr(")
				} else {
					g.buf.WriteString("runtime.MaxIntPtr(")
				}
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "first":
				g.needsRuntime = true
				g.buf.WriteString("runtime.FirstPtr(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "last":
				g.needsRuntime = true
				g.buf.WriteString("runtime.LastPtr(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "sorted":
				g.needsRuntime = true
				g.buf.WriteString("runtime.Sort(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "include?", "contains?":
				g.needsRuntime = true
				g.buf.WriteString("runtime.Contains(")
				g.genExpr(fn.X)
				if len(call.Args) > 0 {
					g.buf.WriteString(", ")
					g.genExpr(call.Args[0])
				}
				g.buf.WriteString(")")
				return
			case "length", "size":
				// Array length uses Go's built-in len()
				g.buf.WriteString("len(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			}
		}

		// Handle set method calls: s.include?(x), s.add(x), s.delete(x), s.size, s.to_a, s.clear
		recvTypeForSet := g.inferTypeFromExpr(fn.X)
		if recvTypeForSet == "Set" || strings.HasPrefix(recvTypeForSet, "Set<") {
			switch fn.Sel {
			case "include?":
				if len(call.Args) != 1 {
					g.addError(fmt.Errorf("include? requires exactly 1 argument"))
					g.buf.WriteString("false")
					return
				}
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetContains(")
				g.genExpr(fn.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")")
				return
			case "add":
				if len(call.Args) != 1 {
					g.addError(fmt.Errorf("add requires exactly 1 argument"))
					return
				}
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetAdd(")
				g.genExpr(fn.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")")
				return
			case "delete":
				if len(call.Args) != 1 {
					g.addError(fmt.Errorf("delete requires exactly 1 argument"))
					return
				}
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetDelete(")
				g.genExpr(fn.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")")
				return
			case "size", "length", "count":
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetSize(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "to_a":
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetToArray(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "clear":
				g.needsRuntime = true
				g.buf.WriteString("runtime.SetClear(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			case "empty?":
				g.buf.WriteString("(len(")
				g.genExpr(fn.X)
				g.buf.WriteString(") == 0)")
				return
			case "any?":
				g.buf.WriteString("(len(")
				g.genExpr(fn.X)
				g.buf.WriteString(") > 0)")
				return
			}
		}

		recvType := g.inferTypeFromExpr(fn.X)
		// Normalize Go types to Rugby type names for stdLib lookup
		lookupType := recvType
		switch recvType {
		case "int", "int64":
			lookupType = "Int"
		case "float64":
			lookupType = "Float"
		case "string":
			lookupType = "String"
		}
		// Also normalize array/map types
		if strings.HasPrefix(recvType, "[]") || strings.HasPrefix(recvType, "Array") {
			lookupType = "Array"
		}
		if strings.HasPrefix(recvType, "map[") || strings.HasPrefix(recvType, "Map") {
			lookupType = "Map"
		}

		var methodDef MethodDef
		found := false

		if ident, ok := fn.X.(*ast.Ident); ok && ident.Name == "Math" {
			if methods, ok := stdLib["Math"]; ok {
				if def, ok := methods[fn.Sel]; ok {
					g.needsRuntime = true
					g.buf.WriteString(def.RuntimeFunc)
					g.buf.WriteString("(")
					for i, arg := range call.Args {
						if i > 0 {
							g.buf.WriteString(", ")
						}
						g.genExpr(arg)
					}
					g.buf.WriteString(")")
					return
				}
			}
		}

		if lookupType != "" {
			// Optimize type conversions to use direct Go casts
			// Int.to_f -> float64(x), Float.to_i -> int(x)
			// Skip float literals for to_i as Go doesn't allow int(3.5)
			if lookupType == "Int" && fn.Sel == "to_f" {
				g.buf.WriteString("float64(")
				g.genExpr(fn.X)
				g.buf.WriteString(")")
				return
			}
			if lookupType == "Float" && fn.Sel == "to_i" {
				if _, isLit := fn.X.(*ast.FloatLit); !isLit {
					g.buf.WriteString("int(")
					g.genExpr(fn.X)
					g.buf.WriteString(")")
					return
				}
			}

			// Optimize String.substring(start, end) to direct slice syntax s[start:end]
			if lookupType == "String" && fn.Sel == "substring" && len(call.Args) == 2 {
				g.genExpr(fn.X)
				g.buf.WriteString("[")
				g.genExpr(call.Args[0])
				g.buf.WriteString(":")
				g.genExpr(call.Args[1])
				g.buf.WriteString("]")
				return
			}

			if methods, ok := stdLib[lookupType]; ok {
				if def, ok := methods[fn.Sel]; ok {
					methodDef = def
					found = true
				}
			}
		}

		// Only check uniqueMethods if this isn't a Go interop call
		// Otherwise strings.split() would incorrectly match the String.split method
		// Also skip for user-defined class receivers (class methods take precedence)
		if !found && !g.isGoInterop(fn.X) {
			receiverClassName := g.getReceiverClassName(fn.X)
			baseClassName := stripTypeParams(receiverClassName)
			isClassReceiver := baseClassName != "" && (g.isClass(baseClassName) || g.isModuleScopedClass(baseClassName))

			if !isClassReceiver {
				if def, ok := uniqueMethods[fn.Sel]; ok {
					methodDef = def
					found = true
				}
			}
		}

		if found {
			g.needsRuntime = true
			g.buf.WriteString(methodDef.RuntimeFunc)
			g.buf.WriteString("(")
			g.genExpr(fn.X)
			if len(call.Args) > 0 {
				g.buf.WriteString(", ")
				for i, arg := range call.Args {
					if i > 0 {
						g.buf.WriteString(", ")
					}
					g.genExpr(arg)
				}
			}
			g.buf.WriteString(")")
			return
		}

		if fn.Sel == "is_a?" {
			if len(call.Args) != 1 {
				g.addError(fmt.Errorf("is_a? requires exactly one type argument, got %d", len(call.Args)))
				g.buf.WriteString("false")
				return
			}
			// Get the target type name to check if it's a class
			var targetType string
			if ident, ok := call.Args[0].(*ast.Ident); ok {
				targetType = ident.Name
			}
			// Check if target is a class (registered in instanceMethods but not a struct)
			// Classes are reference types and need pointer type assertion
			isClass := g.instanceMethods[targetType] != nil && g.structs[targetType] == nil
			// Cast to any first to allow type assertion on concrete types
			g.buf.WriteString("func() bool { _, ok := any(")
			g.genExpr(fn.X)
			g.buf.WriteString(").(")
			if isClass {
				g.buf.WriteString("*")
			}
			g.genExpr(call.Args[0])
			g.buf.WriteString("); return ok }()")
			return
		}

		if fn.Sel == "as" {
			if len(call.Args) != 1 {
				g.addError(fmt.Errorf("as requires exactly one type argument, got %d", len(call.Args)))
				g.buf.WriteString("func() (any, bool) { return nil, false }()")
				return
			}
			// Get the target type name
			var targetType string
			if ident, ok := call.Args[0].(*ast.Ident); ok {
				targetType = ident.Name
			}
			// Check if target is a class (registered in instanceMethods but not a struct)
			// Classes are reference types and need pointer type assertion
			isClass := g.instanceMethods[targetType] != nil && g.structs[targetType] == nil
			// Cast to any first to allow type assertion on concrete types
			g.buf.WriteString("func() (")
			if isClass {
				g.buf.WriteString("*")
			}
			g.genExpr(call.Args[0])
			g.buf.WriteString(", bool) { v, ok := any(")
			g.genExpr(fn.X)
			g.buf.WriteString(").(")
			if isClass {
				g.buf.WriteString("*")
			}
			g.genExpr(call.Args[0])
			g.buf.WriteString("); return v, ok }()")
			return
		}

		receiverType := g.inferTypeFromExpr(fn.X)
		if isOptionalType(receiverType) {
			switch fn.Sel {
			case "ok?", "present?":
				g.buf.WriteString("(")
				g.genExpr(fn.X)
				g.buf.WriteString(" != nil)")
				return
			case "nil?", "absent?":
				g.buf.WriteString("(")
				g.genExpr(fn.X)
				g.buf.WriteString(" == nil)")
				return
			case "unwrap":
				// For class types, the optional T? is the same as *T (pointer),
				// so unwrap should just return the pointer (no dereference needed).
				// For value types, unwrap dereferences the pointer to get the value.
				baseType := strings.TrimSuffix(receiverType, "?")
				if g.typeInfo.IsClass(baseType) {
					// Class optional - no dereference, pointer IS the value
					g.genExpr(fn.X)
				} else {
					// Value type optional - dereference to get value
					g.buf.WriteString("*")
					g.genExpr(fn.X)
				}
				return
			case "unwrap_or":
				if len(call.Args) == 0 {
					g.buf.WriteString("/* unwrap_or requires default argument */")
					return
				}
				g.needsRuntime = true
				// Check if the receiver is a safe navigation expression (returns any)
				_, isSafeNav := fn.X.(*ast.SafeNavExpr)
				// Generate coalescing code based on the optional type
				switch receiverType {
				case "Int?", "Int64?", "Float?", "Bool?":
					g.buf.WriteString("runtime.Coalesce(")
				case "String?":
					// Use CoalesceStringAny for safe navigation (which returns any)
					if isSafeNav {
						g.buf.WriteString("runtime.CoalesceStringAny(")
					} else {
						g.buf.WriteString("runtime.Coalesce(")
					}
				default:
					// Generic optional type - use IIFE
					varName := fmt.Sprintf("_uo%d", g.tempVarCounter)
					g.tempVarCounter++
					g.buf.WriteString("func() ")
					defaultType := g.inferTypeFromExpr(call.Args[0])
					if defaultType != "" {
						g.buf.WriteString(mapType(defaultType))
					} else {
						g.buf.WriteString("any")
					}
					g.buf.WriteString(" { ")
					g.buf.WriteString(varName)
					g.buf.WriteString(" := ")
					g.genExpr(fn.X)
					g.buf.WriteString("; if ")
					g.buf.WriteString(varName)
					g.buf.WriteString(" != nil { return *")
					g.buf.WriteString(varName)
					g.buf.WriteString(" }; return ")
					g.genExpr(call.Args[0])
					g.buf.WriteString(" }()")
					return
				}
				g.genExpr(fn.X)
				g.buf.WriteString(", ")
				g.genExpr(call.Args[0])
				g.buf.WriteString(")")
				return
			}
		}

		// Handle enum instance method calls (Color::Red.to_s, enum_var.value)
		if g.enums != nil {
			if _, isEnum := g.enums[receiverType]; isEnum {
				switch fn.Sel {
				case "to_s":
					// enum.to_s -> enum.String()
					g.genExpr(fn.X)
					g.buf.WriteString(".String()")
					return
				case "value":
					// enum.value -> int(enum)
					g.buf.WriteString("int(")
					g.genExpr(fn.X)
					g.buf.WriteString(")")
					return
				}
			}
		}

		// Handle range method calls - check both literal ranges and variables of type Range
		isRangeExpr := false
		if _, ok := fn.X.(*ast.RangeLit); ok {
			isRangeExpr = true
		} else if recvType := g.inferTypeFromExpr(fn.X); recvType == "Range" {
			isRangeExpr = true
		}
		if isRangeExpr {
			if g.genRangeMethodCall(fn.X, fn.Sel, call.Args) {
				return
			}
		}
		g.genSelectorExpr(fn)
	default:
		g.genExpr(call.Func)
	}

	// Check if this is a Go interop method call - lambda args need 'any' return type
	isGoInteropMethodCall := false
	if fn, ok := call.Func.(*ast.SelectorExpr); ok {
		isGoInteropMethodCall = g.isGoInterop(fn.X)
	}

	g.buf.WriteString("(")
	savedForceAny := g.forceAnyLambdaReturn
	if isGoInteropMethodCall {
		g.forceAnyLambdaReturn = true
	}
	for i, arg := range call.Args {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.genExpr(arg)
	}
	g.forceAnyLambdaReturn = savedForceAny
	g.buf.WriteString(")")
}

// hasKeywordArgs checks if any of the arguments are keyword arguments.
func (g *Generator) hasKeywordArgs(args []ast.Expression) bool {
	for _, arg := range args {
		if _, ok := arg.(*ast.KeywordArg); ok {
			return true
		}
	}
	return false
}

// genKeywordCallArgs generates function call arguments with keyword args reordered.
// It matches keyword arguments to parameter names and fills in defaults for missing args.
func (g *Generator) genKeywordCallArgs(fnDecl *ast.FuncDecl, args []ast.Expression) {
	// Separate positional and keyword arguments
	var positionalArgs []ast.Expression
	keywordArgs := make(map[string]ast.Expression)
	for _, arg := range args {
		if kwArg, ok := arg.(*ast.KeywordArg); ok {
			keywordArgs[kwArg.Name] = kwArg.Value
		} else {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	// Generate arguments in parameter order
	g.buf.WriteString("(")
	for i, param := range fnDecl.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}

		// Check if we have a positional arg for this position
		if i < len(positionalArgs) {
			g.genExpr(positionalArgs[i])
			continue
		}

		// Check if we have a keyword arg for this parameter
		if kwValue, ok := keywordArgs[param.Name]; ok {
			g.genExpr(kwValue)
			continue
		}

		// Use default value if available
		if param.DefaultValue != nil {
			g.genExpr(param.DefaultValue)
			continue
		}

		// No value provided - this should be caught by semantic analysis
		g.buf.WriteString("nil")
	}
	g.buf.WriteString(")")
}

func (g *Generator) genSelectorExpr(sel *ast.SelectorExpr) {
	// Note: Type parameter methods (T.zero) are handled early in genExpr's SelectorExpr case
	// before the selectorKind dispatching, so they never reach this function.

	// Handle special selector expressions: error interface, channel operations, etc.
	switch sel.Sel {
	case "error":
		// .error on error types -> .Error() (Go's error interface)
		g.genExpr(sel.X)
		g.buf.WriteString(".Error()")
		return
	case "receive":
		// ch.receive -> <-ch
		g.buf.WriteString("<-")
		g.genExpr(sel.X)
		return
	case "try_receive":
		// ch.try_receive -> runtime.TryReceivePtr(ch) (returns *T for optional semantics)
		g.needsRuntime = true
		g.buf.WriteString("runtime.TryReceivePtr(")
		g.genExpr(sel.X)
		g.buf.WriteString(")")
		return
	case "close":
		// ch.close -> close(ch) for channels
		// file.close -> file.Close() for Go interop types
		recvType := g.inferTypeFromExpr(sel.X)
		if strings.HasPrefix(recvType, "Chan<") || strings.HasPrefix(recvType, "chan ") {
			g.buf.WriteString("close(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return
		}
		// Fall through to handle as Go interop method .Close()
	case "await":
		// task.await -> runtime.Await(task)
		g.needsRuntime = true
		g.buf.WriteString("runtime.Await(")
		g.genExpr(sel.X)
		g.buf.WriteString(")")
		return
	case "new":
		// Chan<T>.new -> make(chan T)
		if indexExpr, ok := sel.X.(*ast.IndexExpr); ok {
			if chanIdent, ok := indexExpr.Left.(*ast.Ident); ok && chanIdent.Name == "Chan" {
				var chanType string
				if typeIdent, ok := indexExpr.Index.(*ast.Ident); ok {
					chanType = typeIdent.Name
				} else {
					chanType = "any"
				}
				// Use mapParamType to properly convert class types to pointers
				goType := g.mapParamType(chanType)
				g.buf.WriteString("make(chan ")
				g.buf.WriteString(goType)
				g.buf.WriteString(")")
				return
			}
		}
		// Handle Go package type .new calls like sync.WaitGroup.new
		// Generate zero-value initialization: &pkg.Type{}
		if pkgSel, ok := sel.X.(*ast.SelectorExpr); ok {
			if pkgIdent, ok := pkgSel.X.(*ast.Ident); ok && g.imports[pkgIdent.Name] {
				g.buf.WriteString("&")
				g.buf.WriteString(pkgIdent.Name)
				g.buf.WriteString(".")
				g.buf.WriteString(pkgSel.Sel)
				g.buf.WriteString("{}")
				return
			}
		}
	}

	// Handle class method calls (ClassName.method_name) - implicit call without parens
	if ident, ok := sel.X.(*ast.Ident); ok && g.isClass(ident.Name) {
		if methodMap, hasClass := g.classMethods[ident.Name]; hasClass {
			if funcName, hasMethod := methodMap[sel.Sel]; hasMethod {
				// Generate function call: ClassName_methodName()
				g.buf.WriteString(funcName)
				g.buf.WriteString("()")
				return
			}
		}
	}

	// Check for range methods used without parentheses
	// Handle both literal ranges and variables of type Range
	isRangeExpr := false
	if _, ok := sel.X.(*ast.RangeLit); ok {
		isRangeExpr = true
	} else if recvType := g.inferTypeFromExpr(sel.X); recvType == "Range" {
		isRangeExpr = true
	}
	if isRangeExpr {
		switch sel.Sel {
		case "to_a":
			g.needsRuntime = true
			g.buf.WriteString("runtime.RangeToArray(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return
		case "size", "length":
			g.needsRuntime = true
			g.buf.WriteString("runtime.RangeSize(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return
		}
	}

	// Handle enum methods (.value, .to_s, .values)
	receiverType := g.inferTypeFromExpr(sel.X)
	if g.enums != nil {
		if _, isEnum := g.enums[receiverType]; isEnum {
			switch sel.Sel {
			case "value":
				// enum.value -> int(enum)
				g.buf.WriteString("int(")
				g.genExpr(sel.X)
				g.buf.WriteString(")")
				return
			case "to_s":
				// enum.to_s -> enum.String()
				g.genExpr(sel.X)
				g.buf.WriteString(".String()")
				return
			}
		}
		// Handle enum type methods (Color.values)
		if ident, ok := sel.X.(*ast.Ident); ok {
			if _, isEnum := g.enums[ident.Name]; isEnum {
				switch sel.Sel {
				case "values":
					// Color.values -> ColorValues()
					g.buf.WriteString(ident.Name)
					g.buf.WriteString("Values()")
					return
				}
			}
		}
	}

	// Handle struct field access - convert snake_case to PascalCase
	if g.structs != nil {
		if _, isStruct := g.structs[receiverType]; isStruct {
			g.genExpr(sel.X)
			g.buf.WriteString(".")
			g.buf.WriteString(snakeToPascalWithAcronyms(sel.Sel))
			return
		}
	}

	// Handle .length and .size -> len()
	if sel.Sel == "length" || sel.Sel == "size" {
		g.buf.WriteString("len(")
		g.genExpr(sel.X)
		g.buf.WriteString(")")
		return
	}

	// Check for methods on optional types (ok?, nil?, present?, absent?, unwrap)
	if isOptionalType(receiverType) {
		switch sel.Sel {
		case "ok?", "present?":
			g.buf.WriteString("(")
			g.genExpr(sel.X)
			g.buf.WriteString(" != nil)")
			return
		case "nil?", "absent?":
			g.buf.WriteString("(")
			g.genExpr(sel.X)
			g.buf.WriteString(" == nil)")
			return
		case "unwrap":
			// For class types, the optional T? is the same as *T (pointer),
			// so unwrap should just return the pointer (no dereference needed).
			baseType := strings.TrimSuffix(receiverType, "?")
			if g.typeInfo.IsClass(baseType) {
				// Class optional - no dereference
				g.genExpr(sel.X)
			} else {
				// Value type optional - dereference
				g.buf.WriteString("*")
				g.genExpr(sel.X)
			}
			return
		}
	}

	// Fallback for OptionalResult methods (from optional.map operations)
	// These use PascalCase methods: Ok(), Unwrap()
	// Note: ok? is typically wrapped in CallExpr (predicate methods), so don't add ()
	// But unwrap is used as a property and needs () to call the method
	if sel.Sel == "ok?" {
		g.genExpr(sel.X)
		g.buf.WriteString(".Ok")
		return
	}
	if sel.Sel == "unwrap" {
		g.genExpr(sel.X)
		g.buf.WriteString(".Unwrap()")
		return
	}

	// Handle property-style array/map method calls
	if g.genArrayPropertyMethod(sel) {
		return
	}

	// Handle property-style stdLib method calls (Int.even?, Float.floor, etc.)
	if g.genStdLibPropertyMethod(sel) {
		return
	}

	if g.isGoInterop(sel.X) {
		g.genExpr(sel.X)
		g.buf.WriteString(".")
		// Use actual Go name from semantic analysis if available
		if goName := g.typeInfo.GetGoName(sel); goName != "" {
			g.buf.WriteString(goName)
		} else {
			// Fallback to acronym conversion for unresolved cases
			g.buf.WriteString(snakeToPascalWithAcronyms(sel.Sel, true))
		}
		return
	}

	// Check if this method name matches any interface method
	// If so, use PascalCase for Go interface compatibility
	isInterfaceMethod := g.isInterfaceMethod(sel.Sel)

	// Check if receiver is an instance of a pub class
	// Pub class methods use PascalCase
	receiverClassName := g.getReceiverClassName(sel.X)
	isPubClass := g.isPublicClass(receiverClassName)

	// Check if this specific accessor is pub (for non-pub classes with pub accessors)
	isPubAccessor := false
	if classAccessors := g.pubAccessors[receiverClassName]; classAccessors != nil {
		isPubAccessor = classAccessors[sel.Sel]
	}

	// Check if this specific method is pub (for non-pub classes with pub methods)
	isPubMethod := false
	if classMethods := g.pubMethods[receiverClassName]; classMethods != nil {
		isPubMethod = classMethods[sel.Sel]
	}

	// Check if this is a private method (for non-pub classes with private methods)
	isPrivateMethod := false
	if classMethods := g.privateMethods[receiverClassName]; classMethods != nil {
		isPrivateMethod = classMethods[sel.Sel]
	}

	g.genExpr(sel.X)
	g.buf.WriteString(".")

	// Special case: to_s maps to String (Go's Stringer interface convention)
	if sel.Sel == "to_s" {
		g.buf.WriteString("String")
		return
	}

	if isInterfaceMethod || isPubClass || isPubAccessor || isPubMethod {
		g.buf.WriteString(snakeToPascalWithAcronyms(sel.Sel))
	} else if isPrivateMethod {
		g.buf.WriteString("_" + snakeToCamelWithAcronyms(sel.Sel))
	} else {
		g.buf.WriteString(snakeToCamelWithAcronyms(sel.Sel))
	}
}

func (g *Generator) isGoInterop(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		// Check both imports and variables holding Go interop types
		return g.imports[e.Name] || g.goInteropVars[e.Name]
	case *ast.SelectorExpr:
		return g.isGoInterop(e.X)
	case *ast.InstanceVar:
		// Check if instance variable has a Go interop type (like *big.Int)
		fieldType := g.getFieldType(e.Name)
		return g.isGoInteropType(fieldType)
	default:
		return false
	}
}

// isGoInteropType checks if a type annotation refers to a Go interop type.
// Examples: "*bufio.Writer", "io.Reader", "*http.Request"
// This is used to track parameters with Go types for PascalCase method conversion.
func (g *Generator) isGoInteropType(typeName string) bool {
	// Strip pointer prefix
	typeName = strings.TrimPrefix(typeName, "*")

	// Check if type contains a dot (package.Type pattern)
	if strings.Contains(typeName, ".") {
		// Extract package name and check if it's an imported Go package
		parts := strings.Split(typeName, ".")
		if len(parts) >= 2 {
			pkgName := parts[0]
			return g.imports[pkgName]
		}
	}
	return false
}

// isGoInteropTypeConstructor checks if an expression is a Go type constructor like sync.WaitGroup.new
// or a Go struct literal like sync.WaitGroup{}.
// Returns true if the expression creates an instance of a Go type.
func (g *Generator) isGoInteropTypeConstructor(expr ast.Expression) bool {
	// Check for GoStructLit (sync.WaitGroup{})
	if goStruct, ok := expr.(*ast.GoStructLit); ok {
		return g.imports[goStruct.Package]
	}

	// First check for CallExpr (sync.WaitGroup.new())
	if call, ok := expr.(*ast.CallExpr); ok {
		sel, ok := call.Func.(*ast.SelectorExpr)
		if !ok || sel.Sel != "new" {
			return false
		}
		expr = sel // Continue checking the selector
	}

	// Check for SelectorExpr (sync.WaitGroup.new - property style)
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel != "new" {
		return false
	}
	// Check if it's pkg.Type.new pattern
	innerSel, ok := sel.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	// Check if the package is a Go import
	pkgIdent, ok := innerSel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return g.imports[pkgIdent.Name]
}

// isGoInteropCall checks if an expression is a function call from a Go package.
// Examples: http.get(url), json.parse(data), md5.new, etc.
// Also handles BangExpr (!) wrapping: http.get(url)!
// Also handles SelectorExpr with SelectorGoMethod kind (no-arg Go function calls).
func (g *Generator) isGoInteropCall(expr ast.Expression) bool {
	// Unwrap BangExpr (!)
	if bang, ok := expr.(*ast.BangExpr); ok {
		expr = bang.Expr
	}

	// Check for SelectorExpr with SelectorGoMethod kind (no-arg Go function calls like md5.new)
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		selectorKind := g.getSelectorKind(sel)
		if selectorKind == ast.SelectorGoMethod && g.isGoInterop(sel.X) {
			return true
		}
	}

	// Check for CallExpr
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	// Check if function is pkg.func pattern
	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Check if the receiver is a Go import
	return g.isGoInterop(sel.X)
}

// isInterfaceMethod checks if a method name matches any interface method in the program.
// This is used to determine whether to use PascalCase (for interface methods) or camelCase.
func (g *Generator) isInterfaceMethod(methodName string) bool {
	// Check explicit interfaces
	for _, ifaceName := range g.getAllInterfaceNames() {
		if slices.Contains(g.getInterfaceMethodNames(ifaceName), methodName) {
			return true
		}
	}
	// Check module-generated interfaces
	for _, modName := range g.getAllModuleNames() {
		if slices.Contains(g.getModuleMethodNames(modName), methodName) {
			return true
		}
	}
	return false
}

// getReceiverClassName extracts the class name from a receiver expression.
// For example, if expr is a variable of type *User, returns "User".
// Returns empty string if the class name cannot be determined.
func (g *Generator) getReceiverClassName(expr ast.Expression) string {
	// Use type info from semantic analysis (if it's a real type, not any/unknown)
	if typ := g.typeInfo.GetRugbyType(expr); typ != "" && typ != "any" && typ != "unknown" {
		// Strip pointer/optional prefixes if present
		typ = strings.TrimPrefix(typ, "*")
		typ = strings.TrimSuffix(typ, "?")
		return typ
	}

	// Fall back to inferring from expression
	switch e := expr.(type) {
	case *ast.Ident:
		// Fall back to codegen's vars map (variable name -> type)
		if varType, ok := g.vars[e.Name]; ok && varType != "" {
			varType = strings.TrimPrefix(varType, "*")
			varType = strings.TrimSuffix(varType, "?")
			return varType
		}
	case *ast.CallExpr:
		// Constructor call like User.new() -> "User"
		if sel, ok := e.Func.(*ast.SelectorExpr); ok && sel.Sel == "new" {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return ident.Name
			}
		}
	}
	return ""
}

// isClassInstanceMethod checks if the given method name is an instance method
// (including getters) on the given class or any of its parent classes.
// Returns the Go method name if found, empty string otherwise.
func (g *Generator) isClassInstanceMethod(className, methodName string) string {
	// Walk up the inheritance chain
	for className != "" {
		if methods := g.instanceMethods[className]; methods != nil {
			if goName, ok := methods[methodName]; ok {
				return goName
			}
		}
		// Move to parent class
		className = g.classParent[className]
	}
	return ""
}

// genArrayPropertyMethod handles property-style array and map method calls
// like nums.sum, nums.first, m.keys, etc.
// Returns true if the method was handled, false otherwise.
func (g *Generator) genArrayPropertyMethod(sel *ast.SelectorExpr) bool {
	receiverType := g.inferTypeFromExpr(sel.X)
	elemType := g.inferArrayElementGoType(sel.X)

	// Handle array property methods
	if strings.HasPrefix(receiverType, "[]") || receiverType == "Array" || strings.HasPrefix(receiverType, "Array<") {
		switch sel.Sel {
		case "sum":
			g.needsRuntime = true
			if elemType == "float64" {
				g.buf.WriteString("runtime.SumFloat(")
			} else {
				g.buf.WriteString("runtime.SumInt(")
			}
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "min":
			g.needsRuntime = true
			// Use Ptr versions that return *T for optional coalescing
			if elemType == "float64" {
				g.buf.WriteString("runtime.MinFloatPtr(")
			} else {
				g.buf.WriteString("runtime.MinIntPtr(")
			}
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "max":
			g.needsRuntime = true
			if elemType == "float64" {
				g.buf.WriteString("runtime.MaxFloatPtr(")
			} else {
				g.buf.WriteString("runtime.MaxIntPtr(")
			}
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "first":
			g.needsRuntime = true
			g.buf.WriteString("runtime.FirstPtr(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "last":
			g.needsRuntime = true
			g.buf.WriteString("runtime.LastPtr(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "sorted":
			g.needsRuntime = true
			g.buf.WriteString("runtime.Sort(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "any?":
			// any? without a block returns true if array has any elements
			g.buf.WriteString("(len(")
			g.genExpr(sel.X)
			g.buf.WriteString(") > 0)")
			return true
		case "empty?":
			// empty? returns true if array has no elements
			g.buf.WriteString("(len(")
			g.genExpr(sel.X)
			g.buf.WriteString(") == 0)")
			return true
		case "shift":
			// shift removes and returns the first element, needs pointer
			g.needsRuntime = true
			switch elemType {
			case "int":
				g.buf.WriteString("runtime.ShiftInt(&")
			case "string":
				g.buf.WriteString("runtime.ShiftString(&")
			case "float64":
				g.buf.WriteString("runtime.ShiftFloat(&")
			default:
				g.buf.WriteString("runtime.ShiftAny(&")
			}
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "pop":
			// pop removes and returns the last element, needs pointer
			g.needsRuntime = true
			switch elemType {
			case "int":
				g.buf.WriteString("runtime.PopInt(&")
			case "string":
				g.buf.WriteString("runtime.PopString(&")
			case "float64":
				g.buf.WriteString("runtime.PopFloat(&")
			default:
				g.buf.WriteString("runtime.PopAny(&")
			}
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		}
	}

	// Handle map property methods
	if strings.HasPrefix(receiverType, "map[") || receiverType == "Map" {
		g.needsRuntime = true
		switch sel.Sel {
		case "keys":
			g.buf.WriteString("runtime.Keys(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		case "values":
			g.buf.WriteString("runtime.Values(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		}
	}

	return false
}

// genStdLibPropertyMethod handles property-style method calls on Int, Float, String types
// like x.even?, x.abs, f.floor, s.upcase, etc.
// Returns true if the method was handled, false otherwise.
func (g *Generator) genStdLibPropertyMethod(sel *ast.SelectorExpr) bool {
	// Don't intercept Go interop calls - they should use native Go methods
	if g.isGoInterop(sel.X) {
		return false
	}

	// Skip stdLib methods for user-defined class receivers (class methods take precedence)
	receiverClassName := g.getReceiverClassName(sel.X)
	if receiverClassName != "" {
		baseClassName := stripTypeParams(receiverClassName)
		if g.isClass(baseClassName) || g.isModuleScopedClass(baseClassName) {
			return false
		}
	}

	receiverType := g.inferTypeFromExpr(sel.X)

	// Map common Go types to Rugby types for lookup
	lookupType := receiverType
	switch receiverType {
	case "int", "int64":
		lookupType = "Int"
	case "float64":
		lookupType = "Float"
	case "string":
		lookupType = "String"
	}

	// Optimize type conversions to use direct Go casts instead of runtime calls
	// Int.to_f -> float64(x), Float.to_i -> int(x)
	// Skip float literals for to_i as Go doesn't allow int(3.5)
	if lookupType == "Int" && sel.Sel == "to_f" {
		g.buf.WriteString("float64(")
		g.genExpr(sel.X)
		g.buf.WriteString(")")
		return true
	}
	if lookupType == "Float" && sel.Sel == "to_i" {
		if _, isLit := sel.X.(*ast.FloatLit); !isLit {
			g.buf.WriteString("int(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		}
	}

	if methods, ok := stdLib[lookupType]; ok {
		if def, ok := methods[sel.Sel]; ok {
			g.needsRuntime = true
			g.buf.WriteString(def.RuntimeFunc)
			g.buf.WriteString("(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		}
	}

	// Also check uniqueMethods for methods with unique names
	// But skip if the receiver is a user-defined class (class methods take precedence)
	baseType := stripTypeParams(receiverType)
	if !g.isClass(baseType) && !g.isModuleScopedClass(baseType) {
		if def, ok := uniqueMethods[sel.Sel]; ok {
			g.needsRuntime = true
			g.buf.WriteString(def.RuntimeFunc)
			g.buf.WriteString("(")
			g.genExpr(sel.X)
			g.buf.WriteString(")")
			return true
		}
	}

	return false
}
