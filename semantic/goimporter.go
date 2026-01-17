package semantic

import (
	"go/importer"
	"go/types"
	"sync"
)

// GoImporter loads and caches Go package type information.
// It uses go/types to introspect Go packages at compile time,
// enabling Rugby to understand the types returned by Go functions and methods.
type GoImporter struct {
	cache    map[string]*GoPackageInfo
	importer types.Importer
	mu       sync.RWMutex
}

// GoPackageInfo stores extracted type information for a Go package.
type GoPackageInfo struct {
	Path      string
	Types     map[string]*GoTypeDef // Named types (structs, interfaces)
	Functions map[string]*GoFuncDef // Package-level functions
}

// GoTypeDef describes a Go type (struct or interface).
type GoTypeDef struct {
	Name    string
	Methods map[string]*GoMethodDef // Methods on this type
	Fields  map[string]*GoFieldDef  // Fields (for structs)
}

// GoMethodDef describes a method on a Go type.
type GoMethodDef struct {
	Name       string
	Params     []*Type // Parameter types
	Returns    []*Type // Return types
	IsVariadic bool
}

// GoFuncDef describes a package-level Go function.
type GoFuncDef struct {
	Name       string
	Params     []*Type
	Returns    []*Type
	IsVariadic bool
}

// GoFieldDef describes a field on a Go struct.
type GoFieldDef struct {
	Name string
	Type *Type
}

// NewGoImporter creates a new Go package importer.
func NewGoImporter() *GoImporter {
	return &GoImporter{
		cache:    make(map[string]*GoPackageInfo),
		importer: importer.Default(),
	}
}

// Import loads a Go package and extracts its type information.
func (gi *GoImporter) Import(path string) (*GoPackageInfo, error) {
	gi.mu.RLock()
	if cached, ok := gi.cache[path]; ok {
		gi.mu.RUnlock()
		return cached, nil
	}
	gi.mu.RUnlock()

	// Import outside the lock
	pkg, err := gi.importer.Import(path)
	if err != nil {
		return nil, err
	}
	info := gi.extractPackageInfo(pkg)

	gi.mu.Lock()
	defer gi.mu.Unlock()
	// Double-check after acquiring write lock to avoid duplicate imports
	if cached, ok := gi.cache[path]; ok {
		return cached, nil
	}
	gi.cache[path] = info
	return info, nil
}

// extractPackageInfo converts go/types info to our representation.
func (gi *GoImporter) extractPackageInfo(pkg *types.Package) *GoPackageInfo {
	info := &GoPackageInfo{
		Path:      pkg.Path(),
		Types:     make(map[string]*GoTypeDef),
		Functions: make(map[string]*GoFuncDef),
	}

	scope := pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		switch o := obj.(type) {
		case *types.TypeName:
			if o.Exported() {
				info.Types[name] = gi.extractTypeDef(o)
			}
		case *types.Func:
			if o.Exported() {
				info.Functions[name] = gi.extractFuncDef(o)
			}
		}
	}

	return info
}

// extractTypeDef extracts full type info including methods.
func (gi *GoImporter) extractTypeDef(tn *types.TypeName) *GoTypeDef {
	def := &GoTypeDef{
		Name:    tn.Name(),
		Methods: make(map[string]*GoMethodDef),
		Fields:  make(map[string]*GoFieldDef),
	}

	typ := tn.Type()

	// Get methods (including pointer receiver methods)
	mset := types.NewMethodSet(types.NewPointer(typ))
	for i := 0; i < mset.Len(); i++ {
		sel := mset.At(i)
		fn := sel.Obj().(*types.Func)
		if fn.Exported() {
			def.Methods[fn.Name()] = gi.extractMethodDef(fn)
		}
	}

	// Also get methods on value receiver (for non-pointer types)
	msetVal := types.NewMethodSet(typ)
	for i := 0; i < msetVal.Len(); i++ {
		sel := msetVal.At(i)
		fn := sel.Obj().(*types.Func)
		if fn.Exported() {
			if _, exists := def.Methods[fn.Name()]; !exists {
				def.Methods[fn.Name()] = gi.extractMethodDef(fn)
			}
		}
	}

	// Get fields for struct types
	if st, ok := typ.Underlying().(*types.Struct); ok {
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			if f.Exported() {
				def.Fields[f.Name()] = &GoFieldDef{
					Name: f.Name(),
					Type: gi.convertGoType(f.Type()),
				}
			}
		}
	}

	return def
}

// extractFuncDef extracts function signature information.
func (gi *GoImporter) extractFuncDef(fn *types.Func) *GoFuncDef {
	sig := fn.Type().(*types.Signature)
	return &GoFuncDef{
		Name:       fn.Name(),
		Params:     gi.extractParams(sig.Params()),
		Returns:    gi.extractResults(sig.Results()),
		IsVariadic: sig.Variadic(),
	}
}

// extractMethodDef extracts method signature information.
func (gi *GoImporter) extractMethodDef(fn *types.Func) *GoMethodDef {
	sig := fn.Type().(*types.Signature)
	return &GoMethodDef{
		Name:       fn.Name(),
		Params:     gi.extractParams(sig.Params()),
		Returns:    gi.extractResults(sig.Results()),
		IsVariadic: sig.Variadic(),
	}
}

// extractParams converts go/types parameters to our Type slice.
func (gi *GoImporter) extractParams(tuple *types.Tuple) []*Type {
	if tuple == nil {
		return nil
	}
	params := make([]*Type, tuple.Len())
	for i := 0; i < tuple.Len(); i++ {
		params[i] = gi.convertGoType(tuple.At(i).Type())
	}
	return params
}

// extractResults converts go/types results to our Type slice.
func (gi *GoImporter) extractResults(tuple *types.Tuple) []*Type {
	if tuple == nil {
		return nil
	}
	results := make([]*Type, tuple.Len())
	for i := 0; i < tuple.Len(); i++ {
		results[i] = gi.convertGoType(tuple.At(i).Type())
	}
	return results
}

// convertGoType converts a go/types.Type to our Type representation.
func (gi *GoImporter) convertGoType(t types.Type) *Type {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return TypeBoolVal
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
			return TypeIntVal
		case types.Float32, types.Float64:
			return TypeFloatVal
		case types.String:
			return TypeStringVal
		case types.UntypedBool:
			return TypeBoolVal
		case types.UntypedInt:
			return TypeIntVal
		case types.UntypedFloat:
			return TypeFloatVal
		case types.UntypedString:
			return TypeStringVal
			// Note: types.Byte is alias for Uint8, types.Rune is alias for Int32
			// They're already handled above in the int cases
		}
	case *types.Slice:
		elem := gi.convertGoType(t.Elem())
		return NewArrayType(elem)
	case *types.Array:
		elem := gi.convertGoType(t.Elem())
		return NewArrayType(elem)
	case *types.Map:
		key := gi.convertGoType(t.Key())
		val := gi.convertGoType(t.Elem())
		return NewMapType(key, val)
	case *types.Pointer:
		// For pointers, we track the underlying type but mark it as a pointer
		inner := gi.convertGoType(t.Elem())
		// Create a new type that represents the pointer
		return &Type{
			Kind:       inner.Kind,
			Name:       inner.Name,
			Elem:       inner.Elem,
			KeyType:    inner.KeyType,
			ValueType:  inner.ValueType,
			IsPointer:  true,
			GoPackage:  inner.GoPackage,
			GoTypeName: inner.GoTypeName,
		}
	case *types.Named:
		pkg := t.Obj().Pkg()
		pkgPath := ""
		if pkg != nil {
			pkgPath = pkg.Path()
		}
		// Check for the builtin error type (no package, name is "error")
		if pkg == nil && t.Obj().Name() == "error" {
			return TypeErrorVal
		}
		return &Type{
			Kind:       TypeGoType,
			Name:       t.Obj().Name(),
			GoPackage:  pkgPath,
			GoTypeName: t.Obj().Name(),
		}
	case *types.Interface:
		// Check if this is the error interface
		if t.NumMethods() == 1 {
			method := t.Method(0)
			if method.Name() == "Error" {
				sig := method.Type().(*types.Signature)
				if sig.Params().Len() == 0 && sig.Results().Len() == 1 {
					if basic, ok := sig.Results().At(0).Type().(*types.Basic); ok && basic.Kind() == types.String {
						return TypeErrorVal
					}
				}
			}
		}
		// For interface{}/any, return TypeAny
		if t.Empty() {
			return TypeAnyVal
		}
		// For other interfaces, return TypeAny (we can't fully represent them)
		return TypeAnyVal
	case *types.TypeParam:
		// For generic type parameters, return TypeAny as we can't know
		// the concrete type at import time
		return TypeAnyVal
	case *types.Chan:
		elem := gi.convertGoType(t.Elem())
		return NewChanType(elem)
	case *types.Signature:
		// Function types
		params := gi.extractParams(t.Params())
		returns := gi.extractResults(t.Results())
		return NewFuncType(params, returns)
	}
	return TypeUnknownVal
}

// LookupFunction looks up a function in a Go package.
func (gi *GoImporter) LookupFunction(pkgPath, funcName string) *GoFuncDef {
	gi.mu.RLock()
	pkg, ok := gi.cache[pkgPath]
	gi.mu.RUnlock()
	if !ok {
		return nil
	}
	return pkg.Functions[funcName]
}

// LookupType looks up a type in a Go package.
func (gi *GoImporter) LookupType(pkgPath, typeName string) *GoTypeDef {
	gi.mu.RLock()
	pkg, ok := gi.cache[pkgPath]
	gi.mu.RUnlock()
	if !ok {
		return nil
	}
	return pkg.Types[typeName]
}

// LookupMethod looks up a method on a Go type.
func (gi *GoImporter) LookupMethod(pkgPath, typeName, methodName string) *GoMethodDef {
	typeDef := gi.LookupType(pkgPath, typeName)
	if typeDef == nil {
		return nil
	}
	return typeDef.Methods[methodName]
}
