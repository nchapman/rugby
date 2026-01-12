package csv

import (
	"io"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    [][]string
		wantErr bool
	}{
		{
			name:  "simple",
			input: "a,b,c\n1,2,3",
			want:  [][]string{{"a", "b", "c"}, {"1", "2", "3"}},
		},
		{
			name:  "quoted fields",
			input: `"hello, world",b,c`,
			want:  [][]string{{"hello, world", "b", "c"}},
		},
		{
			name:  "empty fields",
			input: "a,,c",
			want:  [][]string{{"a", "", "c"}},
		},
		{
			name:  "newline in quoted",
			input: "\"line1\nline2\",b",
			want:  [][]string{{"line1\nline2", "b"}},
		},
		{
			name:  "escaped quotes",
			input: `"say ""hello""",b`,
			want:  [][]string{{`say "hello"`, "b"}},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBytes(t *testing.T) {
	input := []byte("a,b,c\n1,2,3")
	got, err := ParseBytes(input)
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}
	want := [][]string{{"a", "b", "c"}, {"1", "2", "3"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseBytes() = %v, want %v", got, want)
	}
}

func TestParseWithHeaders(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []map[string]string
		wantErr bool
	}{
		{
			name:  "simple",
			input: "name,age\nAlice,30\nBob,25",
			want: []map[string]string{
				{"name": "Alice", "age": "30"},
				{"name": "Bob", "age": "25"},
			},
		},
		{
			name:  "extra header",
			input: "name,age,city\nAlice,30",
			want: []map[string]string{
				{"name": "Alice", "age": "30", "city": ""},
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:  "headers only",
			input: "name,age",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWithHeaders(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWithHeaders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseWithHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name  string
		input [][]string
		want  string
	}{
		{
			name:  "simple",
			input: [][]string{{"a", "b", "c"}, {"1", "2", "3"}},
			want:  "a,b,c\n1,2,3\n",
		},
		{
			name:  "needs quoting",
			input: [][]string{{"hello, world", "b"}},
			want:  "\"hello, world\",b\n",
		},
		{
			name:  "empty",
			input: [][]string{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Generate(tt.input)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Generate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateBytes(t *testing.T) {
	input := [][]string{{"a", "b"}, {"1", "2"}}
	got, err := GenerateBytes(input)
	if err != nil {
		t.Fatalf("GenerateBytes() error = %v", err)
	}
	want := []byte("a,b\n1,2\n")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GenerateBytes() = %q, want %q", got, want)
	}
}

func TestGenerateWithHeaders(t *testing.T) {
	headers := []string{"name", "age"}
	records := []map[string]string{
		{"name": "Alice", "age": "30"},
		{"name": "Bob", "age": "25"},
	}

	got, err := GenerateWithHeaders(headers, records)
	if err != nil {
		t.Fatalf("GenerateWithHeaders() error = %v", err)
	}

	want := "name,age\nAlice,30\nBob,25\n"
	if got != want {
		t.Errorf("GenerateWithHeaders() = %q, want %q", got, want)
	}
}

func TestRoundtrip(t *testing.T) {
	original := [][]string{
		{"name", "description", "price"},
		{"Widget", "A useful widget", "9.99"},
		{"Gadget", "Contains, commas", "19.99"},
	}

	csv, err := Generate(original)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	parsed, err := Parse(csv)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !reflect.DeepEqual(parsed, original) {
		t.Errorf("Roundtrip failed: got %v, want %v", parsed, original)
	}
}

func TestReader(t *testing.T) {
	input := "a,b,c\n1,2,3\n4,5,6"
	r := NewReader(input)

	// Read first row
	row1, err := r.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if !reflect.DeepEqual(row1, []string{"a", "b", "c"}) {
		t.Errorf("Read() = %v, want [a b c]", row1)
	}

	// Read remaining rows
	rest, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(rest) != 2 {
		t.Errorf("ReadAll() len = %d, want 2", len(rest))
	}
}

func TestReaderDelimiter(t *testing.T) {
	input := "a;b;c"
	r := NewReader(input)
	r.SetDelimiter(';')

	row, err := r.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if !reflect.DeepEqual(row, []string{"a", "b", "c"}) {
		t.Errorf("Read() = %v, want [a b c]", row)
	}
}

func TestReaderComment(t *testing.T) {
	input := "# this is a comment\na,b,c"
	r := NewReader(input)
	r.SetComment('#')

	row, err := r.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if !reflect.DeepEqual(row, []string{"a", "b", "c"}) {
		t.Errorf("Read() = %v, want [a b c]", row)
	}
}

func TestWriter(t *testing.T) {
	w := NewWriter()
	w.Write([]string{"a", "b", "c"})
	w.Write([]string{"1", "2", "3"})

	got := w.String()
	want := "a,b,c\n1,2,3\n"
	if got != want {
		t.Errorf("Writer.String() = %q, want %q", got, want)
	}
}

func TestWriterDelimiter(t *testing.T) {
	w := NewWriter()
	w.SetDelimiter(';')
	w.Write([]string{"a", "b", "c"})

	got := w.String()
	want := "a;b;c\n"
	if got != want {
		t.Errorf("Writer.String() = %q, want %q", got, want)
	}
}

func TestWriterWriteAll(t *testing.T) {
	w := NewWriter()
	w.WriteAll([][]string{{"a", "b"}, {"1", "2"}})

	got := w.String()
	want := "a,b\n1,2\n"
	if got != want {
		t.Errorf("Writer.String() = %q, want %q", got, want)
	}
}

func TestWriterBytes(t *testing.T) {
	w := NewWriter()
	w.Write([]string{"a", "b"})

	got := w.Bytes()
	want := []byte("a,b\n")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Writer.Bytes() = %q, want %q", got, want)
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{`"hello, world",b`, []string{"hello, world", "b"}},
	}

	for _, tt := range tests {
		got, err := ParseLine(tt.input)
		if err != nil && err != io.EOF {
			t.Errorf("ParseLine(%q) error = %v", tt.input, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("ParseLine(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"hello, world", `"hello, world"`},
		{`say "hello"`, `"say ""hello"""`},
		{"line1\nline2", "\"line1\nline2\""},
	}

	for _, tt := range tests {
		got := Escape(tt.input)
		if got != tt.want {
			t.Errorf("Escape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"a,b,c", true},
		{`"quoted"`, true},
		{"", true},
		{`"unclosed`, false},
	}

	for _, tt := range tests {
		got := Valid(tt.input)
		if got != tt.want {
			t.Errorf("Valid(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRowCount(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"a,b,c\n1,2,3", 2},
		{"a,b,c", 1},
		{"", 0},
		{`"unclosed`, -1},
	}

	for _, tt := range tests {
		got := RowCount(tt.input)
		if got != tt.want {
			t.Errorf("RowCount(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestHeaders(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "simple",
			input: "name,age,city\nAlice,30,NYC",
			want:  []string{"name", "age", "city"},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Headers(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Headers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Headers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRow(t *testing.T) {
	row := NewRow("a", "b", "c")
	want := Row{"a", "b", "c"}
	if !reflect.DeepEqual(row, want) {
		t.Errorf("NewRow() = %v, want %v", row, want)
	}
}

func TestParseBytesWithHeaders(t *testing.T) {
	input := []byte("name,age\nAlice,30")
	got, err := ParseBytesWithHeaders(input)
	if err != nil {
		t.Fatalf("ParseBytesWithHeaders() error = %v", err)
	}
	want := []map[string]string{{"name": "Alice", "age": "30"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseBytesWithHeaders() = %v, want %v", got, want)
	}
}

func TestReaderLazyQuotes(t *testing.T) {
	input := `a"b,c`
	r := NewReader(input)
	r.SetLazyQuotes(true)

	row, err := r.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if row[0] != `a"b` {
		t.Errorf("got %q, want %q", row[0], `a"b`)
	}
}

func TestReaderTrimLeadingSpace(t *testing.T) {
	input := "  a,  b"
	r := NewReader(input)
	r.SetTrimLeadingSpace(true)

	row, err := r.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	want := []string{"a", "b"}
	if !reflect.DeepEqual(row, want) {
		t.Errorf("got %v, want %v", row, want)
	}
}

func TestWriterUseCRLF(t *testing.T) {
	w := NewWriter()
	w.SetUseCRLF(true)
	if err := w.Write([]string{"a", "b"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := w.String()
	want := "a,b\r\n"
	if got != want {
		t.Errorf("Writer.String() = %q, want %q", got, want)
	}
}
