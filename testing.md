# Rugby Testing Spec (MVP)

**Goal:** A testing framework that feels like RSpec but compiles to idiomatic Go `testing` package code. Zero reflection, zero magic, full `go test` compatibility.

## 1. Core Philosophy

*   **Compilation:** Rugby tests (`*_test.rg`) compile to standard Go test files (`*_test.go`).
*   **No Runtime Magic:** The DSL is compile-time sugar. It generates `func TestXxx` and `t.Run` calls.
*   **Standard Tooling:** `go test ./...`, `go test -v`, `go test -run`, and IDE integration work out of the box.

## 2. File Organization

*   **Naming:** Files ending in `_test.rg` are treated as test files.
*   **Package:** Tests live in the same package as the code they test (white-box testing).
    *   *Future:* Support `package foo_test` (black-box testing) via explicit directive.
*   **Compilation Target:**
    *   `math_test.rg` → `math_test.go`
    *   These files are only compiled during `rugby test` (which invokes `go test`).

## 3. Syntax & Lowering

### 3.1 Suites (`describe`)

`describe` blocks group related tests. They compile to nested `t.Run` calls.

**Rugby:**
```ruby
describe "String#to_i?" do
  it "parses valid ints" do |t|
    # test body
  end
end
```

**Go Output:**
```go
func TestStringToI(t *testing.T) {
  t.Run("parses valid ints", func(t *testing.T) {
    // test body
  })
}
```

**Naming Rules:**
*   The top-level `describe` generates the `func TestXxx` name.
*   Names are sanitized: `String#to_i?` → `TestStringToI`.
*   **Collision Prevention:** If multiple top-level blocks map to the same name, a stable hash suffix (derived from file/line) is appended: `TestStringToI_A1B2`.

### 3.2 Cases (`it` / `test`)

**Nested Test (`it`):**
Used inside `describe`. Maps to `t.Run`.

```ruby
it "handles zeros" do |t|
  # ...
end
```

**Standalone Test (`test`):**
Used at the top level for simple tests without grouping.

```ruby
test "math/add" do |t|
  assert.equal(t, 2, 1 + 1)
end
```
*Compiles to:* `func TestMathAdd(t *testing.T) { ... }`

**Nested `describe` (Example):**

```ruby
describe "User" do
  describe "Validation" do
    it "requires name" do |t| ... end
  end
end
```

*Compiles to:*
```go
func TestUser(t *testing.T) {
    t.Run("Validation", func(t *testing.T) {
        t.Run("requires name", func(t *testing.T) { ... })
    })
}
```

### 3.3 Table Tests (`table`)

Rugby supports table-driven tests with a clean DSL that compiles to a standard Go table loop.

**Rugby:**
```ruby
table "String#to_i?" do |tt|
  # 1. Define cases
  # NOTE: The compiler infers types strictly from the FIRST case.
  # If the first case has ambiguous types (e.g. nil, 0), use explicit casting
  # or ensure the first case represents the full structure.
  tt.case "ok",   "12",   12, true
  tt.case "zero", "0",    0,  true
  tt.case "bad",  "nope", 0,  false

  # 2. Define execution logic
  tt.run do |t, input, expected, expected_ok|
    got, got_ok = input.to_i?
    assert.equal(t, expected_ok, got_ok)
    if got_ok
      assert.equal(t, expected, got)
    end
  end
end
```

**Go Output (Conceptual):**
```go
func TestStringToI_Table(t *testing.T) {
    cases := []struct {
        name     string
        input    string
        expected int
        ok       bool
    }{
        {"ok", "12", 12, true},
        {"zero", "0", 0, true},
        // ...
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            // Unpack struct fields to arguments
            // Call body
        })
    }
}
```

## 4. Assertions (`rugby/test`)

Assertions are implemented as a standard Go helper package. They wrap `t.Errorf` (assert) or `t.Fatalf` (require).

**Import:** The `rugby/test` package is automatically available as `assert` and `require` in `*_test.rg` files. No manual import required.

**Argument Order:** `(t, expected, actual)` matches standard Go (testify) and Ruby (minitest) conventions.

**API:**
*   `assert.equal(t, expected, actual)`
*   `assert.not_equal(t, expected, actual)`
*   `assert.true(t, bool)` / `assert.false(t, bool)`
*   `assert.nil(t, val)` / `assert.not_nil(t, val)`
*   `assert.error(t, err)` / `assert.no_error(t, err)`

**Fatal Variants:**
*   `require.equal(t, ...)` (stops test execution immediately)

## 5. Lifecycle Hooks (`before` / `after`)

Hooks allow you to set up and tear down state for your tests without global mutable variables.

### 5.1 Context Passing

`before` hooks return a value (the context) which is passed to both `it` and `after` blocks.
*   **Return Type:** Inferred by the compiler from the block's return value. (Explicit arrow syntax `-> T` for blocks is not yet supported).
*   **Teardown (Option A):** Use an `after` block. It receives the same context object.
*   **Teardown (Option B):** Use `t.cleanup { ... }` inside the `before` block (Go-style).

```ruby
describe "Database" do
  before do |t|
    db = open_test_db()
    db # Returns context
  end

  after do |t, db|
    db.close()
  end

  it "saves records" do |t, db|
    assert.not_nil(t, db)
  end
end
```

### 5.2 Inline Cleanup (`t.cleanup`)

For more complex setup where multiple resources are created, `t.cleanup` is often cleaner as it keeps resource creation and destruction logically adjacent.

```ruby
before do |t|
  db = open_db()
  t.cleanup { db.close() }
  
  file = open_log()
  t.cleanup { file.close() }
  
  [db, file] # Return context
end
```

**MVP Restriction (No Cascading):**
To keep the compiler predictable and avoid breaking changes later:
*   `before` and `after` blocks only apply to `it` blocks *directly* within the same `describe` scope.
*   Nested `describe` blocks do **not** inherit or cascade context from parent blocks in the MVP.

**Compilation:**
1.  `before` block becomes a local function `setup(t)`.
2.  `after` block becomes a local function `teardown(t, ctx)`.
3.  Inside `t.Run`:
    *   `ctx := setup(t)`
    *   `t.Cleanup(func() { teardown(t, ctx) })`
    *   `itBody(t, ctx)`

## 6. Concurrency & Control

Standard Go testing features are available directly on `t`.

*   `t.parallel()` → `t.Parallel()`
*   `t.skip("reason")` → `t.Skip("reason")`

## 7. MVP Scope

### 7.1 Compiler Support
*   [ ] Recognizes `*_test.rg` files.
*   [ ] Lowers `describe` / `it` / `test` to `func Test...` / `t.Run`.
*   [ ] Generates deterministic names for test functions.

### 7.2 Runtime Library (`rugby/test`)
*   [ ] Implement basic `assert` functions (`Equal`, `True`, `Nil`, `NoError`).
*   [ ] Implement `require` variants.

### 7.3 Tooling
*   [ ] `rugby test` command acts as a passthrough to `go test`.
