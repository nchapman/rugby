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
