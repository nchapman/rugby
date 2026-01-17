# Rugby Performance Benchmarks

A benchmark suite comparing Rugby performance against Ruby, Go, and Rust.

## Prerequisites

- Go (1.22+)
- Rust (rustc)
- Ruby (3.0+)
- Rugby compiler (install from parent directory)

## Quick Start

```bash
# Build all benchmarks
make build

# Run all benchmarks
make run

# Verify output correctness
make verify

# Clean build artifacts
make clean
```

## Individual Benchmarks

```bash
# Run a single benchmark
make run-nsieve
make run-nbody
make run-binarytrees
make run-spectral-norm
make run-fannkuch-redux

# Build a single benchmark
make build-nsieve

# Verify a single benchmark
make verify-nsieve
```

## Benchmarks

| Benchmark | Category | What it tests |
|-----------|----------|---------------|
| **nsieve** | CPU, arrays | Prime sieve: loops, array allocation, bit manipulation |
| **nbody** | Float, structs | N-body simulation: classes, floating point math |
| **binarytrees** | Memory, recursion | GC pressure, recursive tree structures |
| **spectral-norm** | Numeric | Spectral norm: nested loops, matrix operations |
| **fannkuch-redux** | Algorithm | Pancake flipping: permutations, array manipulation |

## Input Sizes

| Benchmark | Test Input | Benchmark Input |
|-----------|------------|-----------------|
| nsieve | 4 | 9 |
| nbody | 1000 | 5000000 |
| binarytrees | 10 | 21 |
| spectral-norm | 100 | 5500 |
| fannkuch-redux | 7 | 12 |

## Directory Structure

```
benchmarks/
├── Makefile              # Build and run targets
├── README.md             # This file
├── nsieve/
│   ├── nsieve.rg         # Rugby implementation
│   ├── nsieve.go         # Go implementation
│   ├── nsieve.rb         # Ruby implementation
│   └── nsieve.rs         # Rust implementation
├── nbody/
│   └── ...
├── binarytrees/
│   └── ...
├── spectral-norm/
│   └── ...
└── fannkuch-redux/
    └── ...
```

## Verification

Each implementation produces identical output for the same input:

```bash
make verify
```

This runs each implementation with a small input and diffs the output.

## Notes

- Go and Rust implementations are compiled with optimizations (`-O` for Rust)
- Ruby runs interpreted
- Rugby compiles to Go, then to native binary
- All implementations are single-threaded for fair comparison

## Attribution

Benchmark algorithms adapted from [The Computer Language Benchmarks Game](https://benchmarksgame-team.pages.debian.net/benchmarksgame/) and [Programming-Language-Benchmarks](https://github.com/nicholaschapman/Programming-Language-Benchmarks).
