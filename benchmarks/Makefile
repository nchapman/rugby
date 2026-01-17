.PHONY: all build run clean help bench

# Working benchmarks (Rugby compiles and runs)
WORKING_BENCHMARKS = nsieve nbody binarytrees spectral-norm fannkuch-redux merkletrees

# Blocked benchmarks (Rugby has language issues, but Go/Rust/Ruby work)
BLOCKED_BENCHMARKS = coro-prime-sieve fasta knucleotide lru mandelbrot pidigits regex-redux

ALL_BENCHMARKS = $(WORKING_BENCHMARKS) $(BLOCKED_BENCHMARKS)

# Default target
all: build

# Run performance comparison and produce table
bench: build
	@./bench.sh

# Build all working benchmarks
build:
	@for b in $(WORKING_BENCHMARKS); do $(MAKE) build-$$b; done

# Build reference implementations (Go/Rust/Ruby) for blocked benchmarks
build-blocked:
	@for b in $(BLOCKED_BENCHMARKS); do $(MAKE) build-$$b-ref; done

# Run all working benchmarks
run:
	@for b in $(WORKING_BENCHMARKS); do $(MAKE) run-$$b; done

# Clean all build artifacts
clean:
	rm -f nsieve/nsieve_go nsieve/nsieve_rs nsieve/nsieve_rg
	rm -f nbody/nbody_go nbody/nbody_rs nbody/nbody_rg
	rm -f binarytrees/binarytrees_go binarytrees/binarytrees_rs binarytrees/binarytrees_rg
	rm -f spectral-norm/spectral-norm_go spectral-norm/spectral-norm_rs spectral-norm/spectral-norm_rg
	rm -f fannkuch-redux/fannkuch-redux_go fannkuch-redux/fannkuch-redux_rs fannkuch-redux/fannkuch-redux_rg
	rm -f merkletrees/merkletrees_go merkletrees/merkletrees_rs merkletrees/merkletrees_rg
	rm -f coro-prime-sieve/coro-prime-sieve_go coro-prime-sieve/coro-prime-sieve_rs
	rm -f fasta/fasta_go fasta/fasta_rs
	rm -f knucleotide/knucleotide_go knucleotide/knucleotide_rs
	rm -f lru/lru_go lru/lru_rs
	rm -f mandelbrot/mandelbrot_go mandelbrot/mandelbrot_rs
	rm -f pidigits/pidigits_go pidigits/pidigits_rs
	rm -f regex-redux/regex-redux_go regex-redux/regex-redux_rs

# Verify output correctness
verify:
	@for b in $(WORKING_BENCHMARKS); do $(MAKE) verify-$$b; done

#------------------------------------------------------------------------------
# nsieve
#------------------------------------------------------------------------------
build-nsieve:
	@echo "Building nsieve..."
	@cd nsieve && go build -o nsieve_go nsieve.go
	@cd nsieve && rustc -O -o nsieve_rs nsieve.rs
	@cd nsieve && rugby build nsieve.rg -o nsieve_rg

run-nsieve: build-nsieve
	@echo "=== nsieve (N=9) ==="
	@echo "Go:"    && time ./nsieve/nsieve_go 9
	@echo "Rust:"  && time ./nsieve/nsieve_rs 9
	@echo "Rugby:" && time ./nsieve/nsieve_rg 9
	@echo "Ruby:"  && time ruby ./nsieve/nsieve.rb 9
	@echo ""

verify-nsieve: build-nsieve
	@echo "Verifying nsieve..."
	@./nsieve/nsieve_go 4 > /tmp/nsieve_go.out
	@./nsieve/nsieve_rs 4 > /tmp/nsieve_rs.out
	@./nsieve/nsieve_rg 4 > /tmp/nsieve_rg.out
	@ruby ./nsieve/nsieve.rb 4 > /tmp/nsieve_rb.out
	@diff /tmp/nsieve_go.out /tmp/nsieve_rs.out && echo "  Go == Rust: OK"
	@diff /tmp/nsieve_go.out /tmp/nsieve_rg.out && echo "  Go == Rugby: OK"
	@diff /tmp/nsieve_go.out /tmp/nsieve_rb.out && echo "  Go == Ruby: OK"

#------------------------------------------------------------------------------
# nbody
#------------------------------------------------------------------------------
build-nbody:
	@echo "Building nbody..."
	@cd nbody && go build -o nbody_go nbody.go
	@cd nbody && rustc -O -o nbody_rs nbody.rs
	@cd nbody && rugby build nbody.rg -o nbody_rg

run-nbody: build-nbody
	@echo "=== nbody (N=5000000) ==="
	@echo "Go:"    && time ./nbody/nbody_go 5000000
	@echo "Rust:"  && time ./nbody/nbody_rs 5000000
	@echo "Rugby:" && time ./nbody/nbody_rg 5000000
	@echo "Ruby:"  && time ruby ./nbody/nbody.rb 5000000
	@echo ""

verify-nbody: build-nbody
	@echo "Verifying nbody..."
	@./nbody/nbody_go 1000 > /tmp/nbody_go.out
	@./nbody/nbody_rs 1000 > /tmp/nbody_rs.out
	@./nbody/nbody_rg 1000 > /tmp/nbody_rg.out
	@ruby ./nbody/nbody.rb 1000 > /tmp/nbody_rb.out
	@diff /tmp/nbody_go.out /tmp/nbody_rs.out && echo "  Go == Rust: OK"
	@diff /tmp/nbody_go.out /tmp/nbody_rg.out && echo "  Go == Rugby: OK"
	@diff /tmp/nbody_go.out /tmp/nbody_rb.out && echo "  Go == Ruby: OK"

#------------------------------------------------------------------------------
# binarytrees
#------------------------------------------------------------------------------
build-binarytrees:
	@echo "Building binarytrees..."
	@cd binarytrees && go build -o binarytrees_go binarytrees.go
	@cd binarytrees && rustc -O -o binarytrees_rs binarytrees.rs
	@cd binarytrees && rugby build binarytrees.rg -o binarytrees_rg

run-binarytrees: build-binarytrees
	@echo "=== binarytrees (N=21) ==="
	@echo "Go:"    && time ./binarytrees/binarytrees_go 21
	@echo "Rust:"  && time ./binarytrees/binarytrees_rs 21
	@echo "Rugby:" && time ./binarytrees/binarytrees_rg 21
	@echo "Ruby:"  && time ruby ./binarytrees/binarytrees.rb 21
	@echo ""

verify-binarytrees: build-binarytrees
	@echo "Verifying binarytrees..."
	@./binarytrees/binarytrees_go 10 > /tmp/binarytrees_go.out
	@./binarytrees/binarytrees_rs 10 > /tmp/binarytrees_rs.out
	@./binarytrees/binarytrees_rg 10 > /tmp/binarytrees_rg.out
	@ruby ./binarytrees/binarytrees.rb 10 > /tmp/binarytrees_rb.out
	@diff /tmp/binarytrees_go.out /tmp/binarytrees_rs.out && echo "  Go == Rust: OK"
	@diff /tmp/binarytrees_go.out /tmp/binarytrees_rg.out && echo "  Go == Rugby: OK"
	@diff /tmp/binarytrees_go.out /tmp/binarytrees_rb.out && echo "  Go == Ruby: OK"

#------------------------------------------------------------------------------
# spectral-norm
#------------------------------------------------------------------------------
build-spectral-norm:
	@echo "Building spectral-norm..."
	@cd spectral-norm && go build -o spectral-norm_go spectral-norm.go
	@cd spectral-norm && rustc -O -o spectral-norm_rs spectral-norm.rs
	@cd spectral-norm && rugby build spectral-norm.rg -o spectral-norm_rg

run-spectral-norm: build-spectral-norm
	@echo "=== spectral-norm (N=5500) ==="
	@echo "Go:"    && time ./spectral-norm/spectral-norm_go 5500
	@echo "Rust:"  && time ./spectral-norm/spectral-norm_rs 5500
	@echo "Rugby:" && time ./spectral-norm/spectral-norm_rg 5500
	@echo "Ruby:"  && time ruby ./spectral-norm/spectral-norm.rb 5500
	@echo ""

verify-spectral-norm: build-spectral-norm
	@echo "Verifying spectral-norm..."
	@./spectral-norm/spectral-norm_go 100 > /tmp/spectral-norm_go.out
	@./spectral-norm/spectral-norm_rs 100 > /tmp/spectral-norm_rs.out
	@./spectral-norm/spectral-norm_rg 100 > /tmp/spectral-norm_rg.out
	@ruby ./spectral-norm/spectral-norm.rb 100 > /tmp/spectral-norm_rb.out
	@diff /tmp/spectral-norm_go.out /tmp/spectral-norm_rs.out && echo "  Go == Rust: OK"
	@diff /tmp/spectral-norm_go.out /tmp/spectral-norm_rg.out && echo "  Go == Rugby: OK"
	@diff /tmp/spectral-norm_go.out /tmp/spectral-norm_rb.out && echo "  Go == Ruby: OK"

#------------------------------------------------------------------------------
# fannkuch-redux
#------------------------------------------------------------------------------
build-fannkuch-redux:
	@echo "Building fannkuch-redux..."
	@cd fannkuch-redux && go build -o fannkuch-redux_go fannkuch-redux.go
	@cd fannkuch-redux && rustc -O -o fannkuch-redux_rs fannkuch-redux.rs
	@cd fannkuch-redux && rugby build fannkuch-redux.rg -o fannkuch-redux_rg

run-fannkuch-redux: build-fannkuch-redux
	@echo "=== fannkuch-redux (N=12) ==="
	@echo "Go:"    && time ./fannkuch-redux/fannkuch-redux_go 12
	@echo "Rust:"  && time ./fannkuch-redux/fannkuch-redux_rs 12
	@echo "Rugby:" && time ./fannkuch-redux/fannkuch-redux_rg 12
	@echo "Ruby:"  && time ruby ./fannkuch-redux/fannkuch-redux.rb 12
	@echo ""

verify-fannkuch-redux: build-fannkuch-redux
	@echo "Verifying fannkuch-redux..."
	@./fannkuch-redux/fannkuch-redux_go 7 > /tmp/fannkuch-redux_go.out
	@./fannkuch-redux/fannkuch-redux_rs 7 > /tmp/fannkuch-redux_rs.out
	@./fannkuch-redux/fannkuch-redux_rg 7 > /tmp/fannkuch-redux_rg.out
	@ruby ./fannkuch-redux/fannkuch-redux.rb 7 > /tmp/fannkuch-redux_rb.out
	@diff /tmp/fannkuch-redux_go.out /tmp/fannkuch-redux_rs.out && echo "  Go == Rust: OK"
	@diff /tmp/fannkuch-redux_go.out /tmp/fannkuch-redux_rg.out && echo "  Go == Rugby: OK"
	@diff /tmp/fannkuch-redux_go.out /tmp/fannkuch-redux_rb.out && echo "  Go == Ruby: OK"

#------------------------------------------------------------------------------
# merkletrees
#------------------------------------------------------------------------------
build-merkletrees:
	@echo "Building merkletrees..."
	@cd merkletrees && go build -o merkletrees_go merkletrees.go
	@cd merkletrees && rustc -O -o merkletrees_rs merkletrees.rs
	@cd merkletrees && rugby build merkletrees.rg -o merkletrees_rg

run-merkletrees: build-merkletrees
	@echo "=== merkletrees (N=15) ==="
	@echo "Go:"    && time ./merkletrees/merkletrees_go 15
	@echo "Rust:"  && time ./merkletrees/merkletrees_rs 15
	@echo "Rugby:" && time ./merkletrees/merkletrees_rg 15
	@echo "Ruby:"  && time ruby ./merkletrees/merkletrees.rb 15
	@echo ""

verify-merkletrees: build-merkletrees
	@echo "Verifying merkletrees..."
	@./merkletrees/merkletrees_go 5 > /tmp/merkletrees_go.out
	@./merkletrees/merkletrees_rs 5 > /tmp/merkletrees_rs.out
	@./merkletrees/merkletrees_rg 5 > /tmp/merkletrees_rg.out
	@ruby ./merkletrees/merkletrees.rb 5 > /tmp/merkletrees_rb.out
	@diff /tmp/merkletrees_go.out /tmp/merkletrees_rs.out && echo "  Go == Rust: OK"
	@diff /tmp/merkletrees_go.out /tmp/merkletrees_rg.out && echo "  Go == Rugby: OK"
	@diff /tmp/merkletrees_go.out /tmp/merkletrees_rb.out && echo "  Go == Ruby: OK"

#------------------------------------------------------------------------------
# coro-prime-sieve (BLOCKED: spawn closure capture)
#------------------------------------------------------------------------------
build-coro-prime-sieve-ref:
	@echo "Building coro-prime-sieve (reference only)..."
	@cd coro-prime-sieve && go build -o coro-prime-sieve_go coro-prime-sieve.go
	@cd coro-prime-sieve && rustc -O -o coro-prime-sieve_rs coro-prime-sieve.rs

verify-coro-prime-sieve-ref: build-coro-prime-sieve-ref
	@echo "Verifying coro-prime-sieve..."
	@./coro-prime-sieve/coro-prime-sieve_go 100 > /tmp/coro-prime-sieve_go.out
	@./coro-prime-sieve/coro-prime-sieve_rs 100 > /tmp/coro-prime-sieve_rs.out
	@ruby ./coro-prime-sieve/coro-prime-sieve.rb 100 > /tmp/coro-prime-sieve_rb.out
	@diff /tmp/coro-prime-sieve_go.out /tmp/coro-prime-sieve_rs.out && echo "  Go == Rust: OK"
	@diff /tmp/coro-prime-sieve_go.out /tmp/coro-prime-sieve_rb.out && echo "  Go == Ruby: OK"

#------------------------------------------------------------------------------
# fasta (BLOCKED: tuple destructuring/literals)
#------------------------------------------------------------------------------
build-fasta-ref:
	@echo "Building fasta (reference only)..."
	@cd fasta && go build -o fasta_go fasta.go
	@cd fasta && rustc -O -o fasta_rs fasta.rs

verify-fasta-ref: build-fasta-ref
	@echo "Verifying fasta..."
	@./fasta/fasta_go 1000 > /tmp/fasta_go.out
	@./fasta/fasta_rs 1000 > /tmp/fasta_rs.out
	@ruby ./fasta/fasta.rb 1000 > /tmp/fasta_rb.out
	@diff /tmp/fasta_go.out /tmp/fasta_rs.out && echo "  Go == Rust: OK"
	@diff /tmp/fasta_go.out /tmp/fasta_rb.out && echo "  Go == Ruby: OK"

#------------------------------------------------------------------------------
# knucleotide (BLOCKED: bitwise AND)
# Note: Requires input file 25000_in
#------------------------------------------------------------------------------
build-knucleotide-ref:
	@echo "Building knucleotide (reference only)..."
	@cd knucleotide && go build -o knucleotide_go knucleotide.go
	@cd knucleotide && rustc -O -o knucleotide_rs knucleotide.rs

#------------------------------------------------------------------------------
# lru (BLOCKED: Hash type issues)
#------------------------------------------------------------------------------
build-lru-ref:
	@echo "Building lru (reference only)..."
	@cd lru && go build -o lru_go lru.go
	@cd lru && rustc -O -o lru_rs lru.rs

verify-lru-ref: build-lru-ref
	@echo "Verifying lru..."
	@./lru/lru_go 100 1000 > /tmp/lru_go.out
	@./lru/lru_rs 100 1000 > /tmp/lru_rs.out
	@ruby ./lru/lru.rb 100 1000 > /tmp/lru_rb.out
	@diff /tmp/lru_go.out /tmp/lru_rs.out && echo "  Go == Rust: OK"
	@diff /tmp/lru_go.out /tmp/lru_rb.out && echo "  Go == Ruby: OK"

#------------------------------------------------------------------------------
# mandelbrot (BLOCKED: 2D array indexing)
# Note: Rust version needs md5 crate (use cargo)
#------------------------------------------------------------------------------
build-mandelbrot-ref:
	@echo "Building mandelbrot (reference only)..."
	@cd mandelbrot && go build -o mandelbrot_go mandelbrot.go

#------------------------------------------------------------------------------
# pidigits (BLOCKED: pointer type properties)
# Note: Rust version needs num_bigint crate (use cargo)
#------------------------------------------------------------------------------
build-pidigits-ref:
	@echo "Building pidigits (reference only)..."
	@cd pidigits && go build -o pidigits_go pidigits.go

verify-pidigits-ref: build-pidigits-ref
	@echo "Verifying pidigits..."
	@./pidigits/pidigits_go 27 > /tmp/pidigits_go.out
	@ruby ./pidigits/pidigits.rb 27 > /tmp/pidigits_rb.out
	@diff /tmp/pidigits_go.out /tmp/pidigits_rb.out && echo "  Go == Ruby: OK"

#------------------------------------------------------------------------------
# regex-redux (BLOCKED: tuples, []byte.length)
# Note: Rust version needs regex crate (use cargo). Requires input file 25000_in
#------------------------------------------------------------------------------
build-regex-redux-ref:
	@echo "Building regex-redux (reference only)..."
	@cd regex-redux && go build -o regex-redux_go regex-redux.go

#------------------------------------------------------------------------------
# Help
#------------------------------------------------------------------------------
help:
	@echo "Rugby Performance Benchmarks"
	@echo ""
	@echo "Usage:"
	@echo "  make bench              Run benchmarks and show comparison table"
	@echo "  make build              Build all working benchmarks"
	@echo "  make build-blocked      Build reference implementations for blocked benchmarks"
	@echo "  make run                Run all working benchmarks (verbose)"
	@echo "  make verify             Verify output correctness"
	@echo "  make clean              Remove build artifacts"
	@echo ""
	@echo "Individual benchmarks:"
	@echo "  make build-nsieve       Build nsieve benchmark"
	@echo "  make run-nsieve         Run nsieve benchmark"
	@echo "  make verify-nsieve      Verify nsieve output"
	@echo ""
	@echo "Working benchmarks: $(WORKING_BENCHMARKS)"
	@echo "Blocked benchmarks: $(BLOCKED_BENCHMARKS)"
