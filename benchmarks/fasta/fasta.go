package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

const (
	IM         = 139968
	IA         = 3877
	IC         = 29573
	LINE_WIDTH = 60
)

var ALU = "GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGGAGGCGGAGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCCAGCCTGGGCGACAGAGCGAGACTCCGTCTCAAAAA"

type Random struct {
	last int
}

func NewRandom(seed int) *Random {
	return &Random{last: seed}
}

func (r *Random) Next() float64 {
	r.last = (r.last*IA + IC) % IM
	return float64(r.last) / float64(IM)
}

type AminoAcid struct {
	c byte
	p float64
}

func repeatFasta(out *bufio.Writer, seq string, n int) {
	length := len(seq)
	pos := 0

	for n > 0 {
		lineLen := n
		if lineLen > LINE_WIDTH {
			lineLen = LINE_WIDTH
		}

		for i := 0; i < lineLen; i++ {
			out.WriteByte(seq[(pos+i)%length])
		}
		out.WriteByte('\n')

		pos = (pos + lineLen) % length
		n -= lineLen
	}
}

func randomFasta(out *bufio.Writer, genelist []AminoAcid, n int, rng *Random) {
	// Build cumulative probabilities
	cum := 0.0
	probs := make([]float64, len(genelist))
	for i := range genelist {
		cum += genelist[i].p
		probs[i] = cum
	}

	for n > 0 {
		lineLen := n
		if lineLen > LINE_WIDTH {
			lineLen = LINE_WIDTH
		}

		for j := 0; j < lineLen; j++ {
			r := rng.Next()
			for k, prob := range probs {
				if prob >= r {
					out.WriteByte(genelist[k].c)
					break
				}
			}
		}
		out.WriteByte('\n')

		n -= lineLen
	}
}

func main() {
	n := 1000
	if len(os.Args) > 1 {
		n, _ = strconv.Atoi(os.Args[1])
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	iub := []AminoAcid{
		{'a', 0.27},
		{'c', 0.12},
		{'g', 0.12},
		{'t', 0.27},
		{'B', 0.02},
		{'D', 0.02},
		{'H', 0.02},
		{'K', 0.02},
		{'M', 0.02},
		{'N', 0.02},
		{'R', 0.02},
		{'S', 0.02},
		{'V', 0.02},
		{'W', 0.02},
		{'Y', 0.02},
	}

	homosapiens := []AminoAcid{
		{'a', 0.3029549426680},
		{'c', 0.1979883004921},
		{'g', 0.1975473066391},
		{'t', 0.3015094502008},
	}

	rng := NewRandom(42)

	fmt.Fprint(out, ">ONE Homo sapiens alu\n")
	repeatFasta(out, ALU, 2*n)

	fmt.Fprint(out, ">TWO IUB ambiguity codes\n")
	randomFasta(out, iub, 3*n, rng)

	fmt.Fprint(out, ">THREE Homo sapiens frequency\n")
	randomFasta(out, homosapiens, 5*n, rng)
}
