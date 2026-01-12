# Rugby Language Enhancements TODO

This document tracks the implementation of Ruby-inspired syntax shorthands to make Rugby more expressive.

## Features to Implement

### 1. Symbol-to-Proc (`&:method`)
**Priority:** High
**Status:** Pending

**Syntax:**
```ruby
users.map(&:name)           # → users.map { |x| x.name }
items.select(&:valid?)      # → items.select { |x| x.valid? }
numbers.reduce(&:+)         # → numbers.reduce { |a, b| a + b }
```

**Implementation:**
- [ ] Lexer: Tokenize `&:` as a unit or `&` followed by symbol
- [ ] AST: Add `SymbolProcExpr` node or reuse existing nodes
- [ ] Parser: Parse `&:method` in argument position, desugar to block
- [ ] CodeGen: Generate the equivalent closure
- [ ] Tests: Lexer, parser, codegen tests
- [ ] Docs: Update spec.md Section 5.4 (Functional Blocks)

**Notes:**
- For binary operators like `&:+`, need special handling
- Should work with any method that takes a block

---

### 2. Array Append (`<<`)
**Priority:** High
**Status:** Pending

**Syntax:**
```ruby
items << new_item           # append single item
items << a << b << c        # chainable
```

**Implementation:**
- [ ] Lexer: `<<` already exists for channels, reuse token
- [ ] AST: Reuse `BinaryExpr` with `<<` operator
- [ ] Parser: Already parses `<<`, no changes needed
- [ ] CodeGen: Type-check LHS; if array, emit `append()`; if chan, emit `<-`
- [ ] Runtime: May need `Append[T]` helper for chaining
- [ ] Tests: Parser (already works), codegen tests for array vs chan
- [ ] Docs: Update spec.md Section 4.2 (Composite types) and Section 12.3 (Array methods)

**Notes:**
- Chaining `a << b << c` requires `<<` to return the array
- Go's `append` returns new slice; need to handle mutation semantics

---

### 3. Negative Indexing
**Priority:** High
**Status:** ✅ Complete

**Syntax:**
```ruby
arr[-1]     # last element
arr[-2]     # second to last
str[-1]     # last character
```

**Implementation:**
- [x] Lexer: No changes (already handles negative numbers)
- [x] AST: No changes (`IndexExpr` already exists)
- [x] Parser: No changes (already parses `arr[-1]`)
- [x] CodeGen: Detect negative/variable index, emit `runtime.AtIndex(arr, i)`
- [x] Runtime: Added `Index[T]`, `IndexOpt[T]`, `AtIndex`, `AtIndexOpt`
- [x] Runtime: Added `StringIndex`, `StringIndexOpt` for strings
- [x] Tests: Codegen and runtime tests
- [x] Docs: Updated spec.md Section 4.2 (Array and String Indexing)

**Notes:**
- Non-negative literal indices use native Go syntax for efficiency
- Variable and negative indices use `runtime.AtIndex` for safety
- `AtIndex` handles both strings (rune-based) and slices via reflection

---

### 4. Ternary Operator (`? :`)
**Priority:** High
**Status:** Pending

**Syntax:**
```ruby
status = valid? ? "ok" : "error"
x = a > b ? a : b
```

**Implementation:**
- [ ] Lexer: `?` exists for method names; need context-aware handling or new token
- [ ] AST: Add `TernaryExpr` node with `Condition`, `Then`, `Else`
- [ ] Parser: Parse ternary with correct precedence (very low, above assignment)
- [ ] CodeGen: Emit Go inline if-else (may need temp var for expression context)
- [ ] Tests: Lexer, parser, codegen tests
- [ ] Docs: Update spec.md Section 5.2 (Conditionals)

**Notes:**
- Precedence must be lower than comparison but handle chaining
- `a ? b : c ? d : e` should parse as `a ? b : (c ? d : e)`
- Conflict with `?` suffix on methods needs careful lexer handling

---

### 5. Range Slicing
**Priority:** High
**Status:** Pending

**Syntax:**
```ruby
arr[1..3]     # elements at indices 1, 2, 3 (inclusive)
arr[1...3]    # elements at indices 1, 2 (exclusive end)
arr[2..]      # from index 2 to end
arr[..3]      # from start to index 3
str[0..4]     # substring
```

**Implementation:**
- [ ] Lexer: No changes (ranges already tokenized)
- [ ] AST: Extend `IndexExpr` to allow `Index` to be a `RangeLit`
- [ ] Parser: Allow range expression inside `[]`
- [ ] CodeGen: Emit `runtime.Slice(arr, start, end, exclusive)`
- [ ] Runtime: Add `Slice[T](slice []T, start, end int, exclusive bool) []T`
- [ ] Runtime: Add `StringSlice(s string, start, end int, exclusive bool) string`
- [ ] Runtime: Handle nil start/end for open-ended ranges
- [ ] Tests: Parser, codegen, runtime tests
- [ ] Docs: Update spec.md Section 4.2.1 (Range Type) and Section 12.3

**Notes:**
- Open-ended ranges (`2..`, `..3`) need AST support (nil Start/End in RangeLit)
- Need to handle both arrays and strings

---

### 6. Heredocs
**Priority:** Medium
**Status:** Pending

**Syntax:**
```ruby
sql = <<~SQL
  SELECT *
  FROM users
  WHERE active = true
SQL

plain = <<EOF
No dedent here
EOF
```

**Implementation:**
- [ ] Lexer: Recognize `<<~IDENT` and `<<IDENT`, read until terminator
- [ ] Lexer: Track heredoc state, handle dedenting for `<<~`
- [ ] AST: Reuse `StringLit` (heredoc is just a string)
- [ ] Parser: No changes needed if lexer produces STRING token
- [ ] CodeGen: No changes (it's a string)
- [ ] Tests: Lexer tests for various heredoc forms
- [ ] Docs: Update spec.md Section 3.3 (Strings)

**Notes:**
- `<<~` strips leading whitespace based on least-indented line
- `<<` preserves whitespace exactly
- Terminator must be at start of line (or after only whitespace for `<<~`)
- Support interpolation: `<<~SQL` vs `<<~'SQL'` (quoted = no interpolation)

---

### 7. `until` Loop
**Priority:** Medium
**Status:** ✅ Complete

**Syntax:**
```ruby
until queue.empty?
  process(queue.pop)
end
```

**Implementation:**
- [x] Lexer: Added `UNTIL` keyword
- [x] AST: Added `UntilStmt` node
- [x] Parser: Added `parseUntilStmt` similar to `while`
- [x] CodeGen: Emits `for !cond { ... }` with smart parenthesization
- [x] Tests: Lexer, parser, codegen tests
- [x] Docs: Updated spec.md Section 5.3 (Imperative Loops)

**Notes:**
- Semantically equivalent to `while !cond`
- Supports break/next like while loop

---

### 8. Postfix `while`/`until`
**Priority:** Medium
**Status:** ✅ Complete

**Syntax:**
```ruby
puts items.shift while items.any?
process(x) until done?
```

**Implementation:**
- [x] Lexer: No changes (`while`/`until` already keywords)
- [x] AST: Reuses `WhileStmt` and `UntilStmt` nodes
- [x] Parser: Added `parseLoopModifier` to handle postfix forms
- [x] CodeGen: Reuses existing `genWhileStmt` and `genUntilStmt`
- [x] Tests: Parser, codegen tests
- [x] Docs: Updated spec.md Section 5.6 (Loop Modifiers)

**Notes:**
- Follows same pattern as existing `if`/`unless` modifiers
- `stmt while cond` executes stmt repeatedly while cond is true
- Cannot be combined with if/unless modifiers on same statement

---

## Implementation Order

1. **`until` loop** - Simplest, good warmup
2. **Postfix `while`/`until`** - Builds on until, extends existing modifier pattern
3. **Negative indexing** - Self-contained, high value
4. **Array append `<<`** - Builds on existing operator
5. **Ternary operator** - Moderate complexity, high value
6. **Range slicing** - Builds on existing range support
7. **Symbol-to-Proc `&:method`** - Most complex parser change
8. **Heredocs** - Mostly lexer work, can be done anytime

---

## Testing Strategy

For each feature:
1. **Lexer tests** (if new tokens): `lexer/lexer_test.go`
2. **Parser tests**: `parser/parser_test.go`
3. **CodeGen tests**: `codegen/codegen_test.go`
4. **Runtime tests** (if new functions): `runtime/*_test.go`
5. **Integration test**: Add example in `examples/`

---

## Documentation Updates

For each feature, update:
1. `spec.md` - Language specification
2. `CLAUDE.md` - If it affects architecture understanding
3. Code comments - For non-obvious implementation details
