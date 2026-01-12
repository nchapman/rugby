// Package regex provides regular expression operations for Rugby programs.
// Rugby: import rugby/regex
//
// Example:
//
//	if regex.match?(`\d+`, "abc123")
//	  puts "found digits"
//	end
//	matches = regex.find_all(`\w+`, "hello world")
package regex

import (
	"regexp"
)

// Match reports whether the string contains any match of the pattern.
// Ruby: regex.match?(pattern, str)
func Match(pattern, s string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// MustMatch reports whether the string contains any match of the pattern.
// Panics if the pattern is invalid.
// Ruby: regex.must_match?(pattern, str)
func MustMatch(pattern, s string) bool {
	return regexp.MustCompile(pattern).MatchString(s)
}

// Find returns the first match of the pattern in the string.
// Returns empty string and false if no match found.
// Ruby: regex.find(pattern, str)
func Find(pattern, s string) (string, bool) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", false
	}
	loc := re.FindStringIndex(s)
	if loc == nil {
		return "", false
	}
	return s[loc[0]:loc[1]], true
}

// FindAll returns all matches of the pattern in the string.
// Ruby: regex.find_all(pattern, str)
func FindAll(pattern, s string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re.FindAllString(s, -1)
}

// FindGroups returns the first match and its capture groups.
// The first element is the full match, followed by capture groups.
// Ruby: regex.find_groups(pattern, str)
func FindGroups(pattern, s string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re.FindStringSubmatch(s)
}

// FindAllGroups returns all matches with their capture groups.
// Ruby: regex.find_all_groups(pattern, str)
func FindAllGroups(pattern, s string) [][]string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re.FindAllStringSubmatch(s, -1)
}

// Replace replaces all matches of the pattern with the replacement.
// Ruby: regex.replace(str, pattern, replacement)
func Replace(s, pattern, replacement string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return s
	}
	return re.ReplaceAllString(s, replacement)
}

// ReplaceFirst replaces only the first match of the pattern.
// Supports backreferences like $1, $2 in the replacement string.
// Ruby: regex.replace_first(str, pattern, replacement)
func ReplaceFirst(s, pattern, replacement string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return s
	}
	loc := re.FindStringSubmatchIndex(s)
	if loc == nil {
		return s
	}
	var result []byte
	result = re.ExpandString(result, replacement, s, loc)
	return s[:loc[0]] + string(result) + s[loc[1]:]
}

// Split splits the string by the pattern.
// Ruby: regex.split(str, pattern)
func Split(s, pattern string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return []string{s}
	}
	return re.Split(s, -1)
}

// Compile compiles a pattern and returns a Regex object.
// Ruby: regex.compile(pattern)
func Compile(pattern string) (*Regex, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &Regex{re: re}, nil
}

// MustCompile compiles a pattern and panics if invalid.
// Ruby: regex.must_compile(pattern)
func MustCompile(pattern string) *Regex {
	return &Regex{re: regexp.MustCompile(pattern)}
}

// Regex is a compiled regular expression.
type Regex struct {
	re *regexp.Regexp
}

// Match reports whether the string contains a match.
// Ruby: re.match?(str)
func (r *Regex) Match(s string) bool {
	return r.re.MatchString(s)
}

// Find returns the first match.
// Ruby: re.find(str)
func (r *Regex) Find(s string) (string, bool) {
	loc := r.re.FindStringIndex(s)
	if loc == nil {
		return "", false
	}
	return s[loc[0]:loc[1]], true
}

// FindAll returns all matches.
// Ruby: re.find_all(str)
func (r *Regex) FindAll(s string) []string {
	return r.re.FindAllString(s, -1)
}

// FindGroups returns the first match and capture groups.
// Ruby: re.find_groups(str)
func (r *Regex) FindGroups(s string) []string {
	return r.re.FindStringSubmatch(s)
}

// FindAllGroups returns all matches with capture groups.
// Ruby: re.find_all_groups(str)
func (r *Regex) FindAllGroups(s string) [][]string {
	return r.re.FindAllStringSubmatch(s, -1)
}

// Replace replaces all matches with the replacement.
// Ruby: re.replace(str, replacement)
func (r *Regex) Replace(s, replacement string) string {
	return r.re.ReplaceAllString(s, replacement)
}

// ReplaceFirst replaces only the first match.
// Supports backreferences like $1, $2 in the replacement string.
// Ruby: re.replace_first(str, replacement)
func (r *Regex) ReplaceFirst(s, replacement string) string {
	loc := r.re.FindStringSubmatchIndex(s)
	if loc == nil {
		return s
	}
	var result []byte
	result = r.re.ExpandString(result, replacement, s, loc)
	return s[:loc[0]] + string(result) + s[loc[1]:]
}

// Split splits the string by the pattern.
// Ruby: re.split(str)
func (r *Regex) Split(s string) []string {
	return r.re.Split(s, -1)
}

// Pattern returns the pattern string.
// Ruby: re.pattern
func (r *Regex) Pattern() string {
	return r.re.String()
}

// Escape returns a string with all regex metacharacters escaped.
// Ruby: regex.escape(str)
func Escape(s string) string {
	return regexp.QuoteMeta(s)
}
