package builder

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsUpToDate(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rugby-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	project := &Project{Root: tmpDir, GenDir: filepath.Join(tmpDir, "gen")}
	b := &Builder{project: project}

	sourcePath := filepath.Join(tmpDir, "test.rg")
	outputPath := filepath.Join(tmpDir, "test.go")

	// Test: output doesn't exist
	if err := os.WriteFile(sourcePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if b.isUpToDate(sourcePath, outputPath) {
		t.Error("isUpToDate should return false when output doesn't exist")
	}

	// Test: output exists but is older
	if err := os.WriteFile(outputPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	// Make source newer
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(sourcePath, []byte("test2"), 0644); err != nil {
		t.Fatal(err)
	}
	if b.isUpToDate(sourcePath, outputPath) {
		t.Error("isUpToDate should return false when source is newer")
	}

	// Test: output is newer than source
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(outputPath, []byte("test3"), 0644); err != nil {
		t.Fatal(err)
	}
	if !b.isUpToDate(sourcePath, outputPath) {
		t.Error("isUpToDate should return true when output is newer")
	}
}

func TestCachedMeta(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rugby-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	project := &Project{Root: tmpDir}
	b := &Builder{project: project}
	metaPath := filepath.Join(tmpDir, "test.go.meta")

	// Test: read non-existent meta returns false
	if b.readCachedMeta(metaPath) {
		t.Error("readCachedMeta should return false for non-existent file")
	}

	// Test: write and read true
	b.writeCachedMeta(metaPath, true)
	if !b.readCachedMeta(metaPath) {
		t.Error("readCachedMeta should return true after writing true")
	}

	// Test: write and read false
	b.writeCachedMeta(metaPath, false)
	if b.readCachedMeta(metaPath) {
		t.Error("readCachedMeta should return false after writing false")
	}
}

func TestNeedsRebuild(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rugby-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	genDir := filepath.Join(tmpDir, "gen")
	if err := os.MkdirAll(genDir, 0755); err != nil {
		t.Fatal(err)
	}

	project := &Project{Root: tmpDir, GenDir: genDir}
	b := &Builder{project: project}

	binPath := filepath.Join(tmpDir, "test-bin")
	genFile := filepath.Join(genDir, "test.go")

	// Test: binary doesn't exist
	if err := os.WriteFile(genFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if !b.needsRebuild([]string{genFile}, binPath) {
		t.Error("needsRebuild should return true when binary doesn't exist")
	}

	// Test: binary is newer than source
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(binPath, []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if b.needsRebuild([]string{genFile}, binPath) {
		t.Error("needsRebuild should return false when binary is newer")
	}

	// Test: source is newer than binary
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(genFile, []byte("package main // modified"), 0644); err != nil {
		t.Fatal(err)
	}
	if !b.needsRebuild([]string{genFile}, binPath) {
		t.Error("needsRebuild should return true when source is newer")
	}
}

func TestGenerateGoMod(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rugby-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	project := &Project{Root: tmpDir}
	b := &Builder{project: project}

	// Test: no rugby.mod generates minimal go.mod
	content := b.generateGoMod()
	if content == "" {
		t.Error("generateGoMod should return content even without rugby.mod")
	}
	if !contains(content, "module main") {
		t.Error("generated go.mod should have 'module main'")
	}
	if !contains(content, RuntimeModule) {
		t.Errorf("generated go.mod should contain runtime module %s", RuntimeModule)
	}

	// Test: with rugby.mod
	rugbyMod := `module myapp

go 1.25
`
	if err := os.WriteFile(filepath.Join(tmpDir, "rugby.mod"), []byte(rugbyMod), 0644); err != nil {
		t.Fatal(err)
	}
	content = b.generateGoMod()
	if !contains(content, "module myapp") {
		t.Error("generated go.mod should preserve module name from rugby.mod")
	}
	if !contains(content, RuntimeModule) {
		t.Errorf("generated go.mod should contain runtime module %s", RuntimeModule)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseGoError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		genDir   string
		wantFile string
		wantLine int
		wantCol  int
		wantMsg  string
	}{
		{
			name:     "relative path with line and column",
			input:    "../../foo.rg:10:5: undefined: bar",
			genDir:   "/project/.rugby/gen",
			wantFile: "/project/foo.rg",
			wantLine: 10,
			wantCol:  5,
			wantMsg:  "undefined: bar",
		},
		{
			name:     "relative path with line only",
			input:    "../../foo.rg:10: declared and not used: x",
			genDir:   "/project/.rugby/gen",
			wantFile: "/project/foo.rg",
			wantLine: 10,
			wantCol:  0,
			wantMsg:  "declared and not used: x",
		},
		{
			name:     "absolute path",
			input:    "/home/user/project/foo.rg:15:3: type mismatch",
			genDir:   "/project/.rugby/gen",
			wantFile: "/home/user/project/foo.rg",
			wantLine: 15,
			wantCol:  3,
			wantMsg:  "type mismatch",
		},
		{
			name:     "non-rg file returns empty",
			input:    "main.go:10:5: some error",
			genDir:   "/project/.rugby/gen",
			wantFile: "",
			wantLine: 0,
			wantCol:  0,
			wantMsg:  "main.go:10:5: some error",
		},
		{
			name:     "no colon returns empty",
			input:    "some random error message",
			genDir:   "/project/.rugby/gen",
			wantFile: "",
			wantLine: 0,
			wantCol:  0,
			wantMsg:  "some random error message",
		},
		{
			name:     "deeply nested relative path",
			input:    "../../src/main.rg:5: error here",
			genDir:   "/project/.rugby/gen",
			wantFile: "/project/src/main.rg",
			wantLine: 5,
			wantCol:  0,
			wantMsg:  "error here",
		},
		{
			name:     "message with colons",
			input:    "test.rg:1: invalid operation: x * y (mismatched types: int and string)",
			genDir:   "/project",
			wantFile: "/project/test.rg",
			wantLine: 1,
			wantCol:  0,
			wantMsg:  "invalid operation: x * y (mismatched types: int and string)",
		},
		{
			name:     "high column number",
			input:    "test.rg:1:150: error at wide position",
			genDir:   "/project",
			wantFile: "/project/test.rg",
			wantLine: 1,
			wantCol:  150,
			wantMsg:  "error at wide position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, line, col, msg := parseGoError(tt.input, tt.genDir)

			if file != tt.wantFile {
				t.Errorf("file = %q, want %q", file, tt.wantFile)
			}
			if line != tt.wantLine {
				t.Errorf("line = %d, want %d", line, tt.wantLine)
			}
			if col != tt.wantCol {
				t.Errorf("col = %d, want %d", col, tt.wantCol)
			}
			if msg != tt.wantMsg {
				t.Errorf("msg = %q, want %q", msg, tt.wantMsg)
			}
		})
	}
}

func TestFormatGoError(t *testing.T) {
	// Create temp directory with a test source file
	tmpDir, err := os.MkdirTemp("", "rugby-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	genDir := filepath.Join(tmpDir, ".rugby", "gen")
	if err := os.MkdirAll(genDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a test source file
	sourceFile := filepath.Join(tmpDir, "test.rg")
	sourceContent := `def main
  x = 5
  y = 10
  puts x
end
`
	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
		t.Fatal(err)
	}

	project := &Project{Root: tmpDir, GenDir: genDir}
	b := &Builder{project: project}

	t.Run("formats error with source context", func(t *testing.T) {
		// Simulate Go error output with relative path from gen dir
		relPath, _ := filepath.Rel(genDir, sourceFile)
		goOutput := relPath + ":3: declared and not used: y"

		err := b.formatGoError(goOutput)
		errStr := err.Error()

		// Should contain formatted error parts
		if !contains(errStr, "error:") {
			t.Error("expected 'error:' in output")
		}
		if !contains(errStr, "declared and not used: y") {
			t.Error("expected error message in output")
		}
		if !contains(errStr, "y = 10") {
			t.Error("expected source line in output")
		}
	})

	t.Run("filters internal go paths", func(t *testing.T) {
		goOutput := `/usr/local/go/src/runtime/panic.go:1234: internal error
# command-line-arguments
../../test.rg:3: declared and not used: y`

		err := b.formatGoError(goOutput)
		errStr := err.Error()

		// Should not contain internal Go path
		if contains(errStr, "panic.go") {
			t.Error("should filter internal Go paths")
		}
		// Should not contain # line
		if contains(errStr, "# command") {
			t.Error("should filter # lines")
		}
		// Should still have the actual error
		if !contains(errStr, "declared and not used") {
			t.Error("should contain actual error")
		}
	})

	t.Run("handles empty output", func(t *testing.T) {
		goOutput := "# command-line-arguments"

		err := b.formatGoError(goOutput)
		errStr := err.Error()

		// Should return fallback with original output
		if !contains(errStr, "build failed") {
			t.Error("expected 'build failed' for empty filtered output")
		}
	})

	t.Run("handles non-rg file errors", func(t *testing.T) {
		goOutput := "main.go:10: some go error"

		err := b.formatGoError(goOutput)
		errStr := err.Error()

		// Should pass through non-rg errors as-is
		if !contains(errStr, "main.go:10: some go error") {
			t.Error("should pass through non-rg errors")
		}
	})

	t.Run("handles unreadable source file", func(t *testing.T) {
		goOutput := "../../nonexistent.rg:5: some error"

		err := b.formatGoError(goOutput)
		errStr := err.Error()

		// Should fallback to raw error
		if !contains(errStr, "nonexistent.rg") {
			t.Error("should contain filename in fallback")
		}
	})
}
