package semantic

import "testing"

func TestScopeDefineAndLookup(t *testing.T) {
	global := NewGlobalScope()

	// Define a variable in global scope
	x := NewVariable("x", TypeIntVal)
	if err := global.Define(x); err != nil {
		t.Fatalf("Define(x) failed: %v", err)
	}

	// Lookup should find it
	found := global.Lookup("x")
	if found == nil {
		t.Fatal("Lookup(x) returned nil")
	}
	if found.Name != "x" {
		t.Errorf("Lookup(x).Name = %q, want %q", found.Name, "x")
	}

	// Lookup should not find undefined symbol
	if global.Lookup("y") != nil {
		t.Error("Lookup(y) should return nil")
	}
}

func TestScopeRedefinition(t *testing.T) {
	global := NewGlobalScope()

	x1 := NewVariable("x", TypeIntVal)
	x1.Line = 1
	if err := global.Define(x1); err != nil {
		t.Fatalf("Define(x) failed: %v", err)
	}

	// Redefining in the same scope should error
	x2 := NewVariable("x", TypeStringVal)
	x2.Line = 5
	err := global.Define(x2)
	if err == nil {
		t.Fatal("Define(x) should fail on redefinition")
	}

	redefinErr, ok := err.(*RedefinitionError)
	if !ok {
		t.Fatalf("expected RedefinitionError, got %T", err)
	}
	if redefinErr.Name != "x" {
		t.Errorf("RedefinitionError.Name = %q, want %q", redefinErr.Name, "x")
	}
}

func TestScopeNesting(t *testing.T) {
	global := NewGlobalScope()
	x := NewVariable("x", TypeIntVal)
	_ = global.Define(x)

	// Create nested function scope
	fnScope := NewScope(ScopeFunction, global)

	// Should find x from parent
	if fnScope.Lookup("x") == nil {
		t.Error("nested scope should find x from parent")
	}

	// Define y in function scope
	y := NewVariable("y", TypeStringVal)
	_ = fnScope.Define(y)

	// Function scope should find y
	if fnScope.Lookup("y") == nil {
		t.Error("function scope should find y")
	}

	// Global scope should NOT find y
	if global.Lookup("y") != nil {
		t.Error("global scope should not find y defined in child scope")
	}
}

func TestScopeShadowing(t *testing.T) {
	global := NewGlobalScope()
	xGlobal := NewVariable("x", TypeIntVal)
	_ = global.Define(xGlobal)

	// Create nested scope that shadows x
	block := NewScope(ScopeBlock, global)
	xLocal := NewVariable("x", TypeStringVal)
	if err := block.DefineOrShadow(xLocal); err != nil {
		t.Fatalf("DefineOrShadow should allow shadowing: %v", err)
	}

	// Block should find local x (String type)
	found := block.Lookup("x")
	if found == nil {
		t.Fatal("block should find x")
	}
	if !found.Type.Equals(TypeStringVal) {
		t.Errorf("block.Lookup(x).Type = %v, want String", found.Type)
	}

	// Global should still have Int type
	foundGlobal := global.Lookup("x")
	if !foundGlobal.Type.Equals(TypeIntVal) {
		t.Errorf("global.Lookup(x).Type = %v, want Int", foundGlobal.Type)
	}
}

func TestScopeLookupLocal(t *testing.T) {
	global := NewGlobalScope()
	x := NewVariable("x", TypeIntVal)
	_ = global.Define(x)

	block := NewScope(ScopeBlock, global)
	y := NewVariable("y", TypeStringVal)
	_ = block.Define(y)

	// LookupLocal should only find symbols in the current scope
	if block.LookupLocal("x") != nil {
		t.Error("LookupLocal should not find x from parent")
	}
	if block.LookupLocal("y") == nil {
		t.Error("LookupLocal should find y in current scope")
	}
}

func TestScopeClassContext(t *testing.T) {
	global := NewGlobalScope()

	classScope := NewScope(ScopeClass, global)
	classScope.ClassName = "User"

	methodScope := NewScope(ScopeMethod, classScope)

	// Method scope should inherit class name
	if methodScope.ClassName != "User" {
		t.Errorf("ClassName = %q, want %q", methodScope.ClassName, "User")
	}
	if !methodScope.IsInsideClass() {
		t.Error("IsInsideClass() should return true")
	}

	blockScope := NewScope(ScopeBlock, methodScope)
	if blockScope.ClassName != "User" {
		t.Error("Block inside method should inherit class name")
	}
}

func TestScopeFunctionScope(t *testing.T) {
	global := NewGlobalScope()
	fnScope := NewScope(ScopeFunction, global)
	fnScope.ReturnTypes = []*Type{TypeIntVal}

	blockScope := NewScope(ScopeBlock, fnScope)
	innerBlock := NewScope(ScopeBlock, blockScope)

	// FunctionScope should find the enclosing function
	found := innerBlock.FunctionScope()
	if found == nil {
		t.Fatal("FunctionScope() returned nil")
	}
	if found != fnScope {
		t.Error("FunctionScope() should return the enclosing function scope")
	}
	if len(found.ReturnTypes) != 1 || !found.ReturnTypes[0].Equals(TypeIntVal) {
		t.Error("Function scope should have return types")
	}
}

func TestScopeIsInsideLoop(t *testing.T) {
	global := NewGlobalScope()

	if global.IsInsideLoop() {
		t.Error("Global scope should not be inside loop")
	}

	fnScope := NewScope(ScopeFunction, global)
	if fnScope.IsInsideLoop() {
		t.Error("Function scope should not be inside loop")
	}

	// Regular block (like if-then) is NOT a loop
	blockScope := NewScope(ScopeBlock, fnScope)
	if blockScope.IsInsideLoop() {
		t.Error("Block scope (if-then) should not be considered inside loop")
	}

	// ScopeLoop is a loop
	loopScope := NewScope(ScopeLoop, fnScope)
	if !loopScope.IsInsideLoop() {
		t.Error("Loop scope should be inside loop")
	}

	// Block inside a loop IS inside a loop
	innerBlock := NewScope(ScopeBlock, loopScope)
	if !innerBlock.IsInsideLoop() {
		t.Error("Block inside loop should be inside loop")
	}

	// But a new function inside a loop is NOT inside a loop
	innerFn := NewScope(ScopeFunction, loopScope)
	if innerFn.IsInsideLoop() {
		t.Error("Function inside loop should NOT be considered inside loop")
	}
}

func TestScopeUpdate(t *testing.T) {
	global := NewGlobalScope()
	x := NewVariable("x", TypeUnknownVal)
	_ = global.Define(x)

	block := NewScope(ScopeBlock, global)

	// Update from child scope should update the parent
	if !block.Update("x", TypeIntVal) {
		t.Error("Update should return true")
	}

	found := global.Lookup("x")
	if !found.Type.Equals(TypeIntVal) {
		t.Errorf("After update, type = %v, want Int", found.Type)
	}
}
