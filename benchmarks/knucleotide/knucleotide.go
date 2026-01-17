package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	dna := readInput()

	fmt.Print(writeFrequencies(dna, 1))
	fmt.Print(writeFrequencies(dna, 2))
	fmt.Println(writeCount(dna, "GGT"))
	fmt.Println(writeCount(dna, "GGTA"))
	fmt.Println(writeCount(dna, "GGTATT"))
	fmt.Println(writeCount(dna, "GGTATTTTAATT"))
	fmt.Println(writeCount(dna, "GGTATTTTAATTTATAGT"))
}

func readInput() string {
	fileName := "25000_in"
	if len(os.Args) > 1 {
		fileName = os.Args[1]
	}

	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var data strings.Builder

	// Skip until >THREE
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ">THREE") {
			break
		}
	}

	// Read the sequence
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] != '>' {
			data.WriteString(strings.ToUpper(line))
		}
	}

	return data.String()
}

func frequency(seq string, length int) map[string]int {
	counts := make(map[string]int)
	for i := 0; i <= len(seq)-length; i++ {
		counts[seq[i:i+length]]++
	}
	return counts
}

func writeFrequencies(seq string, length int) string {
	counts := frequency(seq, length)
	total := len(seq) - length + 1

	type kv struct {
		key   string
		value int
	}
	var sorted []kv
	for k, v := range counts {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].value != sorted[j].value {
			return sorted[i].value > sorted[j].value
		}
		return sorted[i].key < sorted[j].key
	})

	var result strings.Builder
	for _, kv := range sorted {
		fmt.Fprintf(&result, "%s %.3f\n", kv.key, 100.0*float64(kv.value)/float64(total))
	}
	result.WriteString("\n")
	return result.String()
}

func writeCount(seq, nucleotide string) string {
	counts := frequency(seq, len(nucleotide))
	count := counts[nucleotide]
	return fmt.Sprintf("%d\t%s", count, nucleotide)
}
