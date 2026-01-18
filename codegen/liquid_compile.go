package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/stdlib/liquid"
)

// liquidCompileHandler handles compile-time processing for liquid.compile() and liquid.compile_file().
type liquidCompileHandler struct{}

// Generate implements CompileTimeHandler for Liquid templates.
func (h *liquidCompileHandler) Generate(g *Generator, name string, call *ast.CallExpr, method string) {
	// Require at least one argument
	if len(call.Args) == 0 {
		g.addError(fmt.Errorf("liquid.%s requires a string argument", method))
		return
	}

	switch method {
	case "compile":
		h.generateSingleTemplate(g, name, call)
	case "compile_file":
		h.generateSingleTemplateFromFile(g, name, call)
	case "compile_dir":
		h.generateTemplateMap(g, name, call, false)
	case "compile_glob":
		h.generateTemplateMap(g, name, call, true)
	default:
		g.addError(fmt.Errorf("unknown liquid compile method: %s", method))
	}
}

// generateSingleTemplate handles liquid.compile("template string").
func (h *liquidCompileHandler) generateSingleTemplate(g *Generator, name string, call *ast.CallExpr) {
	strLit, ok := call.Args[0].(*ast.StringLit)
	if !ok {
		g.addError(fmt.Errorf("liquid.compile requires a string literal argument"))
		return
	}

	tmpl, err := liquid.Parse(strLit.Value)
	if err != nil {
		g.addError(fmt.Errorf("liquid template syntax error: %v", err))
		return
	}

	g.genLiquidCompiledTemplate(name, tmpl)
}

// generateSingleTemplateFromFile handles liquid.compile_file("path").
func (h *liquidCompileHandler) generateSingleTemplateFromFile(g *Generator, name string, call *ast.CallExpr) {
	strLit, ok := call.Args[0].(*ast.StringLit)
	if !ok {
		g.addError(fmt.Errorf("liquid.compile_file requires a string literal file path"))
		return
	}

	filePath := strLit.Value
	if !filepath.IsAbs(filePath) && g.sourceDir != "" {
		filePath = filepath.Join(g.sourceDir, filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		g.addError(fmt.Errorf("liquid.compile_file: cannot read file %q: %v", strLit.Value, err))
		return
	}

	tmpl, err := liquid.Parse(string(content))
	if err != nil {
		g.addError(fmt.Errorf("liquid template syntax error in %q: %v", strLit.Value, err))
		return
	}

	g.genLiquidCompiledTemplate(name, tmpl)
}

// generateTemplateMap handles liquid.compile_dir("path") and liquid.compile_glob("pattern").
// compile_dir recursively walks the directory and includes only .liquid files.
// compile_glob matches files according to the glob pattern with no extension filtering.
func (h *liquidCompileHandler) generateTemplateMap(g *Generator, name string, call *ast.CallExpr, isGlob bool) {
	strLit, ok := call.Args[0].(*ast.StringLit)
	if !ok {
		methodName := "compile_dir"
		argType := "directory path"
		if isGlob {
			methodName = "compile_glob"
			argType = "glob pattern"
		}
		g.addError(fmt.Errorf("liquid.%s requires a string literal %s", methodName, argType))
		return
	}

	pattern := strLit.Value
	basePath := pattern

	// Resolve relative path
	if !filepath.IsAbs(basePath) && g.sourceDir != "" {
		basePath = filepath.Join(g.sourceDir, pattern)
	}

	// Collect templates
	templates := make(map[string]*liquid.Template)

	if isGlob {
		// compile_glob: use filepath.Glob
		matches, err := filepath.Glob(basePath)
		if err != nil {
			g.addError(fmt.Errorf("liquid.compile_glob: invalid pattern %q: %v", pattern, err))
			return
		}
		if len(matches) == 0 {
			g.addError(fmt.Errorf("liquid.compile_glob: no files matched pattern %q", pattern))
			return
		}

		// Determine base directory for relative keys
		baseDir := extractGlobBase(basePath)

		for _, filePath := range matches {
			info, err := os.Stat(filePath)
			if err != nil || info.IsDir() {
				continue
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				g.addError(fmt.Errorf("liquid.compile_glob: cannot read file %q: %v", filePath, err))
				return
			}

			tmpl, err := liquid.Parse(string(content))
			if err != nil {
				relPath, _ := filepath.Rel(baseDir, filePath)
				g.addError(fmt.Errorf("liquid template syntax error in %q: %v", relPath, err))
				return
			}

			// Key is relative path from base directory
			key, _ := filepath.Rel(baseDir, filePath)
			templates[key] = tmpl
		}
	} else {
		// compile_dir: walk directory recursively
		info, err := os.Stat(basePath)
		if err != nil {
			g.addError(fmt.Errorf("liquid.compile_dir: cannot access directory %q: %v", pattern, err))
			return
		}
		if !info.IsDir() {
			g.addError(fmt.Errorf("liquid.compile_dir: %q is not a directory", pattern))
			return
		}

		err = filepath.WalkDir(basePath, func(filePath string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			// Filter for .liquid files
			if !strings.HasSuffix(filePath, ".liquid") {
				return nil
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("cannot read file %q: %v", filePath, err)
			}

			tmpl, err := liquid.Parse(string(content))
			if err != nil {
				relPath, _ := filepath.Rel(basePath, filePath)
				return fmt.Errorf("template syntax error in %q: %v", relPath, err)
			}

			// Key is relative path from base directory
			key, _ := filepath.Rel(basePath, filePath)
			templates[key] = tmpl
			return nil
		})

		if err != nil {
			g.addError(fmt.Errorf("liquid.compile_dir: %v", err))
			return
		}

		if len(templates) == 0 {
			g.addError(fmt.Errorf("liquid.compile_dir: no .liquid files found in %q", pattern))
			return
		}
	}

	g.genLiquidTemplateMap(name, templates)
}

// extractGlobBase finds the base directory before any glob wildcards.
func extractGlobBase(pattern string) string {
	// Find the first directory component without wildcards
	dir := filepath.Dir(pattern)
	for {
		if !strings.ContainsAny(dir, "*?[") {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			return "."
		}
		dir = parent
	}
}

// genLiquidCompiledTemplate generates a single liquid.CompiledTemplate variable.
func (g *Generator) genLiquidCompiledTemplate(name string, tmpl *liquid.Template) {
	g.needsLiquid = true
	g.needsStrings = true

	// Register the type in constTypes so inferTypeFromExpr can find it across function boundaries.
	g.constTypes[name] = "liquid.CompiledTemplate"

	g.writeIndent()
	g.buf.WriteString("var ")
	g.buf.WriteString(name)
	g.buf.WriteString(" = liquid.CompiledTemplate{\n")
	g.indent++
	g.writeIndent()
	g.buf.WriteString("Render: func(data map[string]any) (string, error) {\n")
	g.indent++

	g.genLiquidRenderBody(tmpl)

	g.indent--
	g.writeIndent()
	g.buf.WriteString("},\n")
	g.indent--
	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genLiquidTemplateMap generates a map[string]*liquid.CompiledTemplate variable.
// Uses pointers so that MustRender (pointer receiver) can be called on map elements.
func (g *Generator) genLiquidTemplateMap(name string, templates map[string]*liquid.Template) {
	g.needsLiquid = true
	g.needsStrings = true

	// Register the type in constTypes so inferTypeFromExpr can find it across function boundaries.
	// This is necessary because g.vars is cleared for each function, but compile-time generated
	// constants are module-level and need their types available throughout the entire module.
	g.constTypes[name] = "map[string]*liquid.CompiledTemplate"

	g.writeIndent()
	g.buf.WriteString("var ")
	g.buf.WriteString(name)
	g.buf.WriteString(" = map[string]*liquid.CompiledTemplate{\n")
	g.indent++

	// Sort keys for deterministic output
	keys := make([]string, 0, len(templates))
	for k := range templates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		tmpl := templates[key]
		g.writeIndent()
		g.buf.WriteString(strconv.Quote(key))
		g.buf.WriteString(": {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString("Render: func(data map[string]any) (string, error) {\n")
		g.indent++

		g.genLiquidRenderBody(tmpl)

		g.indent--
		g.writeIndent()
		g.buf.WriteString("},\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("},\n")
	}

	g.indent--
	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genLiquidRenderBody generates the body of a Render function for a parsed template.
func (g *Generator) genLiquidRenderBody(tmpl *liquid.Template) {
	g.writeIndent()
	g.buf.WriteString("var buf strings.Builder\n")
	g.writeIndent()
	g.buf.WriteString("ctx := liquid.NewContext(data)\n")

	for _, node := range tmpl.AST() {
		g.genLiquidNode(node)
	}

	g.writeIndent()
	g.buf.WriteString("_ = ctx\n")
	g.writeIndent()
	g.buf.WriteString("return buf.String(), nil\n")
}

// genLiquidNode generates Go code for a single Liquid AST node.
func (g *Generator) genLiquidNode(node liquid.Node) {
	switch n := node.(type) {
	case *liquid.TextNode:
		g.genLiquidTextNode(n)
	case *liquid.OutputNode:
		g.genLiquidOutputNode(n)
	case *liquid.IfTag:
		g.genLiquidIfTag(n)
	case *liquid.UnlessTag:
		g.genLiquidUnlessTag(n)
	case *liquid.ForTag:
		g.genLiquidForTag(n)
	case *liquid.CaseTag:
		g.genLiquidCaseTag(n)
	case *liquid.AssignTag:
		g.genLiquidAssignTag(n)
	case *liquid.CaptureTag:
		g.genLiquidCaptureTag(n)
	case *liquid.BreakTag:
		g.writeIndent()
		g.buf.WriteString("break\n")
	case *liquid.ContinueTag:
		g.writeIndent()
		g.buf.WriteString("continue\n")
	case *liquid.CommentTag:
		// Comments produce no output
	case *liquid.RawTag:
		g.writeIndent()
		g.buf.WriteString("buf.WriteString(")
		g.buf.WriteString(strconv.Quote(n.Content))
		g.buf.WriteString(")\n")
	}
}

// genLiquidTextNode generates code for a TextNode.
func (g *Generator) genLiquidTextNode(n *liquid.TextNode) {
	if n.Text == "" {
		return
	}
	g.writeIndent()
	g.buf.WriteString("buf.WriteString(")
	g.buf.WriteString(strconv.Quote(n.Text))
	g.buf.WriteString(")\n")
}

// genLiquidOutputNode generates code for an OutputNode ({{ expression }}).
func (g *Generator) genLiquidOutputNode(n *liquid.OutputNode) {
	g.writeIndent()
	g.buf.WriteString("buf.WriteString(liquid.ToString(")
	g.genLiquidExpr(n.Expr)
	g.buf.WriteString("))\n")
}

// genLiquidExpr generates Go code for a Liquid expression.
func (g *Generator) genLiquidExpr(expr liquid.Expression) {
	switch e := expr.(type) {
	case *liquid.IdentExpr:
		g.buf.WriteString("ctx.Get(")
		g.buf.WriteString(strconv.Quote(e.Name))
		g.buf.WriteString(")")

	case *liquid.LiteralExpr:
		switch v := e.Value.(type) {
		case string:
			g.buf.WriteString(strconv.Quote(v))
		case int:
			g.buf.WriteString(strconv.Itoa(v))
		case int64:
			g.buf.WriteString(strconv.FormatInt(v, 10))
		case float64:
			g.buf.WriteString(strconv.FormatFloat(v, 'g', -1, 64))
		case bool:
			if v {
				g.buf.WriteString("true")
			} else {
				g.buf.WriteString("false")
			}
		case nil:
			g.buf.WriteString("nil")
		default:
			// For any other type, generate a nil
			g.buf.WriteString("nil")
		}

	case *liquid.DotExpr:
		g.buf.WriteString("liquid.GetProperty(")
		g.genLiquidExpr(e.Object)
		g.buf.WriteString(", ")
		g.buf.WriteString(strconv.Quote(e.Property))
		g.buf.WriteString(")")

	case *liquid.IndexExpr:
		g.buf.WriteString("liquid.GetIndex(")
		g.genLiquidExpr(e.Object)
		g.buf.WriteString(", ")
		g.genLiquidExpr(e.Index)
		g.buf.WriteString(")")

	case *liquid.FilterExpr:
		g.genLiquidFilterExpr(e)

	case *liquid.BinaryExpr:
		g.genLiquidBinaryExpr(e)

	case *liquid.RangeExpr:
		g.buf.WriteString("liquid.MakeRange(")
		g.genLiquidExpr(e.Start)
		g.buf.WriteString(", ")
		g.genLiquidExpr(e.End)
		g.buf.WriteString(")")

	default:
		// Unknown expression type - generate nil
		g.buf.WriteString("nil")
	}
}

// genLiquidFilterExpr generates code for a filter expression.
func (g *Generator) genLiquidFilterExpr(e *liquid.FilterExpr) {
	switch e.Name {
	case "upcase":
		g.buf.WriteString("liquid.FilterUpcase(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(")")
	case "downcase":
		g.buf.WriteString("liquid.FilterDowncase(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(")")
	case "capitalize":
		g.buf.WriteString("liquid.FilterCapitalize(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(")")
	case "strip":
		g.buf.WriteString("liquid.FilterStrip(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(")")
	case "escape":
		g.buf.WriteString("liquid.FilterEscape(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(")")
	case "first":
		g.buf.WriteString("liquid.FilterFirst(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(")")
	case "last":
		g.buf.WriteString("liquid.FilterLast(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(")")
	case "size":
		g.buf.WriteString("liquid.FilterSize(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(")")
	case "reverse":
		g.buf.WriteString("liquid.FilterReverse(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(")")
	case "join":
		g.buf.WriteString("liquid.FilterJoin(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(", ")
		if len(e.Args) > 0 {
			g.buf.WriteString("liquid.ToString(")
			g.genLiquidExpr(e.Args[0])
			g.buf.WriteString(")")
		} else {
			g.buf.WriteString(`" "`)
		}
		g.buf.WriteString(")")
	case "default":
		g.buf.WriteString("liquid.FilterDefault(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(", ")
		if len(e.Args) > 0 {
			g.genLiquidExpr(e.Args[0])
		} else {
			g.buf.WriteString(`""`)
		}
		g.buf.WriteString(")")
	case "plus":
		g.buf.WriteString("liquid.FilterPlus(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(", ")
		if len(e.Args) > 0 {
			g.genLiquidExpr(e.Args[0])
		} else {
			g.buf.WriteString("0")
		}
		g.buf.WriteString(")")
	case "minus":
		g.buf.WriteString("liquid.FilterMinus(")
		g.genLiquidExpr(e.Input)
		g.buf.WriteString(", ")
		if len(e.Args) > 0 {
			g.genLiquidExpr(e.Args[0])
		} else {
			g.buf.WriteString("0")
		}
		g.buf.WriteString(")")
	default:
		// Unknown filter - just pass through input
		g.genLiquidExpr(e.Input)
	}
}

// genLiquidBinaryExpr generates code for a binary expression.
func (g *Generator) genLiquidBinaryExpr(e *liquid.BinaryExpr) {
	switch e.Operator {
	case "and":
		g.buf.WriteString("(liquid.ToBool(")
		g.genLiquidExpr(e.Left)
		g.buf.WriteString(") && liquid.ToBool(")
		g.genLiquidExpr(e.Right)
		g.buf.WriteString("))")
	case "or":
		g.buf.WriteString("(liquid.ToBool(")
		g.genLiquidExpr(e.Left)
		g.buf.WriteString(") || liquid.ToBool(")
		g.genLiquidExpr(e.Right)
		g.buf.WriteString("))")
	default:
		// Use CompareValues for comparison operators
		g.buf.WriteString("liquid.CompareValues(")
		g.genLiquidExpr(e.Left)
		g.buf.WriteString(", ")
		g.buf.WriteString(strconv.Quote(e.Operator))
		g.buf.WriteString(", ")
		g.genLiquidExpr(e.Right)
		g.buf.WriteString(")")
	}
}

// genLiquidIfTag generates code for an if tag.
func (g *Generator) genLiquidIfTag(tag *liquid.IfTag) {
	g.writeIndent()
	g.buf.WriteString("if liquid.ToBool(")
	g.genLiquidExpr(tag.Condition)
	g.buf.WriteString(") {\n")
	g.indent++
	for _, node := range tag.ThenBranch {
		g.genLiquidNode(node)
	}
	g.indent--

	// Handle elsif branches
	for _, elsif := range tag.ElsifBranches {
		g.writeIndent()
		g.buf.WriteString("} else if liquid.ToBool(")
		g.genLiquidExpr(elsif.Condition)
		g.buf.WriteString(") {\n")
		g.indent++
		for _, node := range elsif.Body {
			g.genLiquidNode(node)
		}
		g.indent--
	}

	// Handle else branch
	if len(tag.ElseBranch) > 0 {
		g.writeIndent()
		g.buf.WriteString("} else {\n")
		g.indent++
		for _, node := range tag.ElseBranch {
			g.genLiquidNode(node)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genLiquidUnlessTag generates code for an unless tag.
func (g *Generator) genLiquidUnlessTag(tag *liquid.UnlessTag) {
	g.writeIndent()
	g.buf.WriteString("if !liquid.ToBool(")
	g.genLiquidExpr(tag.Condition)
	g.buf.WriteString(") {\n")
	g.indent++
	for _, node := range tag.Body {
		g.genLiquidNode(node)
	}
	g.indent--

	// Handle else branch
	if len(tag.ElseBranch) > 0 {
		g.writeIndent()
		g.buf.WriteString("} else {\n")
		g.indent++
		for _, node := range tag.ElseBranch {
			g.genLiquidNode(node)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genLiquidForTag generates code for a for loop.
func (g *Generator) genLiquidForTag(tag *liquid.ForTag) {
	// Generate a unique variable name for the collection
	collectionVar := fmt.Sprintf("_lfc%d", g.tempVarCounter)
	g.tempVarCounter++

	// Get the collection
	g.writeIndent()
	g.buf.WriteString(collectionVar)
	g.buf.WriteString(" := liquid.ToSlice(")
	g.genLiquidExpr(tag.Collection)
	g.buf.WriteString(")\n")

	// Handle offset
	if tag.Offset != nil {
		g.writeIndent()
		g.buf.WriteString("if _off := liquid.ToIntValue(")
		g.genLiquidExpr(tag.Offset)
		g.buf.WriteString("); _off > 0 && _off < len(")
		g.buf.WriteString(collectionVar)
		g.buf.WriteString(") {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString(collectionVar)
		g.buf.WriteString(" = ")
		g.buf.WriteString(collectionVar)
		g.buf.WriteString("[_off:]\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}

	// Handle limit
	if tag.Limit != nil {
		g.writeIndent()
		g.buf.WriteString("if _lim := liquid.ToIntValue(")
		g.genLiquidExpr(tag.Limit)
		g.buf.WriteString("); _lim > 0 && _lim < len(")
		g.buf.WriteString(collectionVar)
		g.buf.WriteString(") {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString(collectionVar)
		g.buf.WriteString(" = ")
		g.buf.WriteString(collectionVar)
		g.buf.WriteString("[:_lim]\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}

	// Handle reversed
	if tag.Reversed {
		reversedVar := fmt.Sprintf("_lfr%d", g.tempVarCounter)
		g.tempVarCounter++
		g.writeIndent()
		g.buf.WriteString(reversedVar)
		g.buf.WriteString(" := make([]any, len(")
		g.buf.WriteString(collectionVar)
		g.buf.WriteString("))\n")
		g.writeIndent()
		g.buf.WriteString("for _i, _v := range ")
		g.buf.WriteString(collectionVar)
		g.buf.WriteString(" {\n")
		g.indent++
		g.writeIndent()
		g.buf.WriteString(reversedVar)
		g.buf.WriteString("[len(")
		g.buf.WriteString(collectionVar)
		g.buf.WriteString(")-1-_i] = _v\n")
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
		g.writeIndent()
		g.buf.WriteString(collectionVar)
		g.buf.WriteString(" = ")
		g.buf.WriteString(reversedVar)
		g.buf.WriteString("\n")
	}

	// Handle empty collection (else block)
	if len(tag.ElseBody) > 0 {
		g.writeIndent()
		g.buf.WriteString("if len(")
		g.buf.WriteString(collectionVar)
		g.buf.WriteString(") == 0 {\n")
		g.indent++
		for _, node := range tag.ElseBody {
			g.genLiquidNode(node)
		}
		g.indent--
		g.writeIndent()
		g.buf.WriteString("} else {\n")
		g.indent++
	}

	// Save outer context and create a new scope for the loop
	outerCtxVar := fmt.Sprintf("_lfctx%d", g.tempVarCounter)
	g.tempVarCounter++
	g.writeIndent()
	g.buf.WriteString(outerCtxVar)
	g.buf.WriteString(" := ctx\n")
	g.writeIndent()
	g.buf.WriteString("ctx = ctx.Push()\n")

	// Generate the for loop
	lengthVar := fmt.Sprintf("_lfl%d", g.tempVarCounter)
	g.tempVarCounter++
	g.writeIndent()
	g.buf.WriteString(lengthVar)
	g.buf.WriteString(" := len(")
	g.buf.WriteString(collectionVar)
	g.buf.WriteString(")\n")

	g.writeIndent()
	g.buf.WriteString("for _i, _item := range ")
	g.buf.WriteString(collectionVar)
	g.buf.WriteString(" {\n")
	g.indent++

	// Set the loop variable
	g.writeIndent()
	g.buf.WriteString("ctx.Set(")
	g.buf.WriteString(strconv.Quote(tag.Variable))
	g.buf.WriteString(", _item)\n")

	// Set forloop metadata
	g.writeIndent()
	g.buf.WriteString("ctx.SetForloop(_i, ")
	g.buf.WriteString(lengthVar)
	g.buf.WriteString(")\n")

	// Generate loop body
	for _, node := range tag.Body {
		g.genLiquidNode(node)
	}

	g.indent--
	g.writeIndent()
	g.buf.WriteString("}\n")

	// Restore outer context (pop the loop scope)
	g.writeIndent()
	g.buf.WriteString("ctx = ")
	g.buf.WriteString(outerCtxVar)
	g.buf.WriteString("\n")

	if len(tag.ElseBody) > 0 {
		g.indent--
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

// genLiquidCaseTag generates code for a case tag.
func (g *Generator) genLiquidCaseTag(tag *liquid.CaseTag) {
	valueVar := fmt.Sprintf("_lcv%d", g.tempVarCounter)
	g.tempVarCounter++

	g.writeIndent()
	g.buf.WriteString(valueVar)
	g.buf.WriteString(" := ")
	g.genLiquidExpr(tag.Value)
	g.buf.WriteString("\n")

	for i, when := range tag.Whens {
		g.writeIndent()
		if i == 0 {
			g.buf.WriteString("if ")
		} else {
			g.buf.WriteString("} else if ")
		}

		// Generate condition for all values in when clause
		for j, val := range when.Values {
			if j > 0 {
				g.buf.WriteString(" || ")
			}
			g.buf.WriteString("liquid.CompareValues(")
			g.buf.WriteString(valueVar)
			g.buf.WriteString(", \"==\", ")
			g.genLiquidExpr(val)
			g.buf.WriteString(")")
		}

		g.buf.WriteString(" {\n")
		g.indent++
		for _, node := range when.Body {
			g.genLiquidNode(node)
		}
		g.indent--
	}

	if len(tag.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("} else {\n")
		g.indent++
		for _, node := range tag.Else {
			g.genLiquidNode(node)
		}
		g.indent--
	}

	if len(tag.Whens) > 0 {
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

// genLiquidAssignTag generates code for an assign tag.
func (g *Generator) genLiquidAssignTag(tag *liquid.AssignTag) {
	g.writeIndent()
	g.buf.WriteString("ctx.Set(")
	g.buf.WriteString(strconv.Quote(tag.Variable))
	g.buf.WriteString(", ")
	g.genLiquidExpr(tag.Value)
	g.buf.WriteString(")\n")
}

// genLiquidCaptureTag generates code for a capture tag.
func (g *Generator) genLiquidCaptureTag(tag *liquid.CaptureTag) {
	// Save content written so far
	preVar := fmt.Sprintf("_lcpre%d", g.tempVarCounter)
	g.tempVarCounter++

	g.writeIndent()
	g.buf.WriteString(preVar)
	g.buf.WriteString(" := buf.String()\n")

	// Create new builder for capture
	g.writeIndent()
	g.buf.WriteString("buf = strings.Builder{}\n")

	// Generate body (writes to buf)
	for _, node := range tag.Body {
		g.genLiquidNode(node)
	}

	// Save captured content to context variable
	g.writeIndent()
	g.buf.WriteString("ctx.Set(")
	g.buf.WriteString(strconv.Quote(tag.Variable))
	g.buf.WriteString(", buf.String())\n")

	// Restore buffer with original content
	g.writeIndent()
	g.buf.WriteString("buf = strings.Builder{}\n")
	g.writeIndent()
	g.buf.WriteString("buf.WriteString(")
	g.buf.WriteString(preVar)
	g.buf.WriteString(")\n")
}
