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
		fmt.Fprintln(os.Stderr, "Usage: rugby <file.rg> [file2.rg ...]")
		fmt.Fprintln(os.Stderr, "       rugby <directory>")
		os.Exit(1)
	}

	// Collect all .rg files to compile
	var files []string
	for _, arg := range os.Args[1:] {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if info.IsDir() {
			// Find all .rg files in directory
			entries, err := os.ReadDir(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
				os.Exit(1)
			}
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".rg") {
					files = append(files, filepath.Join(arg, entry.Name()))
				}
			}
		} else {
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No .rg files found")
		os.Exit(1)
	}

	// Compile each file
	hasErrors := false
	for _, inputPath := range files {
		if err := compileFile(inputPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			hasErrors = true
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func compileFile(inputPath string) error {
	// Read input file
	source, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %v", inputPath, err)
	}

	// Lex
	l := lexer.New(string(source))

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		var errMsg strings.Builder
		errMsg.WriteString(fmt.Sprintf("Parse errors in %s:\n", inputPath))
		for _, e := range p.Errors() {
			errMsg.WriteString(fmt.Sprintf("  %s:%s\n", inputPath, e))
		}
		return fmt.Errorf("%s", errMsg.String())
	}

	// Generate
	gen := codegen.New()
	output, err := gen.Generate(program)
	if err != nil {
		return fmt.Errorf("Code generation error in %s: %v", inputPath, err)
	}

	// Write output file
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".go"
	err = os.WriteFile(outputPath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("Error writing file %s: %v", outputPath, err)
	}

	fmt.Printf("Compiled %s -> %s\n", inputPath, outputPath)
	return nil
}
