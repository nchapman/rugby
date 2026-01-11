package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	w, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() { _ = w.Close() }()

	if w.root != tmpDir {
		t.Errorf("root = %q, want %q", w.root, tmpDir)
	}
	if w.debounce != 100*time.Millisecond {
		t.Errorf("debounce = %v, want %v", w.debounce, 100*time.Millisecond)
	}
}

func TestSetDebounce(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	w.SetDebounce(200 * time.Millisecond)
	if w.debounce != 200*time.Millisecond {
		t.Errorf("debounce = %v, want %v", w.debounce, 200*time.Millisecond)
	}
}

func TestAddPath_File(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.rg")
	if writeErr := os.WriteFile(testFile, []byte("test"), 0644); writeErr != nil {
		t.Fatal(writeErr)
	}

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	// Adding a file should watch its parent directory
	if addErr := w.AddPath(testFile); addErr != nil {
		t.Fatalf("AddPath(file) failed: %v", addErr)
	}

	if !w.watchedDirs[tmpDir] {
		t.Error("parent directory should be watched")
	}
}

func TestAddPath_Directory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "src")
	if mkdirErr := os.MkdirAll(subDir, 0755); mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	if addErr := w.AddPath(tmpDir); addErr != nil {
		t.Fatalf("AddPath(dir) failed: %v", addErr)
	}

	// Both directories should be watched
	if !w.watchedDirs[tmpDir] {
		t.Error("root directory should be watched")
	}
	if !w.watchedDirs[subDir] {
		t.Error("subdirectory should be watched")
	}
}

func TestAddPath_SkipsHiddenDirs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create hidden directory
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	if mkdirErr := os.MkdirAll(hiddenDir, 0755); mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	if addErr := w.AddPath(tmpDir); addErr != nil {
		t.Fatalf("AddPath failed: %v", addErr)
	}

	// Hidden directory should not be watched
	if w.watchedDirs[hiddenDir] {
		t.Error("hidden directory should not be watched")
	}
}

func TestIsRelevantFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"main.rg", true},
		{"src/app.rg", true},
		{"helper.go", false},
		{"main.go", false},
		{"readme.md", false},
		{"config.toml", false},
		{"data.json", false},
		{".gitignore", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isRelevantFile(tt.path)
			if got != tt.want {
				t.Errorf("isRelevantFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestWatch_DetectsFileChanges(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create initial file
	testFile := filepath.Join(tmpDir, "test.rg")
	if writeErr := os.WriteFile(testFile, []byte("initial"), 0644); writeErr != nil {
		t.Fatal(writeErr)
	}

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	// Use shorter debounce for tests
	w.SetDebounce(10 * time.Millisecond)

	if addErr := w.AddPath(tmpDir); addErr != nil {
		t.Fatal(addErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events, errors := w.Watch(ctx)

	// Modify the file after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = os.WriteFile(testFile, []byte("modified"), 0644)
	}()

	// Wait for event
	select {
	case event := <-events:
		if event.Path != testFile {
			t.Errorf("event.Path = %q, want %q", event.Path, testFile)
		}
	case watchErr := <-errors:
		t.Fatalf("unexpected error: %v", watchErr)
	case <-ctx.Done():
		t.Fatal("timeout waiting for file change event")
	}
}

func TestWatch_IgnoresNonRelevantFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	w.SetDebounce(10 * time.Millisecond)

	if addErr := w.AddPath(tmpDir); addErr != nil {
		t.Fatal(addErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	events, errors := w.Watch(ctx)

	// Create a non-relevant file
	go func() {
		time.Sleep(20 * time.Millisecond)
		_ = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("hi"), 0644)
	}()

	// Should not receive an event
	select {
	case event := <-events:
		t.Errorf("should not receive event for non-relevant file, got: %v", event)
	case watchErr := <-errors:
		t.Fatalf("unexpected error: %v", watchErr)
	case <-ctx.Done():
		// Expected - timeout with no event
	}
}

func TestWatch_Debounces(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.rg")
	if writeErr := os.WriteFile(testFile, []byte("initial"), 0644); writeErr != nil {
		t.Fatal(writeErr)
	}

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	// Use 50ms debounce
	w.SetDebounce(50 * time.Millisecond)

	if addErr := w.AddPath(tmpDir); addErr != nil {
		t.Fatal(addErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events, _ := w.Watch(ctx)

	// Make rapid changes
	go func() {
		time.Sleep(20 * time.Millisecond)
		for i := range 5 {
			_ = os.WriteFile(testFile, []byte{byte(i)}, 0644)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Count events received
	eventCount := 0
	timeout := time.After(300 * time.Millisecond)

loop:
	for {
		select {
		case _, ok := <-events:
			if !ok {
				break loop
			}
			eventCount++
		case <-timeout:
			break loop
		case <-ctx.Done():
			break loop
		}
	}

	// Should receive only 1-2 events due to debouncing, not 5
	if eventCount > 2 {
		t.Errorf("expected 1-2 debounced events, got %d", eventCount)
	}
}

func TestWatchDirRecursive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create nested structure
	nested := filepath.Join(tmpDir, "src", "lib")
	if mkdirErr := os.MkdirAll(nested, 0755); mkdirErr != nil {
		t.Fatal(mkdirErr)
	}

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close() }()

	if addErr := w.AddPath(tmpDir); addErr != nil {
		t.Fatal(addErr)
	}

	// All non-hidden directories should be watched
	if !w.watchedDirs[tmpDir] {
		t.Error("root should be watched")
	}
	if !w.watchedDirs[filepath.Join(tmpDir, "src")] {
		t.Error("src should be watched")
	}
	if !w.watchedDirs[nested] {
		t.Error("src/lib should be watched")
	}
}

func TestClose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	w, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if closeErr := w.Close(); closeErr != nil {
		t.Errorf("Close() failed: %v", closeErr)
	}
}

func TestClearScreen(t *testing.T) {
	// This is hard to test directly, but we can at least verify it doesn't panic
	ClearScreen()
}
