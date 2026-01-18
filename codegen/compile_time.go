package codegen

import "github.com/nchapman/rugby/ast"

// CompileTimeHandler handles compile-time code generation for specific function calls.
// Packages that want compile-time processing implement this interface.
type CompileTimeHandler interface {
	// Generate produces Go code for a compile-time function call.
	// Parameters:
	//   - g: the code generator
	//   - name: the variable/const name being assigned
	//   - call: the call expression AST node
	//   - method: the specific method being called (e.g., "compile", "compile_file")
	Generate(g *Generator, name string, call *ast.CallExpr, method string)
}

// compileTimeRegistry maps "package.method" to handlers.
// Currently hardcoded; could be made dynamic with a Register() function later.
var compileTimeRegistry = map[string]CompileTimeHandler{
	"template.compile":      &templateCompileHandler{},
	"template.compile_file": &templateCompileHandler{},
	"template.compile_dir":  &templateCompileHandler{},
	"template.compile_glob": &templateCompileHandler{},
}

// getCompileTimeHandler checks if a call expression matches a registered compile-time handler.
// Returns the handler, method name, and true if found; nil, "", false otherwise.
func (g *Generator) getCompileTimeHandler(call *ast.CallExpr) (CompileTimeHandler, string, bool) {
	sel, ok := call.Func.(*ast.SelectorExpr)
	if !ok {
		return nil, "", false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return nil, "", false
	}

	key := ident.Name + "." + sel.Sel
	if handler, ok := compileTimeRegistry[key]; ok {
		return handler, sel.Sel, true
	}

	return nil, "", false
}
