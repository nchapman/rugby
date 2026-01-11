// Package base64 provides base64 encoding and decoding for Rugby programs.
// Rugby: import rugby/base64
//
// Example:
//
//	encoded = base64.encode("hello world".bytes)
//	decoded = base64.decode(encoded)!
package base64

import (
	"encoding/base64"
)

// Encode encodes bytes to a base64 string using standard encoding.
// Ruby: base64.encode(bytes)
func Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// EncodeString encodes a string to base64 using standard encoding.
// Ruby: base64.encode_string(str)
func EncodeString(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Decode decodes a base64 string to bytes using standard encoding.
// Ruby: base64.decode(str)
func Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// DecodeString decodes a base64 string to a string using standard encoding.
// Ruby: base64.decode_string(str)
func DecodeString(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// URLEncode encodes bytes to a URL-safe base64 string.
// Ruby: base64.url_encode(bytes)
func URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// URLEncodeString encodes a string to URL-safe base64.
// Ruby: base64.url_encode_string(str)
func URLEncodeString(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

// URLDecode decodes a URL-safe base64 string to bytes.
// Ruby: base64.url_decode(str)
func URLDecode(s string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(s)
}

// URLDecodeString decodes a URL-safe base64 string to a string.
// Ruby: base64.url_decode_string(str)
func URLDecodeString(s string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// RawEncode encodes bytes to base64 without padding.
// Ruby: base64.raw_encode(bytes)
func RawEncode(data []byte) string {
	return base64.RawStdEncoding.EncodeToString(data)
}

// RawEncodeString encodes a string to base64 without padding.
// Ruby: base64.raw_encode_string(str)
func RawEncodeString(s string) string {
	return base64.RawStdEncoding.EncodeToString([]byte(s))
}

// RawDecode decodes a base64 string without padding.
// Ruby: base64.raw_decode(str)
func RawDecode(s string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(s)
}

// RawDecodeString decodes a base64 string without padding to a string.
// Ruby: base64.raw_decode_string(str)
func RawDecodeString(s string) (string, error) {
	data, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// RawURLEncode encodes bytes to URL-safe base64 without padding.
// Ruby: base64.raw_url_encode(bytes)
func RawURLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// RawURLEncodeString encodes a string to URL-safe base64 without padding.
// Ruby: base64.raw_url_encode_string(str)
func RawURLEncodeString(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

// RawURLDecode decodes a URL-safe base64 string without padding.
// Ruby: base64.raw_url_decode(str)
func RawURLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// RawURLDecodeString decodes a URL-safe base64 string without padding to a string.
// Ruby: base64.raw_url_decode_string(str)
func RawURLDecodeString(s string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Valid reports whether s is valid base64.
// Ruby: base64.valid?(str)
func Valid(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// URLValid reports whether s is valid URL-safe base64.
// Ruby: base64.url_valid?(str)
func URLValid(s string) bool {
	_, err := base64.URLEncoding.DecodeString(s)
	return err == nil
}

// RawValid reports whether s is valid unpadded base64.
// Ruby: base64.raw_valid?(str)
func RawValid(s string) bool {
	_, err := base64.RawStdEncoding.DecodeString(s)
	return err == nil
}

// RawURLValid reports whether s is valid unpadded URL-safe base64.
// Ruby: base64.raw_url_valid?(str)
func RawURLValid(s string) bool {
	_, err := base64.RawURLEncoding.DecodeString(s)
	return err == nil
}
