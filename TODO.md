# Code Cleanup and Optimization

Refactoring work to make the compiler simpler, more elegant, and maintainable.

## Completed

### Token Package
- [x] Fixed hardcoded keyword count in tests - now verifies all keywords via `LookupIdent()` without brittle counts

### Lexer Package
- [x] Converted recursive comment handling to loop - prevents potential stack overflow on files with many consecutive comment lines

### Parser Package
- [x] Extracted `withLookahead()` helper - eliminates ~100 lines of duplicated save/restore boilerplate across 6 lookahead methods
- [x] Added `isCompoundAssignToken()` and `compoundOpSymbol()` helpers - consolidated 5 duplicate compound assignment switch blocks

### Codegen Package
- [x] Split `expressions.go` (5,936 → 2,003 lines) into focused files:
  - `blocks.go` (2,200 lines) - block iteration methods (each, map, select, etc.)
  - `calls.go` (1,765 lines) - function/method calls and selectors
- [x] Split `statements.go` (2,663 → 683 lines) into focused files:
  - `assignments.go` (1,128 lines) - all assignment statement types
  - `control_flow.go` (887 lines) - if, case, while, for loops
- [x] Extracted `ClassContext` and `MethodContext` structs - replaced 8 scattered context fields with cohesive state objects:
  - `ClassContext`: bundles Name, OriginalName, Embeds, TypeParamClause, TypeParamNames, InterfaceMethods, AccessorFields
  - `MethodContext`: bundles Name, IsPub, ReturnTypes, TypeParams
  - Added helper methods (currentClass(), currentMethod(), etc.) for backward-compatible read access
- [x] Consolidated `goMethodName` transformation logic - extracted 3 helper functions:
  - `goMethodName(name, isPub, isPrivate, isInterfaceMethod)` - handles all method naming variations
  - `goSetterName(fieldName, isPub, isPrivate)` - handles setter method naming
  - `goFuncName(name, isPub)` - handles function naming
- [x] Consolidated 8 pre-pass loops into 2 loops - single type switch over definitions collects all metadata in one traversal

---

## Pending

### Runtime Package (do last)
- [ ] Delete redundant type-specific functions - `MapEachString`, `MapKeysString`, etc. can be replaced by generics
- [ ] Consolidate coalesce and optional helpers - `CoalesceInt`, `CoalesceString`, etc. should be unified
- [ ] Consider splitting `array.go` (~600 lines) into smaller, focused files

### AST Package (low priority)
- [ ] Consider consolidating assignment types - `AssignStmt`, `IndexAssignStmt`, `SelectorAssignStmt` have overlapping structures
