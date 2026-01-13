# Rugby Type System Design

## Overview

This document describes the design for a comprehensive type system that will provide rich type information throughout the Rugby compiler pipeline. The goal is to enable proper handling of:

- Method calls without parentheses (distinguishing methods from fields)
- Setter/getter property access
- Go package interop with full type awareness
- Proper type checking and inference
- Accurate error messages

## Current Problems

1. **No distinction between fields and methods** - `obj.foo` is generated as field access even when `foo` is a method
2. **No setter recognition** - `obj.prop = value` doesn't invoke setter methods
3. **Limited Go interop** - We don't know what types/methods Go packages export
4. **Type info doesn't flow to codegen** - Semantic analysis results aren't fully utilized
5. **Incomplete type inference** - Many expressions have unknown types

## Bugs This Will Fix

From `examples/BUGS.md`, this type system will directly fix:

| Bug | Issue | How Type System Fixes It |
|-----|-------|--------------------------|
| BUG-026 | Method calls without parentheses | SelectorExpr annotated with SelectorMethod triggers `()` generation |
| BUG-027 | Compound assignment to instance variable | Proper field type tracking enables correct `+=` codegen |
| BUG-028 | to_s call resolves to wrong method name | Method registry maps `to_s` → `String()` in codegen |
| BUG-029 | Subclass constructors not generated | ClassDef.Parent enables inherited constructor generation |
| BUG-030 | Setter assignment not working | SelectorKind=SelectorSetter triggers `setX()` call |
| BUG-017 | Predicate methods on arrays | Array method registry includes `any?`, `empty?` |
| BUG-022 | String methods not implemented | String type includes method registry |
| BUG-023 | Range.size method | Range type includes method registry |

Additionally enables future features:
- Better error messages ("no method 'foo' on type Bar")
- IDE integration (autocomplete, go-to-definition)
- Proper generics support
- Exhaustive pattern matching

## Relationship to Existing Code

### Current `semantic/` Package

The existing semantic analyzer (`semantic/analyzer.go`) already does:
- Scope management with `Scope` and `Variable`
- Some type checking (function parameters, return types)
- Interface satisfaction checking
- Error collection

**What we keep:**
- Error types and formatting
- Basic scope structure
- Overall analysis flow

**What we enhance:**
- `Variable` becomes richer (full Type instead of just name)
- Add TypeRegistry construction
- Add AST annotation
- Add Go package loading

### Current `codegen/` Package

The codegen currently has ad-hoc type inference:
- `inferTypeFromExpr()` - guesses types from expressions
- `isGoInterop()` - checks if expression involves Go package
- `classFields` map - tracks current class fields
- `interfaceMethods` map - tracks interface methods

**What we remove:**
- Most of `inferTypeFromExpr()` - use annotations instead
- Ad-hoc method/field detection
- Duplicate tracking (already in registry)

**What we add:**
- Use `SelectorExpr.ResolvedKind` for method vs field
- Use `CallExpr.ResolvedFunc` for return types
- Use `Registry` for class/interface lookup

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│   Lexer     │────▶│   Parser    │────▶│    Analyzer     │────▶│   Codegen   │
└─────────────┘     └─────────────┘     └─────────────────┘     └─────────────┘
                           │                     │                      │
                           ▼                     ▼                      ▼
                    ┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
                    │     AST     │     │  TypeRegistry   │     │  Generated  │
                    │  (raw)      │     │  + Annotated    │     │  Go Code    │
                    └─────────────┘     │     AST         │     └─────────────┘
                                        └─────────────────┘
                                                 ▲
                                                 │
                                        ┌─────────────────┐
                                        │  Go Importer    │
                                        │  (go/types)     │
                                        └─────────────────┘
```

## Core Components

### 1. Type Representation (`types/types.go`)

```go
package types

// Kind represents the category of a type
type Kind int

const (
    KindUnknown Kind = iota
    KindVoid         // no return value
    KindNil          // nil literal
    KindBool
    KindInt
    KindFloat
    KindString
    KindArray
    KindMap
    KindOptional
    KindFunc
    KindClass
    KindInterface
    KindModule
    KindGoType       // imported Go type
    KindAny          // any/interface{}
)

// Type represents a Rugby type
type Type interface {
    Kind() Kind
    String() string      // human-readable representation
    GoType() string      // Go type string for codegen
    Equals(other Type) bool
    AssignableTo(target Type) bool
}

// PrimitiveType for bool, int, float, string
type PrimitiveType struct {
    kind Kind
}

// ArrayType for []T
type ArrayType struct {
    Element Type
}

// MapType for map[K]V
type MapType struct {
    Key   Type
    Value Type
}

// OptionalType for T?
type OptionalType struct {
    Inner Type
}

// FuncType for function signatures
type FuncType struct {
    Params  []Param
    Returns []Type
}

type Param struct {
    Name     string
    Type     Type
    Optional bool    // has default value
}

// ClassType references a class definition
type ClassType struct {
    Name    string
    Def     *ClassDef  // resolved reference
    Pointer bool       // *ClassName vs ClassName
}

// InterfaceType references an interface definition
type InterfaceType struct {
    Name string
    Def  *InterfaceDef
}

// GoType represents an imported Go type
type GoType struct {
    Package  string     // e.g., "fmt", "net/http"
    Name     string     // e.g., "Stringer", "Request"
    GoTypes  *types.Type // from go/types, for full reflection
}
```

### 2. Symbol Definitions (`types/symbols.go`)

```go
package types

// ClassDef holds complete information about a class
type ClassDef struct {
    Name       string
    Parent     *ClassDef              // nil if no inheritance
    Implements []*InterfaceDef        // interfaces this class satisfies
    Fields     map[string]*FieldDef
    Methods    map[string]*MethodDef
    IsPublic   bool                   // pub class
}

// FieldDef describes a class field
type FieldDef struct {
    Name       string
    Type       Type
    HasGetter  bool    // getter accessor declared
    HasSetter  bool    // property accessor declared (read+write)
    IsPublic   bool    // pub field
}

// MethodDef describes a class method
type MethodDef struct {
    Name       string
    Params     []Param
    Returns    []Type
    IsPublic   bool
    IsStatic   bool    // class method vs instance method
}

// InterfaceDef holds complete information about an interface
type InterfaceDef struct {
    Name    string
    Parents []*InterfaceDef           // interface inheritance
    Methods map[string]*MethodDef
}

// ModuleDef holds information about a module
type ModuleDef struct {
    Name    string
    Methods map[string]*MethodDef
}

// FuncDef holds information about a top-level function
type FuncDef struct {
    Name     string
    Params   []Param
    Returns  []Type
    IsPublic bool
}
```

### 3. Type Registry (`types/registry.go`)

```go
package types

// Registry is the central repository of all type information
type Registry struct {
    // Rugby definitions
    Classes    map[string]*ClassDef
    Interfaces map[string]*InterfaceDef
    Modules    map[string]*ModuleDef
    Functions  map[string]*FuncDef

    // Go package info (lazily loaded)
    GoPackages map[string]*GoPackageInfo

    // Scope stack for variable resolution
    scopes []*Scope
}

// GoPackageInfo holds reflected information about a Go package
type GoPackageInfo struct {
    Path      string                    // import path
    Types     map[string]*GoTypeDef     // exported types
    Functions map[string]*FuncDef       // exported functions
    Variables map[string]Type           // exported variables/constants
}

// GoTypeDef holds information about a Go type
type GoTypeDef struct {
    Name      string
    Kind      GoTypeKind               // struct, interface, alias, etc.
    Fields    map[string]*FieldDef     // for structs
    Methods   map[string]*MethodDef    // methods with receivers
    Underlying Type                     // for type aliases
}

type GoTypeKind int

const (
    GoTypeStruct GoTypeKind = iota
    GoTypeInterface
    GoTypeAlias
    GoTypeBasic
)

// Scope represents a lexical scope with variable bindings
type Scope struct {
    Parent    *Scope
    Variables map[string]*VarDef
}

type VarDef struct {
    Name string
    Type Type
}
```

### 4. Go Package Importer (`types/goimporter.go`)

```go
package types

import (
    "go/importer"
    "go/types"
)

// GoImporter loads and caches Go package type information
type GoImporter struct {
    cache    map[string]*GoPackageInfo
    importer types.Importer
}

func NewGoImporter() *GoImporter {
    return &GoImporter{
        cache:    make(map[string]*GoPackageInfo),
        importer: importer.Default(),
    }
}

// Import loads a Go package and extracts type information
func (gi *GoImporter) Import(path string) (*GoPackageInfo, error) {
    if cached, ok := gi.cache[path]; ok {
        return cached, nil
    }

    pkg, err := gi.importer.Import(path)
    if err != nil {
        return nil, err
    }

    info := gi.extractPackageInfo(pkg)
    gi.cache[path] = info
    return info, nil
}

// extractPackageInfo converts go/types info to our representation
func (gi *GoImporter) extractPackageInfo(pkg *types.Package) *GoPackageInfo {
    info := &GoPackageInfo{
        Path:      pkg.Path(),
        Types:     make(map[string]*GoTypeDef),
        Functions: make(map[string]*FuncDef),
        Variables: make(map[string]Type),
    }

    scope := pkg.Scope()
    for _, name := range scope.Names() {
        obj := scope.Lookup(name)
        switch o := obj.(type) {
        case *types.TypeName:
            info.Types[name] = gi.extractTypeDef(o)
        case *types.Func:
            info.Functions[name] = gi.extractFuncDef(o)
        case *types.Var, *types.Const:
            info.Variables[name] = gi.convertGoType(o.Type())
        }
    }

    return info
}

// extractTypeDef extracts full type info including methods
func (gi *GoImporter) extractTypeDef(tn *types.TypeName) *GoTypeDef {
    def := &GoTypeDef{
        Name:    tn.Name(),
        Methods: make(map[string]*MethodDef),
        Fields:  make(map[string]*FieldDef),
    }

    typ := tn.Type()

    // Get methods (including pointer receiver methods)
    mset := types.NewMethodSet(types.NewPointer(typ))
    for i := 0; i < mset.Len(); i++ {
        sel := mset.At(i)
        fn := sel.Obj().(*types.Func)
        def.Methods[fn.Name()] = gi.extractMethodDef(fn)
    }

    // Get fields for struct types
    if st, ok := typ.Underlying().(*types.Struct); ok {
        def.Kind = GoTypeStruct
        for i := 0; i < st.NumFields(); i++ {
            f := st.Field(i)
            if f.Exported() {
                def.Fields[f.Name()] = &FieldDef{
                    Name: f.Name(),
                    Type: gi.convertGoType(f.Type()),
                }
            }
        }
    } else if _, ok := typ.Underlying().(*types.Interface); ok {
        def.Kind = GoTypeInterface
    }

    return def
}

// convertGoType converts a go/types.Type to our Type representation
func (gi *GoImporter) convertGoType(t types.Type) Type {
    switch t := t.(type) {
    case *types.Basic:
        switch t.Kind() {
        case types.Bool:
            return PrimitiveBool
        case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
            return PrimitiveInt
        case types.Float32, types.Float64:
            return PrimitiveFloat
        case types.String:
            return PrimitiveString
        }
    case *types.Slice:
        return &ArrayType{Element: gi.convertGoType(t.Elem())}
    case *types.Map:
        return &MapType{
            Key:   gi.convertGoType(t.Key()),
            Value: gi.convertGoType(t.Elem()),
        }
    case *types.Pointer:
        inner := gi.convertGoType(t.Elem())
        if ct, ok := inner.(*ClassType); ok {
            ct.Pointer = true
            return ct
        }
        return inner
    case *types.Named:
        return &GoType{
            Package: t.Obj().Pkg().Path(),
            Name:    t.Obj().Name(),
        }
    }
    return UnknownType
}
```

### 5. Enhanced Semantic Analyzer (`semantic/analyzer.go`)

The analyzer will be enhanced to:

1. **Build the TypeRegistry** during analysis
2. **Resolve all types** for expressions and statements
3. **Annotate AST nodes** with resolved type information
4. **Load Go packages** when import statements are encountered

```go
// Analyzer now maintains a TypeRegistry
type Analyzer struct {
    registry    *types.Registry
    goImporter  *types.GoImporter
    currentFile string
    errors      []error
}

// Analyze processes the AST and builds the type registry
func (a *Analyzer) Analyze(prog *ast.Program) (*types.Registry, error) {
    // Phase 1: Collect all type declarations (classes, interfaces, modules)
    a.collectDeclarations(prog)

    // Phase 2: Resolve all type references and build full definitions
    a.resolveTypes(prog)

    // Phase 3: Load Go package info for imports
    a.loadGoImports(prog)

    // Phase 4: Type-check all expressions and annotate AST
    a.typeCheck(prog)

    return a.registry, a.combineErrors()
}

// resolveExpression returns the type of an expression and annotates the AST
func (a *Analyzer) resolveExpression(expr ast.Expression) Type {
    switch e := expr.(type) {
    case *ast.SelectorExpr:
        return a.resolveSelectorExpr(e)
    case *ast.CallExpr:
        return a.resolveCallExpr(e)
    // ... etc
    }
}

// resolveSelectorExpr determines if obj.sel is a field, method, or Go member
func (a *Analyzer) resolveSelectorExpr(sel *ast.SelectorExpr) Type {
    receiverType := a.resolveExpression(sel.X)

    switch rt := receiverType.(type) {
    case *ClassType:
        // Check if it's a method
        if method, ok := rt.Def.Methods[sel.Sel]; ok {
            sel.ResolvedKind = ast.SelectorMethod
            sel.ResolvedType = method.Returns[0] // simplified
            return sel.ResolvedType
        }
        // Check if it's a field
        if field, ok := rt.Def.Fields[sel.Sel]; ok {
            sel.ResolvedKind = ast.SelectorField
            sel.ResolvedType = field.Type
            return sel.ResolvedType
        }

    case *GoType:
        // Look up in Go package info
        pkgInfo := a.registry.GoPackages[rt.Package]
        typeDef := pkgInfo.Types[rt.Name]
        if method, ok := typeDef.Methods[sel.Sel]; ok {
            sel.ResolvedKind = ast.SelectorGoMethod
            // ...
        }
    }

    return UnknownType
}
```

### 6. AST Annotations (`ast/annotations.go`)

```go
package ast

// SelectorKind indicates what a selector expression refers to
type SelectorKind int

const (
    SelectorUnknown SelectorKind = iota
    SelectorField                // class field access
    SelectorMethod               // class method (needs () call)
    SelectorGetter               // accessor getter (generated method)
    SelectorSetter               // used in assignment context
    SelectorGoField              // Go struct field
    SelectorGoMethod             // Go method
)

// Extend SelectorExpr with resolved info
type SelectorExpr struct {
    X   Expression
    Sel string

    // Resolved during semantic analysis
    ResolvedKind SelectorKind
    ResolvedType types.Type
}

// Extend CallExpr with resolved info
type CallExpr struct {
    Func Expression
    Args []Expression
    Block *BlockExpr

    // Resolved during semantic analysis
    ResolvedFunc  *types.MethodDef  // or FuncDef
    ResolvedTypes []types.Type      // return types
}
```

### 7. Enhanced Codegen (`codegen/codegen.go`)

```go
// Generator now receives the full TypeRegistry
type Generator struct {
    registry *types.Registry
    // ... existing fields
}

func (g *Generator) genSelectorExpr(sel *ast.SelectorExpr) {
    switch sel.ResolvedKind {
    case ast.SelectorField, ast.SelectorGoField:
        // Generate as field access: obj.field
        g.genExpr(sel.X)
        g.buf.WriteString(".")
        g.buf.WriteString(g.goFieldName(sel))

    case ast.SelectorMethod, ast.SelectorGetter:
        // Generate as method call: obj.method()
        g.genExpr(sel.X)
        g.buf.WriteString(".")
        g.buf.WriteString(g.goMethodName(sel))
        g.buf.WriteString("()")

    case ast.SelectorGoMethod:
        // Generate with PascalCase: obj.Method()
        g.genExpr(sel.X)
        g.buf.WriteString(".")
        g.buf.WriteString(sel.Sel) // already PascalCase in Go
        g.buf.WriteString("()")
    }
}

func (g *Generator) genAssignStmt(stmt *ast.AssignStmt) {
    // Check if LHS is a setter
    if sel, ok := stmt.Left.(*ast.SelectorExpr); ok {
        if sel.ResolvedKind == ast.SelectorSetter {
            // Generate as setter call: obj.setField(value)
            g.genExpr(sel.X)
            g.buf.WriteString(".")
            g.buf.WriteString(g.setterName(sel.Sel))
            g.buf.WriteString("(")
            g.genExpr(stmt.Right)
            g.buf.WriteString(")")
            return
        }
    }
    // ... normal assignment
}
```

## Implementation Plan

### Phase 1: Core Type System (Foundation)

1. Create `types/` package with Type interface and implementations
2. Create Registry and Scope structures
3. Create symbol definition types (ClassDef, MethodDef, etc.)
4. Add unit tests for type operations

### Phase 2: Enhance Semantic Analyzer

1. Refactor analyzer to build TypeRegistry
2. Collect class/interface/module declarations in first pass
3. Resolve type references in second pass
4. Store resolved info in registry

### Phase 3: AST Annotations

1. Add annotation fields to AST nodes
2. Implement expression type resolution
3. Annotate SelectorExpr with kind (field/method/getter/setter)
4. Annotate CallExpr with resolved function info

### Phase 4: Go Package Importer

1. Implement GoImporter using go/types
2. Load packages on import statement
3. Extract types, methods, fields from Go packages
4. Handle common stdlib packages (fmt, strings, os, etc.)

### Phase 5: Update Codegen

1. Pass TypeRegistry to Generator
2. Refactor genSelectorExpr to use annotations
3. Refactor genAssignStmt to detect setters
4. Refactor genCallExpr to use resolved info
5. Remove ad-hoc type inference code

### Phase 6: Testing & Cleanup

1. Add comprehensive tests for each phase
2. Run against all examples
3. Fix edge cases
4. Remove deprecated code paths
5. Update documentation

## Benefits

1. **Correct method calls** - `obj.method` automatically becomes `obj.method()`
2. **Working setters** - `obj.prop = value` invokes `obj.setProp(value)`
3. **Full Go interop** - Know exactly what Go packages export
4. **Better errors** - "undefined method 'foo' on type Bar"
5. **Future features** - Type inference, generics, IDE support

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Large refactor | Phased implementation, maintain backward compat |
| Go importer complexity | Start with common packages, lazy loading |
| Performance | Cache aggressively, analyze only what's needed |
| Breaking changes | Comprehensive test suite before starting |

## Open Questions

1. **Caching Go package info** - Should we persist to disk for faster rebuilds?
2. **Incremental analysis** - How to handle when only one file changes?
3. **Generic types** - How to represent Array[T] before T is known?
4. **Error recovery** - Continue analysis after type errors?

## Migration Strategy

Since this is a significant refactor, we need a careful migration path that maintains backward compatibility.

### Step-by-Step Migration

**Step 1: Create types package alongside existing code**
- New `types/` package with Type, Registry, etc.
- Don't modify existing semantic/ or codegen/ yet
- Full unit tests for the new types package

**Step 2: Dual-mode semantic analyzer**
- Add new analysis pass that builds TypeRegistry
- Keep existing analysis for comparison
- Validate that new analysis produces correct results
- AST annotation is additive (new fields, existing code ignores them)

**Step 3: Incremental codegen migration**
- Add TypeRegistry as optional input to Generator
- When available, use annotations for SelectorExpr
- Fall back to existing heuristics when annotations missing
- Migrate one expression type at a time

**Step 4: Go importer integration**
- Start with stdlib packages only (fmt, strings, os, etc.)
- Add caching to avoid repeated parsing
- Handle missing packages gracefully (fall back to heuristics)

**Step 5: Remove legacy code paths**
- Once all examples pass with new system
- Remove old type inference heuristics from codegen
- Remove duplicate logic from semantic analyzer

### Feature Flags

During migration, use feature flags to control behavior:

```go
type AnalyzerOptions struct {
    BuildTypeRegistry bool  // Phase 2+
    AnnotateAST      bool  // Phase 3+
    LoadGoPackages   bool  // Phase 4+
}

type GeneratorOptions struct {
    UseTypeRegistry bool  // Phase 5+
}
```

## Edge Cases & Special Handling

### 1. Method vs Field Ambiguity

When a class has both a field and method with the same name:
```ruby
class Confusing
  @value : Int

  def value -> Int
    @value * 2
  end
end
```

**Resolution:** Methods take precedence. Accessing the raw field requires `@value` syntax.

### 2. Inherited Methods

```ruby
class Parent
  def greet -> String
    "Hello"
  end
end

class Child < Parent
end

c = Child.new
c.greet  # Need to know this is a method from Parent
```

**Resolution:** ClassDef includes inherited methods (flattened or via parent chain lookup).

### 3. Module Inclusion

```ruby
module Greetable
  def greet -> String
    "Hello"
  end
end

class Person
  include Greetable
end

p = Person.new
p.greet  # Method from included module
```

**Resolution:** ClassDef.Methods includes methods from included modules.

### 4. Interface Structural Typing

```ruby
interface Speaker
  def speak -> String
end

class Dog
  def speak -> String
    "Woof"
  end
end

# Dog satisfies Speaker structurally
```

**Resolution:** When checking if ClassDef satisfies InterfaceDef, compare method signatures.

### 5. Go Type Method Receivers

Go methods can have pointer or value receivers:
```go
func (p *Person) SetName(n string)  // pointer receiver
func (p Person) Name() string       // value receiver
```

**Resolution:** GoTypeDef stores receiver type. When calling on `*Person`, both work. When calling on `Person`, only value receiver methods work.

### 6. Chained Method Calls

```ruby
user.profile.name.upcase
```

Each selector needs resolved type to resolve the next:
- `user` -> `*User`
- `user.profile` -> method returns `*Profile`
- `user.profile.name` -> method returns `String`
- `user.profile.name.upcase` -> String method returns `String`

**Resolution:** Type resolution is recursive. Each SelectorExpr resolution uses the resolved type of X.

### 7. Generic Types (Future)

```ruby
arr = Array[Int].new
arr.push(1)
arr.first  # Should know this returns Int
```

**Resolution:** ArrayType has Element field. When Element is known (Int), method return types are specialized.

### 8. Optional Type Methods

```ruby
user = find_user(1)  # returns User?
user.name            # Error: can't call on optional
user&.name           # Safe navigation, returns String?
```

**Resolution:** SelectorExpr on OptionalType is an error unless using safe navigation.

## Standard Library Stubs

For common Go stdlib packages, we can provide pre-built type info instead of runtime reflection:

```go
// types/stdlib/fmt.go
var FmtPackage = &GoPackageInfo{
    Path: "fmt",
    Functions: map[string]*FuncDef{
        "Sprintf": {
            Params:  []Param{{Name: "format", Type: PrimitiveString}, {Name: "args", Type: AnyVariadic}},
            Returns: []Type{PrimitiveString},
        },
        "Printf": {
            Params:  []Param{{Name: "format", Type: PrimitiveString}, {Name: "args", Type: AnyVariadic}},
            Returns: []Type{PrimitiveInt, ErrorType},
        },
        // ...
    },
}
```

Benefits:
- Faster compilation (no reflection needed)
- Works without Go toolchain installed
- Can customize for Rugby conventions

## Appendix: Example Flow

Given this Rugby code:
```ruby
class Point
  @x : Int
  @y : Int

  def initialize(@x, @y)
  end

  def magnitude_squared -> Int
    @x * @x + @y * @y
  end
end

def main
  p = Point.new(3, 4)
  puts p.magnitude_squared  # method call without parens
end
```

**After Phase 2-3 (Semantic Analysis):**

Registry contains:
```
Classes["Point"] = ClassDef{
  Fields: {
    "x": FieldDef{Type: Int},
    "y": FieldDef{Type: Int},
  },
  Methods: {
    "magnitude_squared": MethodDef{Params: [], Returns: [Int]},
  },
}
```

AST annotation on `p.magnitude_squared`:
```
SelectorExpr{
  X: Ident{Name: "p"},
  Sel: "magnitude_squared",
  ResolvedKind: SelectorMethod,
  ResolvedType: Int,
}
```

**After Phase 5 (Codegen):**
```go
p := newPoint(3, 4)
runtime.Puts(p.magnitudeSquared())  // correctly adds ()
```
