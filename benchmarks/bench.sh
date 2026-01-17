#!/bin/bash
# Rugby Performance Benchmark Runner
# Compares Go, Rugby, and Ruby performance

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Benchmark configurations: name:go_bin:rugby_bin:ruby_file:input
BENCHMARKS=(
    "nsieve:nsieve/nsieve_go:nsieve/nsieve_rg:nsieve/nsieve.rb:9"
    "binarytrees:binarytrees/binarytrees_go:binarytrees/binarytrees_rg:binarytrees/binarytrees.rb:18"
    "nbody:nbody/nbody_go:nbody/nbody_rg:nbody/nbody.rb:500000"
    "spectral-norm:spectral-norm/spectral-norm_go:spectral-norm/spectral-norm_rg:spectral-norm/spectral-norm.rb:3000"
    "fannkuch-redux:fannkuch-redux/fannkuch-redux_go:fannkuch-redux/fannkuch-redux_rg:fannkuch-redux/fannkuch-redux.rb:10"
)

# Function to time a command and return seconds
time_cmd() {
    local start end
    start=$(perl -MTime::HiRes=time -e 'print time')
    "$@" > /dev/null 2>&1
    end=$(perl -MTime::HiRes=time -e 'print time')
    perl -e "printf '%.3f', $end - $start"
}

# Print header
echo ""
echo -e "${BLUE}Rugby Performance Benchmarks${NC}"
echo "=============================="
echo ""
printf "%-20s %10s %10s %10s %12s %12s\n" "Benchmark" "Go" "Rugby" "Ruby" "Rugby/Go" "Ruby/Rugby"
printf "%-20s %10s %10s %10s %12s %12s\n" "---------" "--" "-----" "----" "--------" "----------"

# Run each benchmark
for bench in "${BENCHMARKS[@]}"; do
    IFS=':' read -r name go_bin rugby_bin ruby_file input <<< "$bench"

    # Time each implementation
    go_time=$(time_cmd ./$go_bin $input)
    rugby_time=$(time_cmd ./$rugby_bin $input)
    ruby_time=$(time_cmd ruby $ruby_file $input)

    # Calculate ratios
    rugby_go_ratio=$(perl -e "printf '%.1fx', $rugby_time / $go_time")
    ruby_rugby_ratio=$(perl -e "printf '%.1fx', $ruby_time / $rugby_time")

    # Format times
    go_fmt="${go_time}s"
    rugby_fmt="${rugby_time}s"
    ruby_fmt="${ruby_time}s"

    # Color the Rugby/Go ratio (green if <2x, yellow if <10x, red otherwise)
    ratio_val=$(perl -e "print $rugby_time / $go_time")
    if (( $(echo "$ratio_val < 2" | bc -l) )); then
        rugby_go_color="${GREEN}${rugby_go_ratio}${NC}"
    elif (( $(echo "$ratio_val < 10" | bc -l) )); then
        rugby_go_color="${YELLOW}${rugby_go_ratio}${NC}"
    else
        rugby_go_color="${RED}${rugby_go_ratio}${NC}"
    fi

    # Ruby/Rugby is always green (Rugby faster than Ruby)
    ruby_rugby_color="${GREEN}${ruby_rugby_ratio}${NC}"

    printf "%-20s %10s %10s %10s " "$name (n=$input)" "$go_fmt" "$rugby_fmt" "$ruby_fmt"
    printf "%b %b\n" "$rugby_go_color" "$ruby_rugby_color"
done

echo ""
echo -e "${BLUE}Legend:${NC}"
echo -e "  Rugby/Go:    ${GREEN}<2x${NC} | ${YELLOW}2-10x${NC} | ${RED}>10x${NC} slower than Go"
echo -e "  Ruby/Rugby:  How much faster Rugby is than Ruby"
echo ""
