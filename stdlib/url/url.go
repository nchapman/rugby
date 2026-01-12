// Package url provides URL parsing and building for Rugby programs.
// Rugby: import rugby/url
//
// Example:
//
//	u = url.parse("https://example.com/path?q=search")!
//	query = url.query_encode({"q": "hello world", "page": "1"})
//	full = url.join("https://api.example.com", "/v1/users")
package url

import (
	"net/url"
	"path"
	"strings"
)

// URL represents a parsed URL.
type URL struct {
	raw *url.URL
}

// Parse parses a URL string and returns a URL object.
// Ruby: url.parse(str)
func Parse(rawURL string) (*URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &URL{raw: u}, nil
}

// MustParse parses a URL string and panics on error.
// Ruby: url.must_parse(str)
func MustParse(rawURL string) *URL {
	u, err := Parse(rawURL)
	if err != nil {
		panic("url: " + err.Error())
	}
	return u
}

// String returns the URL as a string.
// Ruby: u.to_s
func (u *URL) String() string {
	return u.raw.String()
}

// Scheme returns the URL scheme (e.g., "https").
// Ruby: u.scheme
func (u *URL) Scheme() string {
	return u.raw.Scheme
}

// Host returns the host (e.g., "example.com:8080").
// Ruby: u.host
func (u *URL) Host() string {
	return u.raw.Host
}

// Hostname returns the hostname without port.
// Ruby: u.hostname
func (u *URL) Hostname() string {
	return u.raw.Hostname()
}

// Port returns the port, or empty string if not specified.
// Ruby: u.port
func (u *URL) Port() string {
	return u.raw.Port()
}

// Path returns the URL path.
// Ruby: u.path
func (u *URL) Path() string {
	return u.raw.Path
}

// RawQuery returns the raw query string (without the ?).
// Ruby: u.raw_query
func (u *URL) RawQuery() string {
	return u.raw.RawQuery
}

// Query returns the query parameters as a map.
// Ruby: u.query
func (u *URL) Query() map[string]string {
	result := make(map[string]string)
	for k, v := range u.raw.Query() {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// QueryValues returns the query parameters with multiple values.
// Ruby: u.query_values
func (u *URL) QueryValues() map[string][]string {
	return u.raw.Query()
}

// Fragment returns the fragment (part after #).
// Ruby: u.fragment
func (u *URL) Fragment() string {
	return u.raw.Fragment
}

// User returns the username from user info, if any.
// Ruby: u.user
func (u *URL) User() string {
	if u.raw.User != nil {
		return u.raw.User.Username()
	}
	return ""
}

// Password returns the password from user info, if any.
// Returns empty string and false if no password set.
// Ruby: u.password
func (u *URL) Password() (string, bool) {
	if u.raw.User != nil {
		return u.raw.User.Password()
	}
	return "", false
}

// IsAbsolute reports whether the URL is absolute (has a scheme).
// Ruby: u.absolute?
func (u *URL) IsAbsolute() bool {
	return u.raw.IsAbs()
}

// WithScheme returns a new URL with the given scheme.
// Ruby: u.with_scheme(scheme)
func (u *URL) WithScheme(scheme string) *URL {
	newU := *u.raw
	newU.Scheme = scheme
	return &URL{raw: &newU}
}

// WithHost returns a new URL with the given host.
// Ruby: u.with_host(host)
func (u *URL) WithHost(host string) *URL {
	newU := *u.raw
	newU.Host = host
	return &URL{raw: &newU}
}

// WithPath returns a new URL with the given path.
// Ruby: u.with_path(path)
func (u *URL) WithPath(p string) *URL {
	newU := *u.raw
	newU.Path = p
	return &URL{raw: &newU}
}

// WithQuery returns a new URL with the given query parameters.
// Ruby: u.with_query(params)
func (u *URL) WithQuery(params map[string]string) *URL {
	newU := *u.raw
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	newU.RawQuery = q.Encode()
	return &URL{raw: &newU}
}

// WithFragment returns a new URL with the given fragment.
// Ruby: u.with_fragment(fragment)
func (u *URL) WithFragment(fragment string) *URL {
	newU := *u.raw
	newU.Fragment = fragment
	return &URL{raw: &newU}
}

// Resolve resolves a relative reference against this URL.
// Ruby: u.resolve(ref)
func (u *URL) Resolve(ref string) (*URL, error) {
	refU, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}
	return &URL{raw: u.raw.ResolveReference(refU)}, nil
}

// Join joins a base URL with path segments.
// Note: paths are cleaned (trailing slashes removed, . and .. resolved).
// Ruby: url.join(base, paths...)
func Join(base string, paths ...string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	parts := append([]string{u.Path}, paths...)
	u.Path = path.Join(parts...)
	return u.String(), nil
}

// MustJoin joins a base URL with path segments, panicking on error.
// Ruby: url.must_join(base, paths...)
func MustJoin(base string, paths ...string) string {
	result, err := Join(base, paths...)
	if err != nil {
		panic("url: " + err.Error())
	}
	return result
}

// QueryEncode encodes a map as a URL query string.
// Ruby: url.query_encode(params)
func QueryEncode(params map[string]string) string {
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	return q.Encode()
}

// QueryEncodeValues encodes a map with multiple values as a URL query string.
// Ruby: url.query_encode_values(params)
func QueryEncodeValues(params map[string][]string) string {
	return url.Values(params).Encode()
}

// QueryDecode decodes a URL query string into a map.
// Ruby: url.query_decode(str)
func QueryDecode(query string) (map[string]string, error) {
	values, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result, nil
}

// QueryDecodeValues decodes a URL query string into a map with multiple values.
// Ruby: url.query_decode_values(str)
func QueryDecodeValues(query string) (map[string][]string, error) {
	return url.ParseQuery(query)
}

// Encode URL-encodes a string.
// Ruby: url.encode(str)
func Encode(s string) string {
	return url.QueryEscape(s)
}

// Decode URL-decodes a string.
// Ruby: url.decode(str)
func Decode(s string) (string, error) {
	return url.QueryUnescape(s)
}

// PathEncode encodes a string for use in a URL path.
// Ruby: url.path_encode(str)
func PathEncode(s string) string {
	return url.PathEscape(s)
}

// PathDecode decodes a URL path-encoded string.
// Ruby: url.path_decode(str)
func PathDecode(s string) (string, error) {
	return url.PathUnescape(s)
}

// Valid reports whether a string can be parsed as a URL.
// Note: Go's url.Parse is very permissive; empty strings and relative paths
// are considered valid. Use IsHTTP for stricter HTTP URL validation.
// Ruby: url.valid?(str)
func Valid(rawURL string) bool {
	_, err := url.Parse(rawURL)
	return err == nil
}

// IsHTTP reports whether a URL has http or https scheme.
// Ruby: url.http?(str)
func IsHTTP(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return scheme == "http" || scheme == "https"
}
