package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

func main() {
	fileName := "25000_in"
	if len(os.Args) > 1 {
		fileName = os.Args[1]
	}

	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	originalLen := len(bytes)

	// Clean the input - remove headers and newlines
	cleanRE := regexp.MustCompile("(>[^\n]+)?\n")
	bytes = cleanRE.ReplaceAllLiteral(bytes, []byte(""))
	cleanedLen := len(bytes)

	// Variant patterns to count
	variants := []string{
		"agggtaaa|tttaccct",
		"[cgt]gggtaaa|tttaccc[acg]",
		"a[act]ggtaaa|tttacc[agt]t",
		"ag[act]gtaaa|tttac[agt]ct",
		"agg[act]taaa|ttta[agt]cct",
		"aggg[acg]aaa|ttt[cgt]ccct",
		"agggt[cgt]aa|tt[acg]accct",
		"agggta[cgt]a|t[acg]taccct",
		"agggtaa[cgt]|[acg]ttaccct",
	}

	// Count each variant
	for _, pattern := range variants {
		re := regexp.MustCompile(pattern)
		matches := re.FindAll(bytes, -1)
		count := len(matches)
		fmt.Printf("%s %d\n", pattern, count)
	}

	// Substitutions
	substitutions := []struct {
		pattern     string
		replacement string
	}{
		{"tHa[Nt]", "<4>"},
		{"aND|caN|Ha[DS]|WaS", "<3>"},
		{"a[NSt]|BY", "<2>"},
		{"<[^>]*>", "|"},
		{`\|[^|][^|]*\|`, "-"},
	}

	for _, sub := range substitutions {
		re := regexp.MustCompile(sub.pattern)
		bytes = re.ReplaceAll(bytes, []byte(sub.replacement))
	}

	fmt.Printf("\n%d\n%d\n%d\n", originalLen, cleanedLen, len(bytes))
}
