package json

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantVal any
		wantErr bool
	}{
		{
			name:    "simple object",
			input:   `{"name": "Alice"}`,
			wantKey: "name",
			wantVal: "Alice",
		},
		{
			name:    "nested object",
			input:   `{"user": {"name": "Bob"}}`,
			wantKey: "user",
		},
		{
			name:    "number",
			input:   `{"age": 30}`,
			wantKey: "age",
			wantVal: float64(30),
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.wantVal != nil {
				if got[tt.wantKey] != tt.wantVal {
					t.Errorf("Parse()[%q] = %v, want %v", tt.wantKey, got[tt.wantKey], tt.wantVal)
				}
			}
		})
	}
}

func TestParseArray(t *testing.T) {
	input := `[1, 2, 3]`
	got, err := ParseArray(input)
	if err != nil {
		t.Fatalf("ParseArray() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("ParseArray() len = %d, want 3", len(got))
	}
	if got[0] != float64(1) {
		t.Errorf("ParseArray()[0] = %v, want 1", got[0])
	}
}

func TestGenerate(t *testing.T) {
	input := map[string]any{"name": "Alice", "age": float64(30)}
	got, err := Generate(input)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Parse it back to verify
	parsed, err := Parse(got)
	if err != nil {
		t.Fatalf("Parse(Generate()) error = %v", err)
	}
	if parsed["name"] != "Alice" {
		t.Errorf("roundtrip name = %v, want Alice", parsed["name"])
	}
}

func TestPretty(t *testing.T) {
	input := map[string]any{"name": "Alice"}
	got, err := Pretty(input)
	if err != nil {
		t.Fatalf("Pretty() error = %v", err)
	}
	// Pretty output should contain newlines
	if len(got) <= len(`{"name":"Alice"}`) {
		t.Errorf("Pretty() output too short, expected indented output")
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{`{"name": "Alice"}`, true},
		{`[1, 2, 3]`, true},
		{`"string"`, true},
		{`123`, true},
		{`{invalid}`, false},
		{``, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := Valid(tt.input); got != tt.want {
				t.Errorf("Valid(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
