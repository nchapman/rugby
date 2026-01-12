package regex

import (
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		pattern string
		s       string
		want    bool
	}{
		{`\d+`, "abc123", true},
		{`\d+`, "abc", false},
		{`^hello`, "hello world", true},
		{`^hello`, "say hello", false},
		{`world$`, "hello world", true},
	}

	for _, tt := range tests {
		got := Match(tt.pattern, tt.s)
		if got != tt.want {
			t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.s, got, tt.want)
		}
	}
}

func TestMatchInvalidPattern(t *testing.T) {
	// Invalid pattern should return false, not panic
	got := Match(`[invalid`, "test")
	if got {
		t.Error("Match with invalid pattern should return false")
	}
}

func TestMustMatch(t *testing.T) {
	if !MustMatch(`\d+`, "abc123") {
		t.Error("MustMatch should return true")
	}
}

func TestMustMatchPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustMatch with invalid pattern should panic")
		}
	}()
	MustMatch(`[invalid`, "test")
}

func TestFind(t *testing.T) {
	match, ok := Find(`\d+`, "abc123def456")
	if !ok {
		t.Error("Find() ok = false, want true")
	}
	if match != "123" {
		t.Errorf("Find() = %q, want %q", match, "123")
	}

	match, ok = Find(`\d+`, "abc")
	if ok {
		t.Error("Find() ok = true for no match, want false")
	}
	if match != "" {
		t.Errorf("Find() = %q, want empty string", match)
	}
}

func TestFindAll(t *testing.T) {
	matches := FindAll(`\d+`, "abc123def456ghi789")
	if len(matches) != 3 {
		t.Fatalf("FindAll() len = %d, want 3", len(matches))
	}
	expected := []string{"123", "456", "789"}
	for i, m := range matches {
		if m != expected[i] {
			t.Errorf("FindAll()[%d] = %q, want %q", i, m, expected[i])
		}
	}
}

func TestFindAllNoMatch(t *testing.T) {
	matches := FindAll(`\d+`, "abc")
	if matches != nil {
		t.Errorf("FindAll() = %v, want nil", matches)
	}
}

func TestFindGroups(t *testing.T) {
	groups := FindGroups(`(\w+)@(\w+)\.(\w+)`, "email: test@example.com")
	if len(groups) != 4 {
		t.Fatalf("FindGroups() len = %d, want 4", len(groups))
	}
	expected := []string{"test@example.com", "test", "example", "com"}
	for i, g := range groups {
		if g != expected[i] {
			t.Errorf("FindGroups()[%d] = %q, want %q", i, g, expected[i])
		}
	}
}

func TestFindAllGroups(t *testing.T) {
	groups := FindAllGroups(`(\w+)=(\d+)`, "a=1 b=2 c=3")
	if len(groups) != 3 {
		t.Fatalf("FindAllGroups() len = %d, want 3", len(groups))
	}
	// Each group should have [full, key, value]
	if groups[0][1] != "a" || groups[0][2] != "1" {
		t.Errorf("FindAllGroups()[0] = %v, want [a=1, a, 1]", groups[0])
	}
}

func TestReplace(t *testing.T) {
	result := Replace("hello 123 world 456", `\d+`, "X")
	if result != "hello X world X" {
		t.Errorf("Replace() = %q, want %q", result, "hello X world X")
	}
}

func TestReplaceWithBackreference(t *testing.T) {
	result := Replace("hello world", `(\w+) (\w+)`, "$2 $1")
	if result != "world hello" {
		t.Errorf("Replace() = %q, want %q", result, "world hello")
	}
}

func TestReplaceFirst(t *testing.T) {
	result := ReplaceFirst("hello 123 world 456", `\d+`, "X")
	if result != "hello X world 456" {
		t.Errorf("ReplaceFirst() = %q, want %q", result, "hello X world 456")
	}
}

func TestReplaceFirstNoMatch(t *testing.T) {
	result := ReplaceFirst("hello world", `\d+`, "X")
	if result != "hello world" {
		t.Errorf("ReplaceFirst() = %q, want %q", result, "hello world")
	}
}

func TestReplaceFirstWithBackreference(t *testing.T) {
	result := ReplaceFirst("hello world foo bar", `(\w+) (\w+)`, "$2-$1")
	if result != "world-hello foo bar" {
		t.Errorf("ReplaceFirst() = %q, want %q", result, "world-hello foo bar")
	}
}

func TestSplit(t *testing.T) {
	parts := Split("a1b2c3d", `\d`)
	if len(parts) != 4 {
		t.Fatalf("Split() len = %d, want 4", len(parts))
	}
	expected := []string{"a", "b", "c", "d"}
	for i, p := range parts {
		if p != expected[i] {
			t.Errorf("Split()[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestSplitInvalidPattern(t *testing.T) {
	parts := Split("test", `[invalid`)
	if len(parts) != 1 || parts[0] != "test" {
		t.Errorf("Split() with invalid pattern = %v, want [test]", parts)
	}
}

func TestCompile(t *testing.T) {
	re, err := Compile(`\d+`)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if !re.Match("abc123") {
		t.Error("re.Match() = false, want true")
	}
}

func TestCompileInvalid(t *testing.T) {
	_, err := Compile(`[invalid`)
	if err == nil {
		t.Error("Compile() with invalid pattern should return error")
	}
}

func TestMustCompile(t *testing.T) {
	re := MustCompile(`\d+`)
	if !re.Match("abc123") {
		t.Error("re.Match() = false, want true")
	}
}

func TestMustCompilePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustCompile with invalid pattern should panic")
		}
	}()
	MustCompile(`[invalid`)
}

func TestRegexMethods(t *testing.T) {
	re := MustCompile(`(\w+)`)

	// Match
	if !re.Match("hello") {
		t.Error("re.Match() = false")
	}

	// Find
	match, ok := re.Find("hello world")
	if !ok || match != "hello" {
		t.Errorf("re.Find() = %q, %v, want hello, true", match, ok)
	}

	// FindAll
	matches := re.FindAll("hello world")
	if len(matches) != 2 {
		t.Errorf("re.FindAll() len = %d, want 2", len(matches))
	}

	// FindGroups
	groups := re.FindGroups("hello")
	if len(groups) != 2 || groups[1] != "hello" {
		t.Errorf("re.FindGroups() = %v, want [hello, hello]", groups)
	}

	// Replace
	result := re.Replace("hello world", "X")
	if result != "X X" {
		t.Errorf("re.Replace() = %q, want %q", result, "X X")
	}

	// ReplaceFirst
	result = re.ReplaceFirst("hello world", "X")
	if result != "X world" {
		t.Errorf("re.ReplaceFirst() = %q, want %q", result, "X world")
	}

	// Split - use a different regex
	reDigit := MustCompile(`\d`)
	parts := reDigit.Split("a1b2c")
	if len(parts) != 3 {
		t.Errorf("re.Split() len = %d, want 3", len(parts))
	}

	// Pattern
	if re.Pattern() != `(\w+)` {
		t.Errorf("re.Pattern() = %q, want %q", re.Pattern(), `(\w+)`)
	}
}

func TestEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"hello.world", `hello\.world`},
		{"a+b*c?", `a\+b\*c\?`},
		{"[test]", `\[test\]`},
		{"$100", `\$100`},
	}

	for _, tt := range tests {
		got := Escape(tt.input)
		if got != tt.want {
			t.Errorf("Escape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
