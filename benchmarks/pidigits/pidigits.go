package main

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"strconv"
)

var (
	tmp1  = big.NewInt(0)
	tmp2  = big.NewInt(0)
	y2    = big.NewInt(1)
	bigk  = big.NewInt(0)
	accum = big.NewInt(0)
	denom = big.NewInt(1)
	numer = big.NewInt(1)
	ten   = big.NewInt(10)
	three = big.NewInt(3)
	four  = big.NewInt(4)
)

func nextTerm(k int64) int64 {
	for {
		k++
		y2.SetInt64(k*2 + 1)
		bigk.SetInt64(k)

		tmp1.Lsh(numer, 1)
		accum.Add(accum, tmp1)
		accum.Mul(accum, y2)
		denom.Mul(denom, y2)
		numer.Mul(numer, bigk)

		if accum.Cmp(numer) > 0 {
			return k
		}
	}
}

func extractDigit(nth *big.Int) int64 {
	tmp1.Mul(nth, numer)
	tmp2.Add(tmp1, accum)
	tmp1.Div(tmp2, denom)
	return tmp1.Int64()
}

func nextDigit(k int64) (int64, int64) {
	for {
		k = nextTerm(k)
		d3 := extractDigit(three)
		d4 := extractDigit(four)
		if d3 == d4 {
			return d3, k
		}
	}
}

func eliminateDigit(d int64) {
	tmp1.SetInt64(d)
	accum.Sub(accum, tmp1.Mul(denom, tmp1))
	accum.Mul(accum, ten)
	numer.Mul(numer, ten)
}

func main() {
	n := 27
	if len(os.Args) > 1 {
		if s, err := strconv.Atoi(os.Args[1]); err == nil {
			n = s
		}
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	line := make([]byte, 0, 10)
	var d, k int64
	for i := 1; i <= n; i++ {
		d, k = nextDigit(k)
		line = append(line, byte(d)+'0')
		if len(line) == 10 {
			fmt.Fprintf(w, "%s\t:%d\n", string(line), i)
			line = line[:0]
		}
		eliminateDigit(d)
	}
	if len(line) > 0 {
		fmt.Fprintf(w, "%-10s\t:%d\n", string(line), n)
	}
}
