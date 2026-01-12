// Package crypto provides cryptographic utilities for Rugby programs.
// Rugby: import rugby/crypto
//
// Example:
//
//	hash = crypto.sha256("hello world")
//	random = crypto.random_bytes(16)!
//	valid = crypto.constant_time_compare(a, b)
package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"errors"
)

// SHA256 returns the SHA-256 hash of data as a hex string.
// Ruby: crypto.sha256(data)
func SHA256(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// SHA256String returns the SHA-256 hash of a string as a hex string.
// Ruby: crypto.sha256_string(str)
func SHA256String(s string) string {
	return SHA256([]byte(s))
}

// SHA256Bytes returns the SHA-256 hash of data as bytes.
// Ruby: crypto.sha256_bytes(data)
func SHA256Bytes(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// SHA512 returns the SHA-512 hash of data as a hex string.
// Ruby: crypto.sha512(data)
func SHA512(data []byte) string {
	h := sha512.Sum512(data)
	return hex.EncodeToString(h[:])
}

// SHA512String returns the SHA-512 hash of a string as a hex string.
// Ruby: crypto.sha512_string(str)
func SHA512String(s string) string {
	return SHA512([]byte(s))
}

// SHA512Bytes returns the SHA-512 hash of data as bytes.
// Ruby: crypto.sha512_bytes(data)
func SHA512Bytes(data []byte) []byte {
	h := sha512.Sum512(data)
	return h[:]
}

// SHA1 returns the SHA-1 hash of data as a hex string.
// Note: SHA-1 is cryptographically broken; use SHA-256 for security.
// Ruby: crypto.sha1(data)
func SHA1(data []byte) string {
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
}

// SHA1String returns the SHA-1 hash of a string as a hex string.
// Ruby: crypto.sha1_string(str)
func SHA1String(s string) string {
	return SHA1([]byte(s))
}

// SHA1Bytes returns the SHA-1 hash of data as bytes.
// Ruby: crypto.sha1_bytes(data)
func SHA1Bytes(data []byte) []byte {
	h := sha1.Sum(data)
	return h[:]
}

// MD5 returns the MD5 hash of data as a hex string.
// Note: MD5 is cryptographically broken; use SHA-256 for security.
// Ruby: crypto.md5(data)
func MD5(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:])
}

// MD5String returns the MD5 hash of a string as a hex string.
// Ruby: crypto.md5_string(str)
func MD5String(s string) string {
	return MD5([]byte(s))
}

// MD5Bytes returns the MD5 hash of data as bytes.
// Ruby: crypto.md5_bytes(data)
func MD5Bytes(data []byte) []byte {
	h := md5.Sum(data)
	return h[:]
}

// RandomBytes generates n cryptographically secure random bytes.
// Returns an error if n is negative.
// Ruby: crypto.random_bytes(n)
func RandomBytes(n int) ([]byte, error) {
	if n < 0 {
		return nil, errors.New("crypto: negative byte count")
	}
	if n == 0 {
		return []byte{}, nil
	}
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// RandomHex generates n random bytes and returns them as a hex string.
// Ruby: crypto.random_hex(n)
func RandomHex(n int) (string, error) {
	b, err := RandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ConstantTimeCompare compares two byte slices in constant time.
// Returns true if they are equal.
// Ruby: crypto.constant_time_compare(a, b)
func ConstantTimeCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

// ConstantTimeCompareString compares two strings in constant time.
// Returns true if they are equal.
// Ruby: crypto.constant_time_compare_string(a, b)
func ConstantTimeCompareString(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// HMACSHA256 computes HMAC-SHA256 of data with the given key.
// Returns the result as a hex string.
// Ruby: crypto.hmac_sha256(key, data)
func HMACSHA256(key, data []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// HMACSHA256String computes HMAC-SHA256 of a string with the given key string.
// Ruby: crypto.hmac_sha256_string(key, data)
func HMACSHA256String(key, data string) string {
	return HMACSHA256([]byte(key), []byte(data))
}

// HMACSHA256Bytes computes HMAC-SHA256 of data with the given key.
// Returns the result as bytes.
// Ruby: crypto.hmac_sha256_bytes(key, data)
func HMACSHA256Bytes(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// HMACSHA512 computes HMAC-SHA512 of data with the given key.
// Returns the result as a hex string.
// Ruby: crypto.hmac_sha512(key, data)
func HMACSHA512(key, data []byte) string {
	h := hmac.New(sha512.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// HMACSHA512String computes HMAC-SHA512 of a string with the given key string.
// Ruby: crypto.hmac_sha512_string(key, data)
func HMACSHA512String(key, data string) string {
	return HMACSHA512([]byte(key), []byte(data))
}

// HMACSHA512Bytes computes HMAC-SHA512 of data with the given key.
// Returns the result as bytes.
// Ruby: crypto.hmac_sha512_bytes(key, data)
func HMACSHA512Bytes(key, data []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// VerifyHMACSHA256 verifies that a signature matches the expected HMAC-SHA256.
// Uses constant-time comparison to prevent timing attacks.
// Ruby: crypto.verify_hmac_sha256(key, data, signature)
func VerifyHMACSHA256(key, data []byte, signature string) bool {
	expected := HMACSHA256(key, data)
	return ConstantTimeCompareString(expected, signature)
}

// VerifyHMACSHA512 verifies that a signature matches the expected HMAC-SHA512.
// Uses constant-time comparison to prevent timing attacks.
// Ruby: crypto.verify_hmac_sha512(key, data, signature)
func VerifyHMACSHA512(key, data []byte, signature string) bool {
	expected := HMACSHA512(key, data)
	return ConstantTimeCompareString(expected, signature)
}

// VerifyHMACSHA256Bytes verifies a raw byte signature against HMAC-SHA256.
// Uses constant-time comparison to prevent timing attacks.
// Ruby: crypto.verify_hmac_sha256_bytes(key, data, signature)
func VerifyHMACSHA256Bytes(key, data, signature []byte) bool {
	expected := HMACSHA256Bytes(key, data)
	return ConstantTimeCompare(expected, signature)
}

// VerifyHMACSHA512Bytes verifies a raw byte signature against HMAC-SHA512.
// Uses constant-time comparison to prevent timing attacks.
// Ruby: crypto.verify_hmac_sha512_bytes(key, data, signature)
func VerifyHMACSHA512Bytes(key, data, signature []byte) bool {
	expected := HMACSHA512Bytes(key, data)
	return ConstantTimeCompare(expected, signature)
}
