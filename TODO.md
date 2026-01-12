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
**Status:** Pending

**Syntax:**
```ruby
arr[-1]     # last element
arr[-2]     # second to last
str[-1]     # last character
```

**Implementation:**
- [ ] Lexer: No changes (already handles negative numbers)
- [ ] AST: No changes (`IndexExpr` already exists)
- [ ] Parser: No changes (already parses `arr[-1]`)
- [ ] CodeGen: Detect negative literal index, emit `runtime.Index(arr, -1)`
- [ ] Runtime: Add `Index[T](slice []T, i int) T` that handles negative indices
- [ ] Runtime: Add `StringIndex(s string, i int) string` for strings
- [ ] Tests: Codegen and runtime tests
- [ ] Docs: Update spec.md Section 12.3 (Array methods)

**Notes:**
- Only static negative literals, or all negative indices at runtime?
- Recommend: runtime check for all index operations for safety

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
**Status:** Pending

**Syntax:**
```ruby
until queue.empty?
  process(queue.pop)
end
```

**Implementation:**
- [ ] Lexer: Add `UNTIL` keyword
- [ ] AST: Add `UntilStmt` node (or reuse `WhileStmt` with `Negated` flag)
- [ ] Parser: Parse `until cond ... end` similar to `while`
- [ ] CodeGen: Emit `for !cond { ... }`
- [ ] Tests: Lexer, parser, codegen tests
- [ ] Docs: Update spec.md Section 5.3 (Imperative Loops)

**Notes:**
- Semantically equivalent to `while !cond`
- Simple addition, low risk

---

### 8. Postfix `while`/`until`
**Priority:** Medium
**Status:** Pending

**Syntax:**
```ruby
puts items.shift while items.any?
process(x) until done?
```

**Implementation:**
- [ ] Lexer: No changes (`while`/`until` already keywords)
- [ ] AST: Extend statement modifier support (already have `if`/`unless`)
- [ ] Parser: Parse `<stmt> while <cond>` and `<stmt> until <cond>`
- [ ] CodeGen: Emit `for cond { stmt }` or `for !cond { stmt }`
- [ ] Tests: Parser, codegen tests
- [ ] Docs: Update spec.md Section 5.5 (Statement Modifiers)

**Notes:**
- Follows same pattern as existing `if`/`unless` modifiers
- `stmt while cond` executes stmt repeatedly while cond is true
- Unlike Ruby, we won't support `begin...end while` (do-while)

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
