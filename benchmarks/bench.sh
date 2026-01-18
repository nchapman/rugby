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
    "knucleotide:knucleotide:knucleotide/250000_in"
    "regex-redux:regex-redux:regex-redux/100000_in"
)

# Function to time a command and return seconds
# Runs a warmup first to ensure binary is in disk cache
time_cmd() {
    local start end
    # Warmup run to load binary into disk cache
    "$@" > /dev/null 2>&1
    # Timed run
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
            if [[ -f "Cargo.toml" ]]; then
                # Use cargo for projects with dependencies
                # Cargo converts hyphens to underscores in binary names
                cargo_name="${name//-/_}_rs"
                cargo build --release 2>/dev/null && cp "target/release/${cargo_name}" "${name}_rs" 2>/dev/null || echo -e "${YELLOW}  Rust build failed for $name${NC}"
            else
                rustc -O -o "${name}_rs" "${name}.rs" 2>/dev/null || echo -e "${YELLOW}  Rust build failed for $name${NC}"
            fi
        fi
    fi

    # Build Crystal version if source is newer than binary
    if [[ -f "${name}.cr" ]]; then
        if [[ ! -x "${name}_cr" ]] || [[ "${name}.cr" -nt "${name}_cr" ]]; then
            crystal build --release -o "${name}_cr" "${name}.cr" 2>/dev/null || echo -e "${YELLOW}  Crystal build failed for $name${NC}"
        fi
    fi

    cd "$SCRIPT_DIR"
    return 0
}

# Color a ratio value (padded to 6 chars before color codes)
# Lower is better (1.0x = same speed, 2.0x = twice as slow)
color_ratio() {
    local ratio="$1"
    local padded=$(printf "%6s" "${ratio}x")
    if perl -e "exit($ratio < 1.5 ? 0 : 1)" 2>/dev/null; then
        echo -e "${GREEN}${padded}${NC}"
    elif perl -e "exit($ratio < 3 ? 0 : 1)" 2>/dev/null; then
        echo -e "${YELLOW}${padded}${NC}"
    else
        echo -e "${RED}${padded}${NC}"
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
printf "%-18s %8s %8s %8s %8s %8s  %-6s  %-6s\n" \
    "Benchmark" "Go" "Rugby" "Rust" "Crystal" "Ruby" "vs Go" "vs Rust"
printf "%-18s %8s %8s %8s %8s %8s  %-6s  %-6s\n" \
    "---------" "--" "-----" "----" "-------" "----" "-----" "-------"

# Run each benchmark
for bench in "${BENCHMARKS[@]}"; do
    IFS=':' read -r name dir input <<< "$bench"

    go_bin="$dir/${name}_go"
    rugby_bin="$dir/${name}_rg"
    rust_bin="$dir/${name}_rs"
    crystal_bin="$dir/${name}_cr"
    ruby_file="$dir/${name}.rb"

    # Check if binaries exist
    if [[ ! -x "$rugby_bin" ]]; then
        printf "%-18s ${GRAY}skipped (build failed)${NC}\n" "$name"
        continue
    fi

    # Time each implementation
    go_time="N/A"
    rugby_time="N/A"
    rust_time="N/A"
    crystal_time="N/A"
    ruby_time="N/A"

    if [[ -x "$go_bin" ]]; then
        go_time=$(time_cmd "./$go_bin" "$input")
    fi

    rugby_time=$(time_cmd "./$rugby_bin" "$input")

    if [[ -x "$rust_bin" ]]; then
        rust_time=$(time_cmd "./$rust_bin" "$input")
    fi

    if [[ -x "$crystal_bin" ]]; then
        crystal_time=$(time_cmd "./$crystal_bin" "$input")
    fi

    if [[ -f "$ruby_file" ]] && command -v ruby &> /dev/null; then
        ruby_time=$(time_cmd ruby "$ruby_file" "$input")
    fi

    # Format times
    go_fmt=$([ "$go_time" != "N/A" ] && printf "%6.3fs" "$go_time" || echo "   N/A")
    rugby_fmt=$(printf "%6.3fs" "$rugby_time")
    rust_fmt=$([ "$rust_time" != "N/A" ] && printf "%6.3fs" "$rust_time" || echo "   N/A")
    crystal_fmt=$([ "$crystal_time" != "N/A" ] && printf "%6.3fs" "$crystal_time" || echo "   N/A")
    ruby_fmt=$([ "$ruby_time" != "N/A" ] && printf "%6.3fs" "$ruby_time" || echo "   N/A")

    # Calculate ratios (uncolored for proper alignment)
    if [[ "$go_time" != "N/A" ]]; then
        rugby_go=$(perl -e "printf '%.1f', $rugby_time / $go_time")
    else
        rugby_go="N/A"
    fi

    if [[ "$rust_time" != "N/A" ]]; then
        rugby_rust=$(perl -e "printf '%.1f', $rugby_time / $rust_time")
    else
        rugby_rust="N/A"
    fi

    # Format ratios with color (6 chars each for alignment)
    if [[ "$rugby_go" != "N/A" ]]; then
        rugby_go_fmt=$(color_ratio "$rugby_go")
    else
        rugby_go_fmt="${GRAY}   N/A${NC}"
    fi

    if [[ "$rugby_rust" != "N/A" ]]; then
        rugby_rust_fmt=$(color_ratio "$rugby_rust")
    else
        rugby_rust_fmt="${GRAY}   N/A${NC}"
    fi

    printf "%-18s %8s %8s %8s %8s %8s  %b  %b\n" \
        "$name" "$go_fmt" "$rugby_fmt" "$rust_fmt" "$crystal_fmt" "$ruby_fmt" \
        "$rugby_go_fmt" "$rugby_rust_fmt"
done

echo ""
echo -e "${BLUE}Legend:${NC} ${GREEN}<1.5x${NC} | ${YELLOW}1.5-3x${NC} | ${RED}>3x${NC} slower"
echo ""
