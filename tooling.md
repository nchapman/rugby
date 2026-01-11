# Rugby Tooling & CLI Spec

**Goal:** A zero-config, "batteries-included" CLI that creates a joyous developer experience.

The `rugby` command manages the entire lifecycle: scaffolding, dependencies, building, testing, and formatting.

---

## 1. File Structure Strategy

Rugby keeps the user's source directory clean. All intermediate artifacts live in a hidden directory.

### 1.1 The `.rugby` Directory

Located at the project root.

```text
my-project/
├── main.rg
├── src/
├── go.mod          # Source of truth for deps
├── .rugby/         # MANAGED BY CLI
│   ├── bin/        # Compiled binaries
│   ├── gen/        # Generated Go source tree (mirrors project)
│   ├── cache/      # Incremental compilation artifacts
│   └── debug/      # Logs and error reports
└── .gitignore      # Should include .rugby/
```

### 1.2 Compilation Flow

1.  **Parse:** Rugby parses `.rg` files in the user's directory.
2.  **Gen:** Generates equivalent `.go` files into `.rugby/gen/`.
    *   *Crucial:* Emits `//line original.rg:15` directives in Go files. This ensures stack traces and compiler errors point to the **Rugby** source, not the generated Go code.
3.  **Build:** Runs `go build` inside `.rugby/gen/` (using the project's root `go.mod`).
4.  **Exec:** Runs the binary from `.rugby/bin/`.

---

## 2. Core Commands

### 2.1 Development (`rugby run`)

```bash
rugby run [entrypoint.rg]
```

*   **Behavior:** Transpiles, builds, and executes immediately.
*   **Caching:** Checks file hashes. If `.rg` hasn't changed, skips transpilation.
*   **Arguments:** Pass args after `--` (e.g., `rugby run main.rg -- --port 8080`).

### 2.2 Watch Mode (`rugby watch`)

```bash
rugby watch [entrypoint.rg]
```

*   **Behavior:** Same as `run`, but keeps the process alive.
*   **Trigger:** Re-runs on any file save (`.rg` or `.go` files).
*   **UX:** Clears the console between runs for a fresh view.

### 2.3 Production Build (`rugby build`)

```bash
rugby build -o myapp
```

*   **Behavior:** Produces a highly optimized, standalone binary in the current directory.
*   **Optimization:** Stripes debug symbols (optional), enables aggressive Go compiler optimizations.
*   **Artifact:** The result is a standard executable, dependent only on libc (or static if configured).

---

## 3. Project Management

### 3.1 Initialization (`rugby init`)

```bash
rugby init [name]
```

*   Creates `rugby.toml` (project config) and `go.mod`.
*   Scaffolds a "Hello World" `main.rg`.
*   Sets up `.gitignore` to exclude `.rugby/`.

### 3.2 Dependencies (`rugby add/install`)

Rugby leverages Go modules but provides a friendly wrapper.

```bash
rugby add github.com/gin-gonic/gin
```

*   **Behavior:** Runs `go get ...` and updates `go.mod`.
*   **Auto-Import:** Optionally adds an import alias to your entry file if requested.

```bash
rugby install
```

*   **Behavior:** Ensures all dependencies in `go.mod` are downloaded (wrapper for `go mod download`).

### 3.3 Formatting (`rugby format`)

```bash
rugby format
```

*   **Philosophy:** One standard style (like `gofmt` or `prettier`). No arguments.
*   **Action:** Rewrites `.rg` files to standard indentation (2 spaces), spacing, and block styles.

---

## 4. Testing (`rugby test`)

```bash
rugby test [path]
```

*   **Discovery:** Finds files ending in `_test.rg`.
*   **Structure:**
    ```ruby
    # math_test.rg
    test "addition" do
      assert_equal(2 + 2, 4)
    end
    ```
*   **Execution:**
    1.  Transpiles `*_test.rg` to `*_test.go`.
    2.  Runs `go test` within the `.rugby/gen` context.
    3.  Pipes output through a prettifier to show Rugby syntax highlighting in failures.

---

## 5. Cleaning (`rugby clean`)

```bash
rugby clean
```

*   **Behavior:** Wipes the `.rugby/` directory. Useful if the cache gets corrupted or to force a fresh build.

---

## 6. Error Handling & Diagnostics

The CLI must bridge the gap between Go and Rugby.

### 6.1 Compiler Errors
If the **Go** compiler fails (e.g., type mismatch):
1.  Capture the Go error output.
2.  Parse the line number (which points to Rugby source via `//line`).
3.  Display a code snippet of the **Rugby** file.
4.  Hide the underlying Go error specifics unless `--verbose` is used, translating them to Rugby concepts (e.g., "cannot use Int as String" instead of "cannot use int as type string").

### 6.2 Runtime Panics
1.  Capture stderr.
2.  Filter stack traces to show only user code (hide `runtime/` internals).
3.  Highlight the line in the `.rg` file that caused the crash.

---

## 7. Future / Fancy Features

*   **`rugby bundle`**: Bundle the app into a single distributable `.rg` file (archive) or a Docker image.
*   **`rugby upgrade`**: Self-update the CLI.
*   **`rugby repl`**: Interactive shell (harder with compiled language, but possible via rapid compile-run loop).
