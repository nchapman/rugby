package url

import (
	"testing"
)

func TestParse(t *testing.T) {
	u, err := Parse("https://user:pass@example.com:8080/path?q=search#fragment")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if u.Scheme() != "https" {
		t.Errorf("Scheme() = %q, want %q", u.Scheme(), "https")
	}
	if u.Host() != "example.com:8080" {
		t.Errorf("Host() = %q, want %q", u.Host(), "example.com:8080")
	}
	if u.Hostname() != "example.com" {
		t.Errorf("Hostname() = %q, want %q", u.Hostname(), "example.com")
	}
	if u.Port() != "8080" {
		t.Errorf("Port() = %q, want %q", u.Port(), "8080")
	}
	if u.Path() != "/path" {
		t.Errorf("Path() = %q, want %q", u.Path(), "/path")
	}
	if u.RawQuery() != "q=search" {
		t.Errorf("RawQuery() = %q, want %q", u.RawQuery(), "q=search")
	}
	if u.Fragment() != "fragment" {
		t.Errorf("Fragment() = %q, want %q", u.Fragment(), "fragment")
	}
	if u.User() != "user" {
		t.Errorf("User() = %q, want %q", u.User(), "user")
	}
	pass, ok := u.Password()
	if !ok || pass != "pass" {
		t.Errorf("Password() = %q, %v, want %q, true", pass, ok, "pass")
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse("://invalid")
	if err == nil {
		t.Error("Parse() should return error for invalid URL")
	}
}

func TestMustParse(t *testing.T) {
	u := MustParse("https://example.com")
	if u.Host() != "example.com" {
		t.Errorf("MustParse() Host() = %q, want %q", u.Host(), "example.com")
	}
}

func TestMustParsePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse() should panic on invalid URL")
		}
	}()
	MustParse("://invalid")
}

func TestURLString(t *testing.T) {
	u := MustParse("https://example.com/path")
	if u.String() != "https://example.com/path" {
		t.Errorf("String() = %q, want %q", u.String(), "https://example.com/path")
	}
}

func TestQuery(t *testing.T) {
	u := MustParse("https://example.com?a=1&b=2&a=3")
	q := u.Query()

	// Query returns first value for each key
	if q["a"] != "1" {
		t.Errorf("Query()[a] = %q, want %q", q["a"], "1")
	}
	if q["b"] != "2" {
		t.Errorf("Query()[b] = %q, want %q", q["b"], "2")
	}
}

func TestQueryValues(t *testing.T) {
	u := MustParse("https://example.com?a=1&a=2&b=3")
	q := u.QueryValues()

	if len(q["a"]) != 2 {
		t.Errorf("QueryValues()[a] length = %d, want 2", len(q["a"]))
	}
	if q["a"][0] != "1" || q["a"][1] != "2" {
		t.Errorf("QueryValues()[a] = %v, want [1 2]", q["a"])
	}
}

func TestIsAbsolute(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"/path/to/resource", false},
		{"relative/path", false},
	}

	for _, tt := range tests {
		u := MustParse(tt.url)
		if u.IsAbsolute() != tt.want {
			t.Errorf("IsAbsolute(%q) = %v, want %v", tt.url, u.IsAbsolute(), tt.want)
		}
	}
}

func TestWithScheme(t *testing.T) {
	u := MustParse("http://example.com")
	u2 := u.WithScheme("https")
	if u2.Scheme() != "https" {
		t.Errorf("WithScheme() = %q, want %q", u2.Scheme(), "https")
	}
	// Original unchanged
	if u.Scheme() != "http" {
		t.Error("WithScheme() should not modify original")
	}
}

func TestWithHost(t *testing.T) {
	u := MustParse("https://old.com/path")
	u2 := u.WithHost("new.com")
	if u2.Host() != "new.com" {
		t.Errorf("WithHost() = %q, want %q", u2.Host(), "new.com")
	}
	if u2.Path() != "/path" {
		t.Errorf("WithHost() should preserve path, got %q", u2.Path())
	}
}

func TestWithPath(t *testing.T) {
	u := MustParse("https://example.com/old")
	u2 := u.WithPath("/new/path")
	if u2.Path() != "/new/path" {
		t.Errorf("WithPath() = %q, want %q", u2.Path(), "/new/path")
	}
}

func TestWithQuery(t *testing.T) {
	u := MustParse("https://example.com/path?old=1")
	u2 := u.WithQuery(map[string]string{"new": "2"})
	if u2.Query()["new"] != "2" {
		t.Errorf("WithQuery() new param = %q, want %q", u2.Query()["new"], "2")
	}
	if _, ok := u2.Query()["old"]; ok {
		t.Error("WithQuery() should replace old query params")
	}
}

func TestWithFragment(t *testing.T) {
	u := MustParse("https://example.com/path")
	u2 := u.WithFragment("section")
	if u2.Fragment() != "section" {
		t.Errorf("WithFragment() = %q, want %q", u2.Fragment(), "section")
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		base string
		ref  string
		want string
	}{
		{"https://example.com/a/b", "/c", "https://example.com/c"},
		{"https://example.com/a/b", "c", "https://example.com/a/c"},
		{"https://example.com/a/b/", "c", "https://example.com/a/b/c"},
		{"https://example.com", "https://other.com", "https://other.com"},
	}

	for _, tt := range tests {
		u := MustParse(tt.base)
		resolved, err := u.Resolve(tt.ref)
		if err != nil {
			t.Fatalf("Resolve(%q) error = %v", tt.ref, err)
		}
		if resolved.String() != tt.want {
			t.Errorf("Resolve(%q, %q) = %q, want %q", tt.base, tt.ref, resolved.String(), tt.want)
		}
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		base  string
		paths []string
		want  string
	}{
		{"https://example.com", []string{"a", "b"}, "https://example.com/a/b"},
		{"https://example.com/", []string{"a", "b"}, "https://example.com/a/b"},
		{"https://example.com/base", []string{"path"}, "https://example.com/base/path"},
	}

	for _, tt := range tests {
		got, err := Join(tt.base, tt.paths...)
		if err != nil {
			t.Fatalf("Join() error = %v", err)
		}
		if got != tt.want {
			t.Errorf("Join(%q, %v) = %q, want %q", tt.base, tt.paths, got, tt.want)
		}
	}
}

func TestMustJoin(t *testing.T) {
	got := MustJoin("https://example.com", "a", "b")
	want := "https://example.com/a/b"
	if got != want {
		t.Errorf("MustJoin() = %q, want %q", got, want)
	}
}

func TestMustJoinPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustJoin() should panic on invalid URL")
		}
	}()
	MustJoin("://invalid", "path")
}

func TestQueryEncode(t *testing.T) {
	// Note: Go's url.Values.Encode() sorts keys alphabetically
	got := QueryEncode(map[string]string{"b": "2", "a": "1"})
	if got != "a=1&b=2" {
		t.Errorf("QueryEncode() = %q, want %q", got, "a=1&b=2")
	}
}

func TestQueryEncodeSpecialChars(t *testing.T) {
	got := QueryEncode(map[string]string{"q": "hello world"})
	if got != "q=hello+world" {
		t.Errorf("QueryEncode() = %q, want %q", got, "q=hello+world")
	}
}

func TestQueryEncodeValues(t *testing.T) {
	got := QueryEncodeValues(map[string][]string{"a": {"1", "2"}})
	if got != "a=1&a=2" {
		t.Errorf("QueryEncodeValues() = %q, want %q", got, "a=1&a=2")
	}
}

func TestQueryDecode(t *testing.T) {
	got, err := QueryDecode("a=1&b=2")
	if err != nil {
		t.Fatalf("QueryDecode() error = %v", err)
	}
	if got["a"] != "1" || got["b"] != "2" {
		t.Errorf("QueryDecode() = %v, want map[a:1 b:2]", got)
	}
}

func TestQueryDecodeSpecialChars(t *testing.T) {
	got, err := QueryDecode("q=hello+world")
	if err != nil {
		t.Fatalf("QueryDecode() error = %v", err)
	}
	if got["q"] != "hello world" {
		t.Errorf("QueryDecode() = %q, want %q", got["q"], "hello world")
	}
}

func TestQueryDecodeValues(t *testing.T) {
	got, err := QueryDecodeValues("a=1&a=2&b=3")
	if err != nil {
		t.Fatalf("QueryDecodeValues() error = %v", err)
	}
	if len(got["a"]) != 2 {
		t.Errorf("QueryDecodeValues()[a] length = %d, want 2", len(got["a"]))
	}
}

func TestEncodeDecode(t *testing.T) {
	tests := []string{
		"hello world",
		"a=b&c=d",
		"special!@#$%^&*()",
	}

	for _, tt := range tests {
		encoded := Encode(tt)
		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if decoded != tt {
			t.Errorf("roundtrip(%q) = %q", tt, decoded)
		}
	}
}

func TestPathEncodeDecode(t *testing.T) {
	tests := []string{
		"hello world",
		"path/with/slashes",
		"file name.txt",
	}

	for _, tt := range tests {
		encoded := PathEncode(tt)
		decoded, err := PathDecode(encoded)
		if err != nil {
			t.Fatalf("PathDecode() error = %v", err)
		}
		if decoded != tt {
			t.Errorf("path roundtrip(%q) = %q", tt, decoded)
		}
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://example.com", true},
		{"http://localhost:8080", true},
		{"/path/to/resource", true},
		{"", true}, // empty is valid in Go's url.Parse
	}

	for _, tt := range tests {
		if Valid(tt.url) != tt.want {
			t.Errorf("Valid(%q) = %v, want %v", tt.url, Valid(tt.url), tt.want)
		}
	}
}

func TestIsHTTP(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"HTTP://EXAMPLE.COM", true},
		{"ftp://example.com", false},
		{"file:///path", false},
		{"/path", false},
	}

	for _, tt := range tests {
		if IsHTTP(tt.url) != tt.want {
			t.Errorf("IsHTTP(%q) = %v, want %v", tt.url, IsHTTP(tt.url), tt.want)
		}
	}
}

func TestPasswordNotSet(t *testing.T) {
	u := MustParse("https://user@example.com")
	pass, ok := u.Password()
	if ok {
		t.Errorf("Password() ok = true, want false")
	}
	if pass != "" {
		t.Errorf("Password() = %q, want empty", pass)
	}
}

func TestNoUserInfo(t *testing.T) {
	u := MustParse("https://example.com")
	if u.User() != "" {
		t.Errorf("User() = %q, want empty", u.User())
	}
	pass, ok := u.Password()
	if ok || pass != "" {
		t.Errorf("Password() = %q, %v, want empty, false", pass, ok)
	}
}
