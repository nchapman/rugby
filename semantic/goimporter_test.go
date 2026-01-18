package semantic

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Empty string
		{"", ""},

		// Simple lowercase
		{"new", "new"},
		{"read", "read"},

		// PascalCase -> lowercase
		{"New", "new"},
		{"WriteString", "writestring"},
		{"NewScanner", "newscanner"},
		{"ReadAll", "readall"},

		// snake_case -> lowercase without underscores
		{"new_scanner", "newscanner"},
		{"read_all", "readall"},
		{"write_string", "writestring"},
		{"set_int64", "setint64"},

		// Acronyms normalize the same way
		{"UserID", "userid"},
		{"user_id", "userid"},
		{"ParseJSON", "parsejson"},
		{"parse_json", "parsejson"},
		{"ReadHTTPURL", "readhttpurl"},
		{"read_http_url", "readhttpurl"},
		{"GetAPIKey", "getapikey"},
		{"get_api_key", "getapikey"},

		// Ruby-style suffixes stripped
		{"empty?", "empty"},
		{"valid?", "valid"},
		{"save!", "save"},
		{"update!", "update"},
		{"is_valid?", "isvalid"},
		{"do_save!", "dosave"},

		// Mixed cases all normalize
		{"Int64", "int64"},
		{"int64", "int64"},
		{"SetInt64", "setint64"},
		{"set_int64", "setint64"},

		// Multiple underscores
		{"get_user_by_id", "getuserbyid"},
		{"GetUserByID", "getuserbyid"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalize(tt.input)
			if result != tt.expected {
				t.Errorf("normalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizedMatching verifies that Rugby names match Go names after normalization
func TestNormalizedMatching(t *testing.T) {
	tests := []struct {
		rubyName string
		goName   string
	}{
		// Simple cases
		{"new", "New"},
		{"read", "Read"},
		{"write", "Write"},

		// Multi-word
		{"write_string", "WriteString"},
		{"read_all", "ReadAll"},
		{"new_scanner", "NewScanner"},
		{"set_int64", "SetInt64"},

		// Acronyms
		{"user_id", "UserID"},
		{"parse_json", "ParseJSON"},
		{"read_http_url", "ReadHTTPURL"},
		{"get_api_key", "GetAPIKey"},
		{"new_sql_db", "NewSQLDB"},

		// Already matching
		{"New", "New"},
		{"WriteString", "WriteString"},
	}

	for _, tt := range tests {
		t.Run(tt.rubyName+"->"+tt.goName, func(t *testing.T) {
			rubyNorm := normalize(tt.rubyName)
			goNorm := normalize(tt.goName)
			if rubyNorm != goNorm {
				t.Errorf("normalize(%q) = %q, normalize(%q) = %q, they should match",
					tt.rubyName, rubyNorm, tt.goName, goNorm)
			}
		})
	}
}
