// Package uuid provides UUID generation and parsing for Rugby programs.
// Rugby: import rugby/uuid
//
// Example:
//
//	id = uuid.v4()!
//	parsed = uuid.parse(str)!
//	valid = uuid.valid?(str)
package uuid

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// V4 generates a random UUID (version 4) as a string.
// Returns a lowercase UUID in the format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
// Ruby: uuid.v4()
func V4() (string, error) {
	var u [16]byte
	_, err := rand.Read(u[:])
	if err != nil {
		return "", err
	}

	// Set version 4 (random) - bits 4-7 of byte 6
	u[6] = (u[6] & 0x0f) | 0x40
	// Set variant to RFC 4122 - bits 6-7 of byte 8
	u[8] = (u[8] & 0x3f) | 0x80

	return formatUUID(u), nil
}

// V7 generates a time-ordered UUID (version 7) as a string.
// V7 UUIDs are sortable by creation time and use Unix milliseconds.
// Ruby: uuid.v7()
func V7() (string, error) {
	var u [16]byte

	// Get current time in milliseconds since Unix epoch
	ms := uint64(time.Now().UnixMilli())

	// First 48 bits: Unix timestamp in milliseconds (big endian)
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)

	// Remaining bits: random
	_, err := rand.Read(u[6:])
	if err != nil {
		return "", err
	}

	// Set version 7 - bits 4-7 of byte 6
	u[6] = (u[6] & 0x0f) | 0x70
	// Set variant to RFC 4122 - bits 6-7 of byte 8
	u[8] = (u[8] & 0x3f) | 0x80

	return formatUUID(u), nil
}

// Parse parses a UUID string and returns the normalized lowercase form.
// Accepts UUIDs with or without hyphens, in any case.
// Ruby: uuid.parse(str)
func Parse(s string) (string, error) {
	// Remove hyphens if present
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ToLower(s)

	if len(s) != 32 {
		return "", errors.New("uuid: invalid length")
	}

	// Validate hex characters
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", errors.New("uuid: invalid character")
		}
	}

	// Format with hyphens
	return s[0:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:32], nil
}

// Valid reports whether s is a valid UUID string.
// Accepts UUIDs with or without hyphens, in any case.
// Ruby: uuid.valid?(str)
func Valid(s string) bool {
	_, err := Parse(s)
	return err == nil
}

// Nil returns the nil UUID (all zeros).
// Ruby: uuid.nil()
func Nil() string {
	return "00000000-0000-0000-0000-000000000000"
}

// IsNil reports whether s is the nil UUID.
// Ruby: uuid.nil?(str)
func IsNil(s string) bool {
	parsed, err := Parse(s)
	if err != nil {
		return false
	}
	return parsed == Nil()
}

// Version returns the version number of the UUID (1-7).
// Returns 0 if the UUID is invalid.
// Ruby: uuid.version(str)
func Version(s string) int {
	parsed, err := Parse(s)
	if err != nil {
		return 0
	}
	// Version is in the 13th character (position 14 with hyphens)
	// Format: xxxxxxxx-xxxx-Vxxx-xxxx-xxxxxxxxxxxx
	v := parsed[14]
	if v >= '0' && v <= '9' {
		return int(v - '0')
	}
	if v >= 'a' && v <= 'f' {
		return int(v - 'a' + 10)
	}
	return 0
}

// Bytes converts a UUID string to its 16-byte representation.
// Ruby: uuid.bytes(str)
func Bytes(s string) ([]byte, error) {
	parsed, err := Parse(s)
	if err != nil {
		return nil, err
	}

	// Remove hyphens and decode
	hexStr := strings.ReplaceAll(parsed, "-", "")
	return hex.DecodeString(hexStr)
}

// FromBytes creates a UUID string from a 16-byte slice.
// Ruby: uuid.from_bytes(bytes)
func FromBytes(b []byte) (string, error) {
	if len(b) != 16 {
		return "", errors.New("uuid: invalid byte length, expected 16")
	}
	var u [16]byte
	copy(u[:], b)
	return formatUUID(u), nil
}

// Timestamp extracts the timestamp from a V7 UUID.
// Returns the time as Unix milliseconds.
// Returns 0 if the UUID is not a valid V7 UUID.
// Ruby: uuid.timestamp(str)
func Timestamp(s string) int64 {
	b, err := Bytes(s)
	if err != nil {
		return 0
	}
	if Version(s) != 7 {
		return 0
	}
	// First 48 bits are the timestamp in milliseconds
	ms := uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 |
		uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5])
	return int64(ms)
}

// Time extracts the timestamp from a V7 UUID as a time.Time.
// Returns the zero time if the UUID is not a valid V7 UUID.
// Ruby: uuid.time(str)
func Time(s string) time.Time {
	ms := Timestamp(s)
	if ms == 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

// Compare compares two UUIDs lexicographically.
// Returns -1 if a < b, 0 if a == b, and 1 if a > b.
// Returns 0 if either UUID is invalid.
// Ruby: uuid.compare(a, b)
func Compare(a, b string) int {
	parsedA, errA := Parse(a)
	parsedB, errB := Parse(b)
	if errA != nil || errB != nil {
		return 0
	}
	return strings.Compare(parsedA, parsedB)
}

// Equal reports whether two UUIDs are equal.
// Handles case differences and hyphen presence/absence.
// Ruby: uuid.equal?(a, b)
func Equal(a, b string) bool {
	parsedA, errA := Parse(a)
	parsedB, errB := Parse(b)
	if errA != nil || errB != nil {
		return false
	}
	return parsedA == parsedB
}

// Deterministic generates a deterministic UUID from namespace and name.
// Uses SHA-256 to hash namespace+name and formats as UUID v4 style.
// This is NOT a standard UUID v5, but useful for reproducible IDs.
// Ruby: uuid.deterministic(namespace, name)
func Deterministic(namespace, name string) string {
	data := []byte(namespace + "\x00" + name)
	hash := sha256.Sum256(data)

	var u [16]byte
	copy(u[:], hash[:16])

	// Set version 4 and variant bits
	u[6] = (u[6] & 0x0f) | 0x40
	u[8] = (u[8] & 0x3f) | 0x80

	return formatUUID(u)
}

// formatUUID formats a 16-byte array as a UUID string.
func formatUUID(u [16]byte) string {
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], u[10:16])
	return string(buf)
}

// MustV4 generates a V4 UUID and panics on error.
// Ruby: uuid.must_v4()
func MustV4() string {
	u, err := V4()
	if err != nil {
		panic(err)
	}
	return u
}

// MustV7 generates a V7 UUID and panics on error.
// Ruby: uuid.must_v7()
func MustV7() string {
	u, err := V7()
	if err != nil {
		panic(err)
	}
	return u
}
