package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rugby/codegen"
	"rugby/lexer"
	"rugby/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: rugby <file.rby>")
		os.Exit(1)
	}

	inputPath := os.Args[1]

	// Read input file
	source, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Lex
	l := lexer.New(string(source))

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Parse errors:")
		for _, e := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", e)
		}
		os.Exit(1)
	}

	// Generate
	gen := codegen.New()
	output, err := gen.Generate(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Code generation error: %v\n", err)
		os.Exit(1)
	}

	// Write output file
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".go"
	err = os.WriteFile(outputPath, []byte(output), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Compiled %s -> %s\n", inputPath, outputPath)
}
