// Package csv provides CSV parsing and generation for Rugby programs.
// Rugby: import rugby/csv
//
// Example:
//
//	rows = csv.parse(str)!
//	records = csv.parse_with_headers(str)!
//	output = csv.generate(rows)!
package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"strings"
)

// Parse parses a CSV string into rows of fields.
// Ruby: csv.parse(str)
func Parse(s string) ([][]string, error) {
	r := csv.NewReader(strings.NewReader(s))
	return r.ReadAll()
}

// ParseBytes parses CSV bytes into rows of fields.
// Ruby: csv.parse_bytes(bytes)
func ParseBytes(b []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(b))
	return r.ReadAll()
}

// ParseWithHeaders parses a CSV string using the first row as headers.
// Returns a slice of maps where each map represents a row with header keys.
// Ruby: csv.parse_with_headers(str)
func ParseWithHeaders(s string) ([]map[string]string, error) {
	r := csv.NewReader(strings.NewReader(s))
	return parseWithHeaders(r)
}

// ParseBytesWithHeaders parses CSV bytes using the first row as headers.
// Ruby: csv.parse_bytes_with_headers(bytes)
func ParseBytesWithHeaders(b []byte) ([]map[string]string, error) {
	r := csv.NewReader(bytes.NewReader(b))
	return parseWithHeaders(r)
}

func parseWithHeaders(r *csv.Reader) ([]map[string]string, error) {
	// Allow variable field counts per record
	r.FieldsPerRecord = -1

	headers, err := r.Read()
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("csv: empty input")
		}
		return nil, err
	}

	var result []map[string]string
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		record := make(map[string]string)
		for i, header := range headers {
			if i < len(row) {
				record[header] = row[i]
			} else {
				record[header] = ""
			}
		}
		result = append(result, record)
	}

	return result, nil
}

// Generate converts rows of fields to a CSV string.
// Ruby: csv.generate(rows)
func Generate(rows [][]string) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.WriteAll(rows); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateBytes converts rows of fields to CSV bytes.
// Ruby: csv.generate_bytes(rows)
func GenerateBytes(rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.WriteAll(rows); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateWithHeaders generates a CSV string from maps using specified headers.
// The headers define the order of columns in the output.
// Ruby: csv.generate_with_headers(headers, records)
func GenerateWithHeaders(headers []string, records []map[string]string) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Write header row
	if err := w.Write(headers); err != nil {
		return "", err
	}

	// Write data rows
	for _, record := range records {
		row := make([]string, len(headers))
		for i, h := range headers {
			row[i] = record[h]
		}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Row represents a single row for building CSV data.
type Row []string

// NewRow creates a new row with the given fields.
// Ruby: csv.row("a", "b", "c")
func NewRow(fields ...string) Row {
	return fields
}

// Reader provides streaming CSV parsing.
type Reader struct {
	r *csv.Reader
}

// NewReader creates a reader from a string.
// Ruby: csv.reader(str)
func NewReader(s string) *Reader {
	return &Reader{r: csv.NewReader(strings.NewReader(s))}
}

// NewReaderFromBytes creates a reader from bytes.
// Ruby: csv.reader_bytes(bytes)
func NewReaderFromBytes(b []byte) *Reader {
	return &Reader{r: csv.NewReader(bytes.NewReader(b))}
}

// Read reads and returns the next row.
// Returns nil, io.EOF at end of input.
// Ruby: reader.read()
func (r *Reader) Read() ([]string, error) {
	return r.r.Read()
}

// ReadAll reads and returns all remaining rows.
// Ruby: reader.read_all()
func (r *Reader) ReadAll() ([][]string, error) {
	return r.r.ReadAll()
}

// SetDelimiter sets the field delimiter (default is comma).
// Ruby: reader.delimiter = ';'
func (r *Reader) SetDelimiter(delim rune) {
	r.r.Comma = delim
}

// SetComment sets the comment character.
// Lines beginning with this character are ignored.
// Ruby: reader.comment = '#'
func (r *Reader) SetComment(c rune) {
	r.r.Comment = c
}

// SetLazyQuotes allows lazy handling of quotes.
// Ruby: reader.lazy_quotes = true
func (r *Reader) SetLazyQuotes(lazy bool) {
	r.r.LazyQuotes = lazy
}

// SetTrimLeadingSpace trims leading whitespace from fields.
// Ruby: reader.trim_leading_space = true
func (r *Reader) SetTrimLeadingSpace(trim bool) {
	r.r.TrimLeadingSpace = trim
}

// Writer provides streaming CSV writing.
type Writer struct {
	w   *csv.Writer
	buf *bytes.Buffer
}

// NewWriter creates a new CSV writer.
// Ruby: csv.writer()
func NewWriter() *Writer {
	buf := &bytes.Buffer{}
	return &Writer{
		w:   csv.NewWriter(buf),
		buf: buf,
	}
}

// Write writes a single row.
// Ruby: writer.write(row)
func (w *Writer) Write(row []string) error {
	return w.w.Write(row)
}

// WriteAll writes multiple rows.
// Ruby: writer.write_all(rows)
func (w *Writer) WriteAll(rows [][]string) error {
	return w.w.WriteAll(rows)
}

// SetDelimiter sets the field delimiter (default is comma).
// Ruby: writer.delimiter = ';'
func (w *Writer) SetDelimiter(delim rune) {
	w.w.Comma = delim
}

// SetUseCRLF sets whether to use \r\n as line terminator.
// Ruby: writer.use_crlf = true
func (w *Writer) SetUseCRLF(use bool) {
	w.w.UseCRLF = use
}

// Flush flushes buffered data to the underlying buffer.
// Ruby: writer.flush()
func (w *Writer) Flush() {
	w.w.Flush()
}

// String returns the CSV data as a string.
// Ruby: writer.to_s
func (w *Writer) String() string {
	w.w.Flush()
	return w.buf.String()
}

// Bytes returns the CSV data as bytes.
// Ruby: writer.bytes
func (w *Writer) Bytes() []byte {
	w.w.Flush()
	return w.buf.Bytes()
}

// Error returns any error that occurred during writing.
// Ruby: writer.error
func (w *Writer) Error() error {
	return w.w.Error()
}

// ParseLine parses a single CSV line into fields.
// Ruby: csv.parse_line(str)
func ParseLine(s string) ([]string, error) {
	r := csv.NewReader(strings.NewReader(s))
	return r.Read()
}

// Escape quotes a field for CSV output if needed.
// Ruby: csv.escape(str)
func Escape(s string) string {
	needsQuoting := strings.ContainsAny(s, ",\"\n\r")
	if !needsQuoting {
		return s
	}
	// Double any existing quotes and wrap in quotes
	escaped := strings.ReplaceAll(s, "\"", "\"\"")
	return "\"" + escaped + "\""
}

// Valid reports whether s is valid CSV.
// Ruby: csv.valid?(str)
func Valid(s string) bool {
	r := csv.NewReader(strings.NewReader(s))
	for {
		_, err := r.Read()
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}
	}
}

// RowCount returns the number of rows in a CSV string.
// Returns -1 if the CSV is invalid.
// Ruby: csv.row_count(str)
func RowCount(s string) int {
	r := csv.NewReader(strings.NewReader(s))
	count := 0
	for {
		_, err := r.Read()
		if err == io.EOF {
			return count
		}
		if err != nil {
			return -1
		}
		count++
	}
}

// Headers returns the first row of a CSV as headers.
// Ruby: csv.headers(str)
func Headers(s string) ([]string, error) {
	r := csv.NewReader(strings.NewReader(s))
	headers, err := r.Read()
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("csv: empty input")
		}
		return nil, err
	}
	return headers, nil
}
