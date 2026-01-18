package semantic

import "testing"

func TestSnakeToPascal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Empty string
		{"", ""},

		// Single words - capitalize first letter
		{"new", "New"},
		{"read", "Read"},
		{"write", "Write"},

		// Single word acronyms - all uppercase
		{"id", "ID"},
		{"url", "URL"},
		{"http", "HTTP"},
		{"json", "JSON"},
		{"api", "API"},
		{"sql", "SQL"},
		{"cpu", "CPU"},
		{"io", "IO"},
		{"ip", "IP"},
		{"tcp", "TCP"},
		{"udp", "UDP"},
		{"ssl", "SSL"},
		{"tls", "TLS"},
		{"dns", "DNS"},
		{"ssh", "SSH"},
		{"uid", "UID"},
		{"pid", "PID"},
		{"uuid", "UUID"},
		{"eof", "EOF"},
		{"md5", "MD5"},
		{"sha", "SHA"},
		{"rsa", "RSA"},
		{"aes", "AES"},
		{"css", "CSS"},

		// Snake_case to PascalCase
		{"new_scanner", "NewScanner"},
		{"read_all", "ReadAll"},
		{"write_string", "WriteString"},
		{"set_int64", "SetInt64"},

		// Snake_case with acronyms
		{"user_id", "UserID"},
		{"parse_json", "ParseJSON"},
		{"read_http_url", "ReadHTTPURL"},
		{"get_api_key", "GetAPIKey"},
		{"new_sql_db", "NewSQLDb"},

		// Ruby-style suffixes stripped
		{"empty?", "Empty"},
		{"valid?", "Valid"},
		{"save!", "Save"},
		{"update!", "Update"},
		{"is_valid?", "IsValid"},
		{"do_save!", "DoSave"},

		// Already PascalCase - capitalize first letter only
		{"Scanner", "Scanner"},
		{"NewReader", "NewReader"},

		// Mixed cases
		{"readLine", "ReadLine"},

		// Multiple underscores
		{"get_user_by_id", "GetUserByID"},
		{"parse_http_url", "ParseHTTPURL"},

		// Leading/trailing underscores (edge cases)
		{"_private", "Private"},
		{"trailing_", "Trailing"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := snakeToPascal(tt.input)
			if result != tt.expected {
				t.Errorf("snakeToPascal(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
