package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nchapman/rugby/ast"
	"github.com/nchapman/rugby/stdlib/template"
)

// templateCompileHandler handles compile-time processing for template.compile() and template.compile_file().
type templateCompileHandler struct{}

// Generate implements CompileTimeHandler for template compilation.
func (h *templateCompileHandler) Generate(g *Generator, name string, call *ast.CallExpr, method string) {
	// Require at least one argument
	if len(call.Args) == 0 {
		g.addError(fmt.Errorf("template.%s requires a string argument", method))
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
		g.addError(fmt.Errorf("unknown template compile method: %s", method))
	}
}

// generateSingleTemplate handles template.compile("template string").
func (h *templateCompileHandler) generateSingleTemplate(g *Generator, name string, call *ast.CallExpr) {
	strLit, ok := call.Args[0].(*ast.StringLit)
	if !ok {
		g.addError(fmt.Errorf("template.compile requires a string literal argument"))
		return
	}

	tmpl, err := template.Parse(strLit.Value)
	if err != nil {
		g.addError(fmt.Errorf("template syntax error: %v", err))
		return
	}

	g.genTemplateCompiledTemplate(name, tmpl)
}

// generateSingleTemplateFromFile handles template.compile_file("path").
func (h *templateCompileHandler) generateSingleTemplateFromFile(g *Generator, name string, call *ast.CallExpr) {
	strLit, ok := call.Args[0].(*ast.StringLit)
	if !ok {
		g.addError(fmt.Errorf("template.compile_file requires a string literal file path"))
		return
	}

	filePath := strLit.Value
	if !filepath.IsAbs(filePath) && g.sourceDir != "" {
		filePath = filepath.Join(g.sourceDir, filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		g.addError(fmt.Errorf("template.compile_file: cannot read file %q: %v", strLit.Value, err))
		return
	}

	tmpl, err := template.Parse(string(content))
	if err != nil {
		g.addError(fmt.Errorf("template syntax error in %q: %v", strLit.Value, err))
		return
	}

	g.genTemplateCompiledTemplate(name, tmpl)
}

// generateTemplateMap handles template.compile_dir("path") and template.compile_glob("pattern").
// compile_dir recursively walks the directory and includes only template files (.tmpl or .liquid).
// compile_glob matches files according to the glob pattern with no extension filtering.
func (h *templateCompileHandler) generateTemplateMap(g *Generator, name string, call *ast.CallExpr, isGlob bool) {
	strLit, ok := call.Args[0].(*ast.StringLit)
	if !ok {
		methodName := "compile_dir"
		argType := "directory path"
		if isGlob {
			methodName = "compile_glob"
			argType = "glob pattern"
		}
		g.addError(fmt.Errorf("template.%s requires a string literal %s", methodName, argType))
		return
	}

	pattern := strLit.Value
	basePath := pattern

	// Resolve relative path
	if !filepath.IsAbs(basePath) && g.sourceDir != "" {
		basePath = filepath.Join(g.sourceDir, pattern)
	}

	// Collect templates
	templates := make(map[string]*template.Template)

	if isGlob {
		// compile_glob: use filepath.Glob
		matches, err := filepath.Glob(basePath)
		if err != nil {
			g.addError(fmt.Errorf("template.compile_glob: invalid pattern %q: %v", pattern, err))
			return
		}
		if len(matches) == 0 {
			g.addError(fmt.Errorf("template.compile_glob: no files matched pattern %q", pattern))
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
				g.addError(fmt.Errorf("template.compile_glob: cannot read file %q: %v", filePath, err))
				return
			}

			tmpl, err := template.Parse(string(content))
			if err != nil {
				relPath, _ := filepath.Rel(baseDir, filePath)
				g.addError(fmt.Errorf("template syntax error in %q: %v", relPath, err))
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
			g.addError(fmt.Errorf("template.compile_dir: cannot access directory %q: %v", pattern, err))
			return
		}
		if !info.IsDir() {
			g.addError(fmt.Errorf("template.compile_dir: %q is not a directory", pattern))
			return
		}

		err = filepath.WalkDir(basePath, func(filePath string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			// Filter for template files (.tmpl or .liquid)
			if !isTemplateFile(filePath) {
				return nil
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("cannot read file %q: %v", filePath, err)
			}

			tmpl, err := template.Parse(string(content))
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
			g.addError(fmt.Errorf("template.compile_dir: %v", err))
			return
		}

		if len(templates) == 0 {
			g.addError(fmt.Errorf("template.compile_dir: no template files (.tmpl or .liquid) found in %q", pattern))
			return
		}
	}

	g.genTemplateTemplateMap(name, templates)
}

// isTemplateFile returns true if the file has a recognized template extension.
func isTemplateFile(path string) bool {
	return strings.HasSuffix(path, ".tmpl") || strings.HasSuffix(path, ".liquid")
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

// genTemplateCompiledTemplate generates a single template.CompiledTemplate variable.
func (g *Generator) genTemplateCompiledTemplate(name string, tmpl *template.Template) {
	g.needsTemplate = true
	g.needsStrings = true

	// Register the type in constTypes so inferTypeFromExpr can find it across function boundaries.
	g.constTypes[name] = "template.CompiledTemplate"

	g.writeIndent()
	g.buf.WriteString("var ")
	g.buf.WriteString(name)
	g.buf.WriteString(" = template.CompiledTemplate{\n")
	g.indent++
	g.writeIndent()
	g.buf.WriteString("Render: func(data map[string]any) (string, error) {\n")
	g.indent++

	g.genTemplateRenderBody(tmpl)

	g.indent--
	g.writeIndent()
	g.buf.WriteString("},\n")
	g.indent--
	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genTemplateTemplateMap generates a map[string]*template.CompiledTemplate variable.
// Uses pointers so that MustRender (pointer receiver) can be called on map elements.
func (g *Generator) genTemplateTemplateMap(name string, templates map[string]*template.Template) {
	g.needsTemplate = true
	g.needsStrings = true

	// Register the type in constTypes so inferTypeFromExpr can find it across function boundaries.
	// This is necessary because g.vars is cleared for each function, but compile-time generated
	// constants are module-level and need their types available throughout the entire module.
	g.constTypes[name] = "map[string]*template.CompiledTemplate"

	g.writeIndent()
	g.buf.WriteString("var ")
	g.buf.WriteString(name)
	g.buf.WriteString(" = map[string]*template.CompiledTemplate{\n")
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

		g.genTemplateRenderBody(tmpl)

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

// genTemplateRenderBody generates the body of a Render function for a parsed template.
func (g *Generator) genTemplateRenderBody(tmpl *template.Template) {
	g.writeIndent()
	g.buf.WriteString("var buf strings.Builder\n")
	g.writeIndent()
	g.buf.WriteString("ctx := template.NewContext(data)\n")

	for _, node := range tmpl.AST() {
		g.genTemplateNode(node)
	}

	g.writeIndent()
	g.buf.WriteString("_ = ctx\n")
	g.writeIndent()
	g.buf.WriteString("return buf.String(), nil\n")
}

// genTemplateNode generates Go code for a single template AST node.
func (g *Generator) genTemplateNode(node template.Node) {
	switch n := node.(type) {
	case *template.TextNode:
		g.genTemplateTextNode(n)
	case *template.OutputNode:
		g.genTemplateOutputNode(n)
	case *template.IfTag:
		g.genTemplateIfTag(n)
	case *template.UnlessTag:
		g.genTemplateUnlessTag(n)
	case *template.ForTag:
		g.genTemplateForTag(n)
	case *template.CaseTag:
		g.genTemplateCaseTag(n)
	case *template.AssignTag:
		g.genTemplateAssignTag(n)
	case *template.CaptureTag:
		g.genTemplateCaptureTag(n)
	case *template.BreakTag:
		g.writeIndent()
		g.buf.WriteString("break\n")
	case *template.ContinueTag:
		g.writeIndent()
		g.buf.WriteString("continue\n")
	case *template.CommentTag:
		// Comments produce no output
	case *template.RawTag:
		g.writeIndent()
		g.buf.WriteString("buf.WriteString(")
		g.buf.WriteString(strconv.Quote(n.Content))
		g.buf.WriteString(")\n")
	}
}

// genTemplateTextNode generates code for a TextNode.
func (g *Generator) genTemplateTextNode(n *template.TextNode) {
	if n.Text == "" {
		return
	}
	g.writeIndent()
	g.buf.WriteString("buf.WriteString(")
	g.buf.WriteString(strconv.Quote(n.Text))
	g.buf.WriteString(")\n")
}

// genTemplateOutputNode generates code for an OutputNode ({{ expression }}).
func (g *Generator) genTemplateOutputNode(n *template.OutputNode) {
	g.writeIndent()
	g.buf.WriteString("buf.WriteString(template.ToString(")
	g.genTemplateExpr(n.Expr)
	g.buf.WriteString("))\n")
}

// genTemplateExpr generates Go code for a template expression.
func (g *Generator) genTemplateExpr(expr template.Expression) {
	switch e := expr.(type) {
	case *template.IdentExpr:
		g.buf.WriteString("ctx.Get(")
		g.buf.WriteString(strconv.Quote(e.Name))
		g.buf.WriteString(")")

	case *template.LiteralExpr:
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

	case *template.DotExpr:
		g.buf.WriteString("template.GetProperty(")
		g.genTemplateExpr(e.Object)
		g.buf.WriteString(", ")
		g.buf.WriteString(strconv.Quote(e.Property))
		g.buf.WriteString(")")

	case *template.IndexExpr:
		g.buf.WriteString("template.GetIndex(")
		g.genTemplateExpr(e.Object)
		g.buf.WriteString(", ")
		g.genTemplateExpr(e.Index)
		g.buf.WriteString(")")

	case *template.FilterExpr:
		g.genTemplateFilterExpr(e)

	case *template.BinaryExpr:
		g.genTemplateBinaryExpr(e)

	case *template.RangeExpr:
		g.buf.WriteString("template.MakeRange(")
		g.genTemplateExpr(e.Start)
		g.buf.WriteString(", ")
		g.genTemplateExpr(e.End)
		g.buf.WriteString(")")

	default:
		// Unknown expression type - generate nil
		g.buf.WriteString("nil")
	}
}

// genTemplateFilterExpr generates code for a filter expression.
func (g *Generator) genTemplateFilterExpr(e *template.FilterExpr) {
	switch e.Name {
	case "upcase":
		g.buf.WriteString("template.FilterUpcase(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(")")
	case "downcase":
		g.buf.WriteString("template.FilterDowncase(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(")")
	case "capitalize":
		g.buf.WriteString("template.FilterCapitalize(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(")")
	case "strip":
		g.buf.WriteString("template.FilterStrip(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(")")
	case "escape":
		g.buf.WriteString("template.FilterEscape(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(")")
	case "first":
		g.buf.WriteString("template.FilterFirst(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(")")
	case "last":
		g.buf.WriteString("template.FilterLast(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(")")
	case "size":
		g.buf.WriteString("template.FilterSize(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(")")
	case "reverse":
		g.buf.WriteString("template.FilterReverse(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(")")
	case "join":
		g.buf.WriteString("template.FilterJoin(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(", ")
		if len(e.Args) > 0 {
			g.buf.WriteString("template.ToString(")
			g.genTemplateExpr(e.Args[0])
			g.buf.WriteString(")")
		} else {
			g.buf.WriteString(`" "`)
		}
		g.buf.WriteString(")")
	case "default":
		g.buf.WriteString("template.FilterDefault(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(", ")
		if len(e.Args) > 0 {
			g.genTemplateExpr(e.Args[0])
		} else {
			g.buf.WriteString(`""`)
		}
		g.buf.WriteString(")")
	case "plus":
		g.buf.WriteString("template.FilterPlus(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(", ")
		if len(e.Args) > 0 {
			g.genTemplateExpr(e.Args[0])
		} else {
			g.buf.WriteString("0")
		}
		g.buf.WriteString(")")
	case "minus":
		g.buf.WriteString("template.FilterMinus(")
		g.genTemplateExpr(e.Input)
		g.buf.WriteString(", ")
		if len(e.Args) > 0 {
			g.genTemplateExpr(e.Args[0])
		} else {
			g.buf.WriteString("0")
		}
		g.buf.WriteString(")")
	default:
		// Unknown filter - just pass through input
		g.genTemplateExpr(e.Input)
	}
}

// genTemplateBinaryExpr generates code for a binary expression.
func (g *Generator) genTemplateBinaryExpr(e *template.BinaryExpr) {
	switch e.Operator {
	case "and":
		g.buf.WriteString("(template.ToBool(")
		g.genTemplateExpr(e.Left)
		g.buf.WriteString(") && template.ToBool(")
		g.genTemplateExpr(e.Right)
		g.buf.WriteString("))")
	case "or":
		g.buf.WriteString("(template.ToBool(")
		g.genTemplateExpr(e.Left)
		g.buf.WriteString(") || template.ToBool(")
		g.genTemplateExpr(e.Right)
		g.buf.WriteString("))")
	default:
		// Use CompareValues for comparison operators
		g.buf.WriteString("template.CompareValues(")
		g.genTemplateExpr(e.Left)
		g.buf.WriteString(", ")
		g.buf.WriteString(strconv.Quote(e.Operator))
		g.buf.WriteString(", ")
		g.genTemplateExpr(e.Right)
		g.buf.WriteString(")")
	}
}

// genTemplateIfTag generates code for an if tag.
func (g *Generator) genTemplateIfTag(tag *template.IfTag) {
	g.writeIndent()
	g.buf.WriteString("if template.ToBool(")
	g.genTemplateExpr(tag.Condition)
	g.buf.WriteString(") {\n")
	g.indent++
	for _, node := range tag.ThenBranch {
		g.genTemplateNode(node)
	}
	g.indent--

	// Handle elsif branches
	for _, elsif := range tag.ElsifBranches {
		g.writeIndent()
		g.buf.WriteString("} else if template.ToBool(")
		g.genTemplateExpr(elsif.Condition)
		g.buf.WriteString(") {\n")
		g.indent++
		for _, node := range elsif.Body {
			g.genTemplateNode(node)
		}
		g.indent--
	}

	// Handle else branch
	if len(tag.ElseBranch) > 0 {
		g.writeIndent()
		g.buf.WriteString("} else {\n")
		g.indent++
		for _, node := range tag.ElseBranch {
			g.genTemplateNode(node)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genTemplateUnlessTag generates code for an unless tag.
func (g *Generator) genTemplateUnlessTag(tag *template.UnlessTag) {
	g.writeIndent()
	g.buf.WriteString("if !template.ToBool(")
	g.genTemplateExpr(tag.Condition)
	g.buf.WriteString(") {\n")
	g.indent++
	for _, node := range tag.Body {
		g.genTemplateNode(node)
	}
	g.indent--

	// Handle else branch
	if len(tag.ElseBranch) > 0 {
		g.writeIndent()
		g.buf.WriteString("} else {\n")
		g.indent++
		for _, node := range tag.ElseBranch {
			g.genTemplateNode(node)
		}
		g.indent--
	}

	g.writeIndent()
	g.buf.WriteString("}\n")
}

// genTemplateForTag generates code for a for loop.
func (g *Generator) genTemplateForTag(tag *template.ForTag) {
	// Generate a unique variable name for the collection
	collectionVar := fmt.Sprintf("_lfc%d", g.tempVarCounter)
	g.tempVarCounter++

	// Get the collection
	g.writeIndent()
	g.buf.WriteString(collectionVar)
	g.buf.WriteString(" := template.ToSlice(")
	g.genTemplateExpr(tag.Collection)
	g.buf.WriteString(")\n")

	// Handle offset
	if tag.Offset != nil {
		g.writeIndent()
		g.buf.WriteString("if _off := template.ToIntValue(")
		g.genTemplateExpr(tag.Offset)
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
		g.buf.WriteString("if _lim := template.ToIntValue(")
		g.genTemplateExpr(tag.Limit)
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
			g.genTemplateNode(node)
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
		g.genTemplateNode(node)
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

// genTemplateCaseTag generates code for a case tag.
func (g *Generator) genTemplateCaseTag(tag *template.CaseTag) {
	valueVar := fmt.Sprintf("_lcv%d", g.tempVarCounter)
	g.tempVarCounter++

	g.writeIndent()
	g.buf.WriteString(valueVar)
	g.buf.WriteString(" := ")
	g.genTemplateExpr(tag.Value)
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
			g.buf.WriteString("template.CompareValues(")
			g.buf.WriteString(valueVar)
			g.buf.WriteString(", \"==\", ")
			g.genTemplateExpr(val)
			g.buf.WriteString(")")
		}

		g.buf.WriteString(" {\n")
		g.indent++
		for _, node := range when.Body {
			g.genTemplateNode(node)
		}
		g.indent--
	}

	if len(tag.Else) > 0 {
		g.writeIndent()
		g.buf.WriteString("} else {\n")
		g.indent++
		for _, node := range tag.Else {
			g.genTemplateNode(node)
		}
		g.indent--
	}

	if len(tag.Whens) > 0 {
		g.writeIndent()
		g.buf.WriteString("}\n")
	}
}

// genTemplateAssignTag generates code for an assign tag.
func (g *Generator) genTemplateAssignTag(tag *template.AssignTag) {
	g.writeIndent()
	g.buf.WriteString("ctx.Set(")
	g.buf.WriteString(strconv.Quote(tag.Variable))
	g.buf.WriteString(", ")
	g.genTemplateExpr(tag.Value)
	g.buf.WriteString(")\n")
}

// genTemplateCaptureTag generates code for a capture tag.
func (g *Generator) genTemplateCaptureTag(tag *template.CaptureTag) {
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
		g.genTemplateNode(node)
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
