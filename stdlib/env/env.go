// Package env provides environment variable access for Rugby programs.
// Rugby: import rugby/env
//
// Example:
//
//	port = env.fetch("PORT", "8080")
//	if let home = env.get("HOME")
//	  puts home
//	end
package env

import (
	"os"
	"strings"
)

// Get returns the value of an environment variable, or empty string and false if not set.
// Ruby: env.get(key)
func Get(key string) (string, bool) {
	return os.LookupEnv(key)
}

// Fetch returns the value of an environment variable, or the default if not set.
// Ruby: env.fetch(key, default)
func Fetch(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// Set sets an environment variable.
// Ruby: env.set(key, value)
func Set(key, value string) error {
	return os.Setenv(key, value)
}

// Delete removes an environment variable.
// Ruby: env.delete(key)
func Delete(key string) error {
	return os.Unsetenv(key)
}

// All returns all environment variables as a map.
// Ruby: env.all
func All() map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// Has reports whether an environment variable is set.
// Ruby: env.has?(key)
func Has(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

// Keys returns all environment variable names.
// Ruby: env.keys
func Keys() []string {
	var keys []string
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) >= 1 {
			keys = append(keys, parts[0])
		}
	}
	return keys
}

// Clear removes all environment variables.
// Ruby: env.clear
func Clear() {
	os.Clearenv()
}
