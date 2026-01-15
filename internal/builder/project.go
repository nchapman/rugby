package builder

import (
	"os"
	"path/filepath"
	"strings"
)

// Project represents a Rugby project and its directory structure.
type Project struct {
	Root     string // Project root (where go.mod lives)
	RugbyDir string // .rugby/ directory path
	GenDir   string // .rugby/gen/ for generated Go
	BinDir   string // .rugby/bin/ for compiled binaries
	CacheDir string // .rugby/cache/ for incremental builds
	DebugDir string // .rugby/debug/ for logs
}

// FindProject locates the project root by walking up from cwd looking for rugby.mod.
// If no rugby.mod is found, the current directory is used as root.
func FindProject() (*Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	root := findProjectRoot(cwd)
	return newProject(root), nil
}

// FindProjectFrom locates the project root starting from the given path.
func FindProjectFrom(startPath string) (*Project, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, err
	}

	// If it's a file, start from its directory
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		absPath = filepath.Dir(absPath)
	}

	root := findProjectRoot(absPath)
	return newProject(root), nil
}

func findProjectRoot(startDir string) string {
	dir := startDir
	for {
		// Look for rugby.mod (user's dependency file)
		rugbyModPath := filepath.Join(dir, "rugby.mod")
		if _, err := os.Stat(rugbyModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, use original directory
			return startDir
		}
		dir = parent
	}
}

func newProject(root string) *Project {
	rugbyDir := filepath.Join(root, ".rugby")
	return &Project{
		Root:     root,
		RugbyDir: rugbyDir,
		GenDir:   filepath.Join(rugbyDir, "gen"),
		BinDir:   filepath.Join(rugbyDir, "bin"),
		CacheDir: filepath.Join(rugbyDir, "cache"),
		DebugDir: filepath.Join(rugbyDir, "debug"),
	}
}

// EnsureDirs creates the .rugby directory structure.
func (p *Project) EnsureDirs() error {
	dirs := []string{p.GenDir, p.BinDir, p.CacheDir, p.DebugDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// GenPath maps a .rg source file to its generated .go path in .rugby/gen/.
// Example: src/main.rg -> .rugby/gen/src/main.go
// Note: Files ending in _test.rg become _test_.go to avoid Go treating them as test files.
func (p *Project) GenPath(rgFile string) string {
	// Get path relative to project root
	relPath, err := filepath.Rel(p.Root, rgFile)
	if err != nil {
		// Fall back to just the filename
		relPath = filepath.Base(rgFile)
	}

	// Change extension from .rg to .go
	baseName := strings.TrimSuffix(relPath, ".rg")

	// Avoid generating *_test.go files - Go treats these as test files
	// and excludes them from regular builds
	if strings.HasSuffix(baseName, "_test") {
		baseName = baseName + "_"
	}

	return filepath.Join(p.GenDir, baseName+".go")
}

// BinPath returns the path for a binary in .rugby/bin/.
func (p *Project) BinPath(name string) string {
	return filepath.Join(p.BinDir, name)
}

// Clean removes the entire .rugby directory.
func (p *Project) Clean() error {
	return os.RemoveAll(p.RugbyDir)
}

// RelPath returns the path relative to project root.
func (p *Project) RelPath(absPath string) string {
	rel, err := filepath.Rel(p.Root, absPath)
	if err != nil {
		return absPath
	}
	return rel
}

// RugbyModPath returns the path to rugby.mod in the project root.
func (p *Project) RugbyModPath() string {
	return filepath.Join(p.Root, "rugby.mod")
}
