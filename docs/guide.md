# Rugby Compiler Developer Guide

This guide explains how to correctly add new language features to the Rugby compiler.

## Architecture Overview

The Rugby compiler is a **5-stage pipeline**:

```
Source (.rg) → Lexer → Parser → Semantic Analysis → CodeGen → Go (.go)
```

| Stage | Package | Responsibility |
|-------|---------|----------------|
| **Lexer** | `lexer/` | Tokenizes source into tokens with position tracking |
| **Parser** | `parser/` | Builds AST from tokens using recursive descent |
| **Semantic** | `semantic/` | Type inference, validation, symbol resolution |
| **CodeGen** | `codegen/` | Emits Go code from typed AST |
| **Builder** | `internal/builder/` | Orchestrates pipeline, invokes `go build` |

### The Golden Rule

> **Semantic analysis owns all type inference. CodeGen is a pure emitter.**

This separation is critical. When adding features:
- Put type logic in `semantic/`
- Put code emission in `codegen/`
- Never mix them

## The Type System

### Semantic Types (`semantic/types.go`)

The analyzer tracks types for every AST node:

```go
type Type struct {
    Kind       TypeKind    // Int, String, Array, Optional, Class, etc.
    Elem       *Type       // Element type (Array[T], Optional[T])
    KeyType    *Type       // Map key type
    ValueType  *Type       // Map value type
    Name       string      // Class/Interface name
    Params     []*Type     // Function parameters
    Returns    []*Type     // Function return types
}
```

### TypeInfo Interface (`codegen/codegen.go`)

CodeGen accesses semantic information through this interface:

```go
type TypeInfo interface {
    GetTypeKind(node ast.Node) TypeKind     // For optimization decisions
    GetGoType(node ast.Node) string         // "int", "[]string", etc.
    GetRugbyType(node ast.Node) string      // "Int", "Array[String]", etc.
    GetSelectorKind(node ast.Node) SelectorKind  // Field vs method
    GetElementType(node ast.Node) string    // Inner type for composites
    GetKeyValueTypes(node ast.Node) (string, string)  // Map types
    IsVariableUsedAt(node ast.Node, name string) bool
}
```

### Information Flow

```
Parser                  Semantic                CodeGen
  │                        │                       │
  ├─► AST nodes ──────────►│                       │
  │                        ├─► Type inference      │
  │                        ├─► Symbol resolution   │
  │                        ├─► Validation          │
  │                        │                       │
  │                        ├─► nodeTypes map ─────►│ (via TypeInfo)
  │                        ├─► selectorKinds ─────►│
  │                        │                       │
  │                        │                       ├─► Go code
```

## Adding a New Language Feature

### Step 1: Write the Failing Spec Test

Always start with a spec test in `tests/spec/errors/`:

```ruby
#@ compile-fail
#
# Test new_feature syntax
#
# Expected behavior:
#   - What it should do
#   - Example usage
#
# Currently fails with: <expected error>

class Example
  # new syntax here
end

#~ ERROR: <pattern matching expected error>
```

### Step 2: Update the AST (`ast/ast.go`)

Add or modify AST nodes to represent the new syntax:

```go
// Example: Adding IsClassMethod to MethodDecl
type MethodDecl struct {
    Name          string
    Params        []*Param
    ReturnTypes   []string
    Body          []Statement
    Pub           bool
    IsClassMethod bool  // ← New field
    Line          int
}
```

**Guidelines:**
- AST nodes are data structures, not behavior
- Include all syntactic information needed downstream
- Add position info (Line, Column) for error messages

### Step 3: Update the Parser (`parser/`)

Modify parsing to recognize the new syntax:

```go
// Example: Parsing def self.method
func (p *Parser) parseMethodDeclWithDoc(doc *ast.CommentGroup) *ast.MethodDecl {
    p.nextToken() // consume 'def'

    // Check for class method (def self.method_name)
    isClassMethod := false
    if p.curTokenIs(token.SELF) {
        isClassMethod = true
        p.nextToken() // consume 'self'
        if !p.curTokenIs(token.DOT) {
            p.errorAt(p.curToken.Line, p.curToken.Column,
                "expected '.' after self in class method")
            return nil
        }
        p.nextToken() // consume '.'
    }

    // ... rest of parsing

    return &ast.MethodDecl{
        Name:          methodName,
        IsClassMethod: isClassMethod,  // ← Set the flag
        // ...
    }
}
```

**Guidelines:**
- Parser only builds AST, no type checking
- Report clear syntax errors with positions
- Handle edge cases (trailing commas, optional parts)

### Step 4: Update Semantic Analysis (`semantic/`)

This is where the heavy lifting happens. Add type inference and validation:

```go
// Example: Analyzing class methods
func (a *Analyzer) analyzeMethodDecl(className string, method *ast.MethodDecl) {
    // 1. Create scope for method body
    a.pushScope()
    defer a.popScope()

    // 2. For instance methods, 'self' refers to the instance
    //    For class methods, 'self' is not available
    if !method.IsClassMethod {
        selfType := NewClassType(className)
        a.scope.Define("self", NewVariable("self", selfType))
    }

    // 3. Define parameters in scope with their types
    for _, param := range method.Params {
        paramType := a.resolveType(param.Type)
        a.scope.Define(param.Name, NewVariable(param.Name, paramType))
    }

    // 4. Analyze body statements
    for _, stmt := range method.Body {
        a.analyzeStmt(stmt)
    }

    // 5. Validate return type matches declared type
    if len(method.ReturnTypes) > 0 {
        actualReturn := a.inferBlockReturnType(method.Body)
        expectedReturn := a.resolveType(method.ReturnTypes[0])
        if !a.isAssignable(actualReturn, expectedReturn) {
            a.addError(/* type mismatch error */)
        }
    }

    // 6. Store type information for codegen
    a.setNodeType(method, NewFunctionType(params, returns))
}
```

**Guidelines:**
- Infer types for ALL expressions via `analyzeExpr()`
- Store results with `setNodeType(node, type)`
- Validate constraints and report errors
- Track symbols in scopes for resolution

### Step 5: Update CodeGen (`codegen/`)

Finally, emit Go code using the type information:

```go
// Example: Generating class methods
func (g *Generator) genClassMethod(className string, method *ast.MethodDecl) {
    // Query type info - DON'T infer types here!
    returnType := ""
    if g.typeInfo != nil {
        returnType = g.typeInfo.GetGoType(method)
    }

    // Generate function signature
    funcName := className + "_" + snakeToCamel(method.Name)
    g.buf.WriteString(fmt.Sprintf("func %s(", funcName))

    // Parameters
    for i, param := range method.Params {
        if i > 0 {
            g.buf.WriteString(", ")
        }
        paramType := g.mapParamType(param.Type)
        g.buf.WriteString(fmt.Sprintf("%s %s", param.Name, paramType))
    }
    g.buf.WriteString(")")

    // Return type
    if returnType != "" {
        g.buf.WriteString(" " + returnType)
    }

    // Body
    g.buf.WriteString(" {\n")
    g.indent++
    for _, stmt := range method.Body {
        g.genStatement(stmt)
    }
    g.indent--
    g.buf.WriteString("}\n")
}
```

**Guidelines:**
- Query `typeInfo` for type decisions
- Don't infer types - use what semantic analysis computed
- Focus on Go syntax emission
- Handle edge cases in output formatting

### Step 6: Move Test and Verify

Once implementation works:

1. Move test from `tests/spec/errors/` to appropriate directory
2. Change `#@ compile-fail` to `#@ run-pass`
3. Add `#@ check-output` with expected output
4. Run `make check` to verify

## Anti-Patterns to Avoid

### ❌ Type Inference in CodeGen

**Bad:**
```go
// In codegen - DON'T DO THIS
func (g *Generator) genArrayLit(arr *ast.ArrayLit) {
    // Inferring element type during code generation
    elemType := "any"
    if len(arr.Elements) > 0 {
        elemType = g.guessTypeFromExpr(arr.Elements[0])
    }
    g.buf.WriteString(fmt.Sprintf("[]%s{", elemType))
}
```

**Good:**
```go
// In codegen - DO THIS
func (g *Generator) genArrayLit(arr *ast.ArrayLit) {
    // Query pre-computed type from semantic analysis
    elemType := "any"
    if g.typeInfo != nil {
        elemType = g.typeInfo.GetElementType(arr)
    }
    g.buf.WriteString(fmt.Sprintf("[]%s{", elemType))
}
```

### ❌ Validation in CodeGen

**Bad:**
```go
// In codegen - DON'T DO THIS
func (g *Generator) genMethodDecl(method *ast.MethodDecl) {
    if strings.HasSuffix(method.Name, "?") {
        if method.ReturnTypes[0] != "Bool" {
            g.addError("predicate must return Bool")  // Wrong place!
        }
    }
}
```

**Good:**
```go
// In semantic - DO THIS
func (a *Analyzer) analyzeMethodDecl(method *ast.MethodDecl) {
    if strings.HasSuffix(method.Name, "?") {
        if method.ReturnTypes[0] != "Bool" {
            a.addError(&TypeError{
                Message: "predicate method must return Bool",
                Line:    method.Line,
            })
        }
    }
}
```

### ❌ Symbol Resolution in CodeGen

**Bad:**
```go
// In codegen - DON'T DO THIS
func (g *Generator) genIdent(ident *ast.Ident) {
    // Looking up what 'ident' refers to during codegen
    if g.classes[ident.Name] {
        // it's a class
    } else if g.vars[ident.Name] != "" {
        // it's a variable
    } else {
        // unknown - guess?
    }
}
```

**Good:**
```go
// In semantic - resolve and tag the AST node
func (a *Analyzer) analyzeIdent(ident *ast.Ident) *Type {
    sym := a.scope.Lookup(ident.Name)
    if sym == nil {
        a.addError(&UndefinedError{Name: ident.Name, Line: ident.Line})
        return TypeUnknownVal
    }
    a.setNodeType(ident, sym.Type)
    return sym.Type
}

// In codegen - just use the resolved info
func (g *Generator) genIdent(ident *ast.Ident) {
    goType := g.typeInfo.GetGoType(ident)
    // Generate code knowing exactly what this identifier is
}
```

## Testing Your Changes

### Unit Tests

Add tests for each layer:

```bash
# Parser tests
go test ./parser -run TestYourFeature -v

# Semantic tests
go test ./semantic -run TestYourFeature -v

# Codegen tests
go test ./codegen -run TestYourFeature -v
```

### Integration Tests (Spec Tests)

```bash
# Run all spec tests
make test-spec

# Run specific test
go test ./tests/... -run "TestSpecs/category/test_name" -v

# Update golden files if output changed
make test-spec-bless
```

### Full Check

```bash
# Run everything before committing
make check
```

## Checklist for New Features

- [ ] Spec test written (starts as `compile-fail`)
- [ ] AST node added/modified
- [ ] Parser updated with clear error messages
- [ ] Semantic analysis: type inference added
- [ ] Semantic analysis: validation added
- [ ] Semantic analysis: `setNodeType()` called for new nodes
- [ ] CodeGen queries `typeInfo` (no inference)
- [ ] CodeGen emits correct Go code
- [ ] Unit tests for parser, semantic, codegen
- [ ] Spec test passes (`run-pass`)
- [ ] `make check` passes
- [ ] Code reviewed

## Common Patterns

### Adding a New Expression Type

1. **AST**: Add struct in `ast/ast.go`
2. **Parser**: Add case in `parseExpression()` or `parsePrefixExpression()`
3. **Semantic**: Add case in `analyzeExpr()`, call `setNodeType()`
4. **CodeGen**: Add case in `genExpr()`

### Adding a New Statement Type

1. **AST**: Add struct in `ast/ast.go`
2. **Parser**: Add case in `parseStatement()`
3. **Semantic**: Add case in `analyzeStmt()`
4. **CodeGen**: Add case in `genStatement()`

### Adding a New Declaration Type

1. **AST**: Add struct in `ast/ast.go`
2. **Parser**: Add to `parseDeclaration()` or class/module body parsing
3. **Semantic**: Add to `analyzeDecl()` or appropriate container analysis
4. **CodeGen**: Add to `genDeclaration()` or appropriate container generation

### Adding a Built-in Method

1. **Semantic**: Add to `getBuiltinMethodType()` in `analyzer.go`
2. **CodeGen**: Add to `genCallExpr()` method handling or `genBlockCall()`
3. **Runtime**: Add Go implementation in `runtime/` package if needed

## Debugging Tips

### See Generated Go Code

```bash
# The .go file is in .rugby/ directory
go run . your_file.rg
cat .rugby/your_file.go
```

### Print AST

```go
// In parser tests
ast.Print(program)  // if available, or use %+v
```

### Print Type Info

```go
// In semantic analyzer
fmt.Printf("Type of %T: %v\n", node, a.GetType(node))
```

### Trace CodeGen

```go
// Add to Generator methods temporarily
fmt.Printf("genExpr: %T\n", expr)
```
