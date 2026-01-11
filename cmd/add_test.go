package cmd

import "testing"

func TestPackageExists(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		pkg      string
		expected bool
	}{
		{
			name:     "exact match in require block",
			content:  "module test\n\nrequire (\n\tgithub.com/foo/bar v1.0.0\n)",
			pkg:      "github.com/foo/bar",
			expected: true,
		},
		{
			name:     "no match",
			content:  "module test\n\nrequire (\n\tgithub.com/foo/bar v1.0.0\n)",
			pkg:      "github.com/other/pkg",
			expected: false,
		},
		{
			name:     "substring should not match",
			content:  "module test\n\nrequire (\n\tgithub.com/foo/bar v1.0.0\n)",
			pkg:      "github.com/foo/bar-extra",
			expected: false,
		},
		{
			name:     "prefix should not match",
			content:  "module test\n\nrequire (\n\tgithub.com/foo/bar-extra v1.0.0\n)",
			pkg:      "github.com/foo/bar",
			expected: false,
		},
		{
			name:     "single line require",
			content:  "module test\n\nrequire github.com/foo/bar v1.0.0",
			pkg:      "github.com/foo/bar",
			expected: true,
		},
		{
			name:     "package without version",
			content:  "module test\n\nrequire (\n\tgithub.com/foo/bar\n)",
			pkg:      "github.com/foo/bar",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := packageExists(tt.content, tt.pkg)
			if result != tt.expected {
				t.Errorf("packageExists(%q, %q) = %v, want %v", tt.content, tt.pkg, result, tt.expected)
			}
		})
	}
}

func TestAddDependency(t *testing.T) {
	tests := []struct {
		name              string
		content           string
		pkg               string
		version           string
		expectContains    []string
		expectNotContains []string
	}{
		{
			name:           "add to empty require block",
			content:        "module test\n\ngo 1.25\n\nrequire (\n)",
			pkg:            "github.com/foo/bar",
			version:        "",
			expectContains: []string{"github.com/foo/bar"},
		},
		{
			name:           "add with version",
			content:        "module test\n\ngo 1.25\n\nrequire (\n)",
			pkg:            "github.com/foo/bar",
			version:        "1.0.0",
			expectContains: []string{"github.com/foo/bar v1.0.0"},
		},
		{
			name:           "add with v-prefixed version",
			content:        "module test\n\ngo 1.25\n\nrequire (\n)",
			pkg:            "github.com/foo/bar",
			version:        "v1.0.0",
			expectContains: []string{"github.com/foo/bar v1.0.0"},
		},
		{
			name:              "latest means no version",
			content:           "module test\n\ngo 1.25\n\nrequire (\n)",
			pkg:               "github.com/foo/bar",
			version:           "latest",
			expectContains:    []string{"github.com/foo/bar"},
			expectNotContains: []string{"latest"},
		},
		{
			name:           "create require block if missing",
			content:        "module test\n\ngo 1.25\n",
			pkg:            "github.com/foo/bar",
			version:        "",
			expectContains: []string{"require (", "github.com/foo/bar", ")"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addDependency(tt.content, tt.pkg, tt.version)
			for _, s := range tt.expectContains {
				if !containsString(result, s) {
					t.Errorf("addDependency result should contain %q, got:\n%s", s, result)
				}
			}
			for _, s := range tt.expectNotContains {
				if containsString(result, s) {
					t.Errorf("addDependency result should not contain %q, got:\n%s", s, result)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
