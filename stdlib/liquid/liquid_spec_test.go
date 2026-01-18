package liquid

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestLiquidSpecs runs all .liquid spec tests
func TestLiquidSpecs(t *testing.T) {
	specDir := "../../tests/spec/stdlib/liquid"

	entries, err := os.ReadDir(specDir)
	if err != nil {
		t.Fatalf("failed to read spec directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".liquid") {
			continue
		}
		// Skip .liquid.out files
		if strings.HasSuffix(entry.Name(), ".liquid.out") {
			continue
		}

		testName := strings.TrimSuffix(entry.Name(), ".liquid")
		t.Run(testName, func(t *testing.T) {
			runLiquidSpec(t, filepath.Join(specDir, entry.Name()))
		})
	}
}

func runLiquidSpec(t *testing.T, templatePath string) {
	// Read template file
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("failed to read template: %v", err)
	}
	templateContent := string(templateBytes)

	// Extract data from {# data: {...} #} comment
	data := extractData(t, templateContent)

	// Remove the data directive from template
	template := removeDataDirective(templateContent)

	// Render
	result, err := Render(template, data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	// Read expected output
	expectedPath := templatePath + ".out"
	expectedBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected output %s: %v", expectedPath, err)
	}
	expected := string(expectedBytes)

	// Compare
	if result != expected {
		t.Errorf("output mismatch:\n--- expected ---\n%s\n--- actual ---\n%s", expected, result)
	}
}

var dataDirectiveRegex = regexp.MustCompile(`\{#\s*data:\s*(\{.*?\})\s*#\}`)

func extractData(t *testing.T, content string) map[string]any {
	matches := dataDirectiveRegex.FindStringSubmatch(content)
	if matches == nil {
		return make(map[string]any)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(matches[1]), &data); err != nil {
		t.Fatalf("failed to parse data directive: %v", err)
	}
	return data
}

func removeDataDirective(content string) string {
	// Remove the data directive line
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		if dataDirectiveRegex.MatchString(line) {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}
