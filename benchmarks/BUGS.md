# Rugby Benchmark Bugs

Issues encountered while implementing benchmarks.

## Fixed Issues

### 1. Bitwise Shift Operators - FIXED
`<<` and `>>` now work and return the correct `Int` type.

### 2. Scientific Notation for Floats - FIXED
Scientific notation like `4.84e+00` now works.

### 3. Pre-sized Arrays - FIXED
`Array<T>.new(size, default)` now works.

### 4. Top-level Constants with `def main` - FIXED
Use `const` keyword: `const IM = 139968`

### 5. Array Index Returns `any` Type - FIXED
Typed arrays now return correct element type when indexed.

### 6. Multi-line Function Calls - FIXED
Multi-line function calls now parse correctly.

### 7. Reassignment with Multi-return Functions - FIXED
Can now reassign variables in multi-return contexts.

### 10. Compound Assignment on Properties - FIXED
`body.vx += 1.0` now works correctly.

### Class Array Type Mismatch - FIXED
`Array<Body>` now correctly generates `[]*Body` and array indexing works properly.

### Compound Assignment on Array Index - FIXED
`arr[i] += value` now works correctly.

### `next` as Method/Property Name - FIXED
Can now use `next` as both a method name and property name.

### Interface Return Types - FIXED
Functions with interface return type can now return implementing classes.

### `continue` Keyword - FIXED
`continue` now works in loops.

### Hex/Binary/Octal Literals - FIXED
`0x80`, `0b1010`, `0o755` now parse correctly.

### Hash Type Alias - FIXED
`Hash<K, V>` now works as an alias for `Map`. Initialize with `{}`.

### Array `push` Method - FIXED
`arr.push(val)` now works correctly.

### Bitwise AND/OR/XOR Operators - FIXED
`&`, `|`, and `^` operators now work for integer types (bitwise AND, OR, XOR).

### Tuple Destructuring and Literals - FIXED
Tuple literals in arrays now generate proper struct literals, and tuple destructuring from array index now works.

```ruby
# Works now:
arr = [("a", 0.27), ("b", 0.12)]  # Array of tuples
_, p = genelist[i]                # Destructure from array index
```

### Spawn Closure Semantics - CLARIFIED
`spawn` now requires block syntax (`spawn { }` or `spawn do ... end`) to make closure capture semantics explicit. Variables are captured by reference (like Ruby blocks). Use local variables to capture values:

```ruby
# Correct pattern when variable will be reassigned:
input = ch           # capture current value
ch1 = Chan<Int>.new(1)
spawn { filter(input, ch1, prime) }  # uses captured value
ch = ch1             # reassignment doesn't affect spawned goroutine
```

---

## Blocking Issues (Preventing Benchmarks)

### `map` is Reserved Keyword
**Affects:** lru
**Status:** Cannot use `map` as a property name (conflicts with Go's map keyword)

```ruby
class LRU
  property map: Hash<Int, Node>  # Error: unexpected keyword map
end
```
**Workaround:** Rename to `cache` or similar.

### Hash Literal Type Inference
**Affects:** lru
**Status:** Empty hash literal `{}` infers `map[any]any` instead of typed map

```ruby
class LRU
  property cache: Hash<Int, Node>

  def initialize
    @cache = {}  # Error: cannot use map[any]any as map[int]Node
  end
end
```

### Hash Indexing Returns Value, Not Optional
**Affects:** lru
**Status:** Hash lookup returns value type directly, not optional, so nil checks don't work

```ruby
node = @cache[key]
if node == nil        # Error: mismatched types Node and untyped nil
  return 0, false
end
```

### Must Use `self.` to Call Own Methods
**Affects:** merkletrees (workaround applied)
**Status:** Cannot call methods on current object without explicit `self.`

```ruby
class Leaf
  def has_hash: Bool
    @hash != nil
  end

  def check: Bool
    has_hash      # Error: undefined: hasHash
    self.has_hash # Works - but shouldn't be required
  end
end
```

### 2D Array Indexing
**Affects:** mandelbrot
**Status:** Nested array indexing has type inference issues

```ruby
xloc = Array<Array<Float>>.new(size, Array<Float>.new(8, 0.0))
xloc[i / 8][i % 8] = value  # Error: undefined: unknown
```

### Pointer Types in Properties
**Affects:** pidigits
**Status:** Cannot declare pointer type properties, and value vs pointer mismatch

```ruby
class PiState
  property tmp1: big.Int  # Declares value type

  def initialize
    @tmp1 = big.NewInt(0)  # Error: *big.Int cannot be assigned to big.Int
  end
end
```
Need syntax like `property tmp1: *big.Int` or automatic pointer handling.

### `[]byte` Missing `.length` Method
**Affects:** regex-redux
**Status:** Go's `[]byte` type doesn't have `.length` method

```ruby
bytes, _ = ioutil.ReadAll(f)
len = bytes.length  # Error: type []byte has no field or method Length
```

---

## Benchmark Status

| Benchmark | Go | Rust | Ruby | Rugby | Status |
|-----------|:--:|:----:|:----:|:-----:|--------|
| nsieve | ✅ | ✅ | ✅ | ✅ | Working |
| binarytrees | ✅ | ✅ | ✅ | ✅ | Working |
| nbody | ✅ | ✅ | ✅ | ✅ | Working |
| spectral-norm | ✅ | ✅ | ✅ | ✅ | Working |
| fannkuch-redux | ✅ | ✅ | ✅ | ✅ | Working |
| merkletrees | ✅ | ✅ | ✅ | ✅ | Working |
| fasta | ✅ | ✅ | ✅ | ✅ | Working |
| mandelbrot | ✅ | ✅ | - | ❌ | Blocked: 2D arrays |
| coro-prime-sieve | ✅ | ✅ | ✅ | ✅ | Working |
| lru | ✅ | ✅ | ✅ | ❌ | Blocked: Hash type issues |
| pidigits | ✅ | ✅ | ✅ | ❌ | Blocked: pointer type properties |
| knucleotide | ✅ | ✅ | ✅ | ❌ | Blocked: Hash type inference |
| regex-redux | ✅ | ✅ | ✅ | ❌ | Blocked: []byte.length |

**8 of 13 Rugby benchmarks working**

## Notes

### String Indexing Returns String, Not Rune
String indexing `str[i]` returns a single-character `String`, not a `rune`. This follows Ruby semantics.
When using `fmt.Printf`, use `%s` instead of `%c`:
```ruby
fmt.Printf("%s", seq[i])  # Correct: prints character as string
fmt.Printf("%c", seq[i])  # Error: %c expects rune, not string
```

## Idiomatic Rugby Features Used

- `const NAME = value` - Top-level constants
- `Array<T>.new(n, default)` - Pre-sized array allocation
- `n.times -> { }` - Block iteration
- `iterations.times -> { }` - Loop with block
- `1 << n`, `n >> 1` - Bitwise shift operators
- `a & b`, `a | b`, `a ^ b` - Bitwise AND, OR, XOR operators
- `0x80`, `0b1010` - Hex and binary literals
- Multi-line `Body.new(...)` calls
- `property`, `getter` class declarations
- Optional types with `Node?`
- Lambda syntax `-> { |i| ... }`
- `Hash<K, V>` type alias with `{}` literal
- `continue` in loops
- `spawn { ... }` block syntax for goroutines
- `Chan<T>` channels with `<<` send and `.receive`
- Interface return types
- Tuple types `(T1, T2)` with array literal and destructuring support
