package path

import (
	"runtime"
	"strings"
	"testing"
)

func TestJoin(t *testing.T) {
	tests := []struct {
		parts []string
		want  string
	}{
		{[]string{"a", "b", "c"}, "a/b/c"},
		{[]string{"/a", "b", "c"}, "/a/b/c"},
		{[]string{"a", "b/c"}, "a/b/c"},
		{[]string{"a", "", "c"}, "a/c"},
		{[]string{}, ""},
	}

	for _, tt := range tests {
		got := Join(tt.parts...)
		// Normalize for Windows
		want := tt.want
		if runtime.GOOS == "windows" {
			want = windowsPath(want)
		}
		if got != want {
			t.Errorf("Join(%v) = %q, want %q", tt.parts, got, want)
		}
	}
}

func TestSplit(t *testing.T) {
	dir, file := Split("/a/b/c.txt")
	if dir != "/a/b/" {
		t.Errorf("Split() dir = %q, want %q", dir, "/a/b/")
	}
	if file != "c.txt" {
		t.Errorf("Split() file = %q, want %q", file, "c.txt")
	}
}

func TestDirname(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/a/b/c.txt", "/a/b"},
		{"/a/b/c", "/a/b"},
		{"c.txt", "."},
		{"/", "/"},
	}

	for _, tt := range tests {
		got := Dirname(tt.path)
		if got != tt.want {
			t.Errorf("Dirname(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestBasename(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/a/b/c.txt", "c.txt"},
		{"/a/b/c", "c"},
		{"c.txt", "c.txt"},
		{"/", "/"},
	}

	for _, tt := range tests {
		got := Basename(tt.path)
		if got != tt.want {
			t.Errorf("Basename(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestExtname(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"file.txt", ".txt"},
		{"file.tar.gz", ".gz"},
		{"file", ""},
		{".gitignore", ".gitignore"}, // Go treats the whole name as extension
		{"a/b/file.txt", ".txt"},
	}

	for _, tt := range tests {
		got := Extname(tt.path)
		if got != tt.want {
			t.Errorf("Extname(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestAbsolute(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/a/b/c", true},
		{"a/b/c", false},
		{"./a/b", false},
		{"../a/b", false},
	}

	// Adjust for Windows where absolute paths start with drive letter
	if runtime.GOOS == "windows" {
		tests = []struct {
			path string
			want bool
		}{
			{"C:\\a\\b\\c", true},
			{"a\\b\\c", false},
			{".\\a\\b", false},
			{"..\\a\\b", false},
		}
	}

	for _, tt := range tests {
		got := Absolute(tt.path)
		if got != tt.want {
			t.Errorf("Absolute(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestRelative(t *testing.T) {
	rel, err := Relative("/a/b", "/a/b/c/d")
	if err != nil {
		t.Fatalf("Relative() error = %v", err)
	}
	if rel != "c/d" {
		t.Errorf("Relative() = %q, want %q", rel, "c/d")
	}

	rel, err = Relative("/a/b/c", "/a/b")
	if err != nil {
		t.Fatalf("Relative() error = %v", err)
	}
	if rel != ".." {
		t.Errorf("Relative() = %q, want %q", rel, "..")
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"a/b/../c", "a/c"},
		{"a/./b/c", "a/b/c"},
		{"a//b//c", "a/b/c"},
		{"./a/b", "a/b"},
	}

	for _, tt := range tests {
		got := Clean(tt.path)
		if got != tt.want {
			t.Errorf("Clean(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestAbs(t *testing.T) {
	abs, err := Abs(".")
	if err != nil {
		t.Fatalf("Abs() error = %v", err)
	}
	if !Absolute(abs) {
		t.Errorf("Abs(\".\") = %q is not absolute", abs)
	}
}

func TestStem(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"file.txt", "file"},
		{"file.tar.gz", "file.tar"},
		{"file", "file"},
		{".gitignore", ""}, // Go treats .gitignore as extension only
		{"a/b/file.txt", "file"},
	}

	for _, tt := range tests {
		got := Stem(tt.path)
		if got != tt.want {
			t.Errorf("Stem(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"*.txt", "file.txt", true},
		{"*.txt", "file.go", false},
		{"file.*", "file.txt", true},
		{"file?.txt", "file1.txt", true},
		{"file?.txt", "file12.txt", false},
	}

	for _, tt := range tests {
		got := Match(tt.pattern, tt.name)
		if got != tt.want {
			t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
		}
	}
}

// windowsPath converts a Unix-style path to Windows-style for testing.
func windowsPath(p string) string {
	return strings.ReplaceAll(p, "/", "\\")
}
