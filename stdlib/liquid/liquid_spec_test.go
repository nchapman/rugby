package liquid

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func extractData(t *testing.T, content string) map[string]any {
	// Try to find {# data: ... #} which may span multiple lines
	start := strings.Index(content, "{# data:")
	if start == -1 {
		return make(map[string]any)
	}

	// Find the closing #}
	end := strings.Index(content[start:], "#}")
	if end == -1 {
		return make(map[string]any)
	}
	end += start + 2 // include the #}

	// Extract just the JSON part
	directive := content[start:end]
	jsonStart := strings.Index(directive, "{# data:") + len("{# data:")
	jsonEnd := strings.LastIndex(directive, "#}")

	jsonStr := strings.TrimSpace(directive[jsonStart:jsonEnd])

	var data map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("failed to parse data directive: %v\nJSON: %s", err, jsonStr)
	}
	return data
}

func removeDataDirective(content string) string {
	// Find and remove {# data: ... #} which may span multiple lines
	start := strings.Index(content, "{# data:")
	if start == -1 {
		return content
	}

	end := strings.Index(content[start:], "#}")
	if end == -1 {
		return content
	}
	end += start + 2 // include the #}

	// Also remove the trailing newline if present
	if end < len(content) && content[end] == '\n' {
		end++
	}

	return content[:start] + content[end:]
}
