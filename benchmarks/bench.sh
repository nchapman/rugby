#!/bin/bash
# Rugby Performance Benchmark Runner
# Compares Go, Rugby, Rust, and Ruby performance

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUGBY_ROOT="$(dirname "$SCRIPT_DIR")"
RUGBY="$RUGBY_ROOT/rugby"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
GRAY='\033[0;90m'
NC='\033[0m' # No Color

# Benchmark configurations: name:dir:input
BENCHMARKS=(
    "nsieve:nsieve:9"
    "binarytrees:binarytrees:18"
    "nbody:nbody:500000"
    "spectral-norm:spectral-norm:3000"
    "fannkuch-redux:fannkuch-redux:10"
    "fasta:fasta:1000000"
    "mandelbrot:mandelbrot:4000"
    "pidigits:pidigits:2000"
    "merkletrees:merkletrees:15"
    "lru:lru:10000"
    "coro-prime-sieve:coro-prime-sieve:1000"
    "knucleotide:knucleotide:knucleotide/25000_in"
    "regex-redux:regex-redux:regex-redux/100000_in"
)

# Function to time a command and return seconds
time_cmd() {
    local start end
    start=$(perl -MTime::HiRes=time -e 'print time')
    "$@" > /dev/null 2>&1
    end=$(perl -MTime::HiRes=time -e 'print time')
    perl -e "printf '%.3f', $end - $start"
}

# Build the rugby compiler if needed
build_rugby() {
    if [[ ! -x "$RUGBY" ]] || [[ "$RUGBY_ROOT/main.go" -nt "$RUGBY" ]]; then
        echo -e "${BLUE}Building rugby compiler...${NC}"
        (cd "$RUGBY_ROOT" && go build -o rugby .)
    fi
}

# Build a benchmark's Go, Rugby, and Rust versions
build_benchmark() {
    local dir="$1"
    local name="$2"

    cd "$SCRIPT_DIR/$dir"

    # Build Go version if source is newer than binary
    if [[ -f "${name}.go" ]]; then
        if [[ ! -x "${name}_go" ]] || [[ "${name}.go" -nt "${name}_go" ]]; then
            go build -o "${name}_go" "${name}.go" 2>/dev/null || echo -e "${YELLOW}  Go build failed for $name${NC}"
        fi
    fi

    # Build Rugby version if source is newer than binary
    if [[ -f "${name}.rg" ]]; then
        if [[ ! -x "${name}_rg" ]] || [[ "${name}.rg" -nt "${name}_rg" ]]; then
            "$RUGBY" build "${name}.rg" -o "${name}_rg" 2>/dev/null || echo -e "${YELLOW}  Rugby build failed for $name${NC}"
        fi
    fi

    # Build Rust version if source is newer than binary
    if [[ -f "${name}.rs" ]]; then
        if [[ ! -x "${name}_rs" ]] || [[ "${name}.rs" -nt "${name}_rs" ]]; then
            rustc -O -o "${name}_rs" "${name}.rs" 2>/dev/null || echo -e "${YELLOW}  Rust build failed for $name${NC}"
        fi
    fi

    cd "$SCRIPT_DIR"
    return 0
}

# Color a ratio value
color_ratio() {
    local ratio="$1"
    if perl -e "exit($ratio < 1.5 ? 0 : 1)" 2>/dev/null; then
        echo -e "${GREEN}${ratio}x${NC}"
    elif perl -e "exit($ratio < 3 ? 0 : 1)" 2>/dev/null; then
        echo -e "${YELLOW}${ratio}x${NC}"
    else
        echo -e "${RED}${ratio}x${NC}"
    fi
}

# Main
cd "$SCRIPT_DIR"

# Build rugby compiler
build_rugby

echo ""
echo -e "${BLUE}Building benchmarks...${NC}"

# Build all benchmarks first
for bench in "${BENCHMARKS[@]}"; do
    IFS=':' read -r name dir input <<< "$bench"
    build_benchmark "$dir" "$name"
done

echo -e "${GREEN}Done${NC}"
echo ""
echo -e "${BLUE}Rugby Performance Benchmarks${NC}"
echo "=============================="
echo ""
printf "%-22s %8s %8s %8s %8s  %9s %9s\n" "Benchmark" "Go" "Rugby" "Rust" "Ruby" "Rugby/Go" "Rugby/Rust"
printf "%-22s %8s %8s %8s %8s  %9s %9s\n" "---------" "--" "-----" "----" "----" "--------" "----------"

# Run each benchmark
for bench in "${BENCHMARKS[@]}"; do
    IFS=':' read -r name dir input <<< "$bench"

    go_bin="$dir/${name}_go"
    rugby_bin="$dir/${name}_rg"
    rust_bin="$dir/${name}_rs"
    ruby_file="$dir/${name}.rb"

    # Check if binaries exist
    if [[ ! -x "$rugby_bin" ]]; then
        printf "%-22s ${GRAY}skipped (build failed)${NC}\n" "$name"
        continue
    fi

    # Time each implementation
    go_time="N/A"
    rugby_time="N/A"
    rust_time="N/A"
    ruby_time="N/A"

    if [[ -x "$go_bin" ]]; then
        go_time=$(time_cmd ./$go_bin $input)
    fi

    rugby_time=$(time_cmd ./$rugby_bin $input)

    if [[ -x "$rust_bin" ]]; then
        rust_time=$(time_cmd ./$rust_bin $input)
    fi

    if [[ -f "$ruby_file" ]] && command -v ruby &> /dev/null; then
        ruby_time=$(time_cmd ruby $ruby_file $input)
    fi

    # Format times
    go_fmt=$([ "$go_time" != "N/A" ] && echo "${go_time}s" || echo "N/A")
    rugby_fmt="${rugby_time}s"
    rust_fmt=$([ "$rust_time" != "N/A" ] && echo "${rust_time}s" || echo "N/A")
    ruby_fmt=$([ "$ruby_time" != "N/A" ] && echo "${ruby_time}s" || echo "N/A")

    # Calculate and color ratios
    if [[ "$go_time" != "N/A" ]]; then
        rugby_go=$(perl -e "printf '%.1f', $rugby_time / $go_time")
        rugby_go_colored=$(color_ratio "$rugby_go")
    else
        rugby_go_colored="${GRAY}N/A${NC}"
    fi

    if [[ "$rust_time" != "N/A" ]]; then
        rugby_rust=$(perl -e "printf '%.1f', $rugby_time / $rust_time")
        rugby_rust_colored=$(color_ratio "$rugby_rust")
    else
        rugby_rust_colored="${GRAY}N/A${NC}"
    fi

    printf "%-22s %8s %8s %8s %8s  " "$name" "$go_fmt" "$rugby_fmt" "$rust_fmt" "$ruby_fmt"
    printf "%b %b\n" "$rugby_go_colored" "$rugby_rust_colored"
done

echo ""
echo -e "${BLUE}Legend:${NC} ${GREEN}<1.5x${NC} | ${YELLOW}1.5-3x${NC} | ${RED}>3x${NC} slower"
echo ""
