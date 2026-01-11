// Package http provides a simple HTTP client for Rugby programs.
// Rugby: import rugby/http
//
// Example:
//
//	resp = http.get("https://api.example.com/users")!
//	data = resp.json()!
package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultTimeout is the default request timeout.
var DefaultTimeout = 30 * time.Second

// Response wraps an HTTP response with convenient methods.
type Response struct {
	Status     int               // HTTP status code (200, 404, etc.)
	StatusText string            // HTTP status text ("OK", "Not Found", etc.)
	Headers    map[string]string // Response headers (first value only for multi-value headers)
	body       []byte            // Cached body
}

// Ok returns true if the status code is 2xx.
// Ruby: resp.ok?
func (r *Response) Ok() bool {
	return r.Status >= 200 && r.Status < 300
}

// Body returns the response body as a string.
// Ruby: resp.body
func (r *Response) Body() string {
	return string(r.body)
}

// Bytes returns the response body as bytes.
// Ruby: resp.bytes
func (r *Response) Bytes() []byte {
	return r.body
}

// JSON parses the response body as JSON into a map.
// Ruby: resp.json
func (r *Response) JSON() (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal(r.body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// JSONArray parses the response body as a JSON array.
// Ruby: resp.json_array
func (r *Response) JSONArray() ([]any, error) {
	var result []any
	if err := json.Unmarshal(r.body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// JSONInto parses the response body into the provided struct.
// Ruby: resp.json_into(target)
func (r *Response) JSONInto(v any) error {
	return json.Unmarshal(r.body, v)
}

// Options configures an HTTP request.
type Options struct {
	Headers map[string]string // Request headers
	Timeout time.Duration     // Request timeout
	JSON    any               // JSON body (will be marshaled)
	Form    map[string]string // Form data (application/x-www-form-urlencoded)
	Body    []byte            // Raw body bytes
}

// Get performs an HTTP GET request.
// Ruby: http.get(url) or http.get(url, options)
func Get(rawURL string, opts ...Options) (*Response, error) {
	return doRequest("GET", rawURL, mergeOpts(opts))
}

// Post performs an HTTP POST request.
// Ruby: http.post(url, options)
func Post(rawURL string, opts ...Options) (*Response, error) {
	return doRequest("POST", rawURL, mergeOpts(opts))
}

// Put performs an HTTP PUT request.
// Ruby: http.put(url, options)
func Put(rawURL string, opts ...Options) (*Response, error) {
	return doRequest("PUT", rawURL, mergeOpts(opts))
}

// Patch performs an HTTP PATCH request.
// Ruby: http.patch(url, options)
func Patch(rawURL string, opts ...Options) (*Response, error) {
	return doRequest("PATCH", rawURL, mergeOpts(opts))
}

// Delete performs an HTTP DELETE request.
// Ruby: http.delete(url, options)
func Delete(rawURL string, opts ...Options) (*Response, error) {
	return doRequest("DELETE", rawURL, mergeOpts(opts))
}

// Head performs an HTTP HEAD request.
// Ruby: http.head(url, options)
func Head(rawURL string, opts ...Options) (*Response, error) {
	return doRequest("HEAD", rawURL, mergeOpts(opts))
}

func mergeOpts(opts []Options) Options {
	if len(opts) == 0 {
		return Options{}
	}
	return opts[0]
}

func doRequest(method, rawURL string, opts Options) (*Response, error) {
	// Build request body
	var bodyReader io.Reader
	var contentType string

	if opts.JSON != nil {
		jsonBytes, err := json.Marshal(opts.JSON)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBytes)
		contentType = "application/json"
	} else if opts.Form != nil {
		form := url.Values{}
		for k, v := range opts.Form {
			form.Set(k, v)
		}
		bodyReader = strings.NewReader(form.Encode())
		contentType = "application/x-www-form-urlencoded"
	} else if opts.Body != nil {
		bodyReader = bytes.NewReader(opts.Body)
	}

	// Create request
	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set content type if we have a body
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Set custom headers
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// Create client with timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	client := &http.Client{Timeout: timeout}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract headers
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &Response{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    headers,
		body:       body,
	}, nil
}
