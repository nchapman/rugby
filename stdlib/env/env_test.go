package env

import (
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	key := "RUGBY_TEST_GET"
	defer func() { _ = os.Unsetenv(key) }()

	// Not set
	val, ok := Get(key)
	if ok {
		t.Errorf("Get(%q) ok = true, want false", key)
	}
	if val != "" {
		t.Errorf("Get(%q) = %q, want empty string", key, val)
	}

	// Set
	if err := os.Setenv(key, "test-value"); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}
	val, ok = Get(key)
	if !ok {
		t.Errorf("Get(%q) ok = false, want true", key)
	}
	if val != "test-value" {
		t.Errorf("Get(%q) = %q, want %q", key, val, "test-value")
	}
}

func TestGetEmptyValue(t *testing.T) {
	key := "RUGBY_TEST_EMPTY"
	defer func() { _ = os.Unsetenv(key) }()

	if err := os.Setenv(key, ""); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}

	val, ok := Get(key)
	if !ok {
		t.Errorf("Get(%q) ok = false, want true (empty value is set)", key)
	}
	if val != "" {
		t.Errorf("Get(%q) = %q, want empty string", key, val)
	}
}

func TestFetch(t *testing.T) {
	key := "RUGBY_TEST_FETCH"
	defer func() { _ = os.Unsetenv(key) }()

	// Not set - use default
	val := Fetch(key, "default")
	if val != "default" {
		t.Errorf("Fetch(%q, %q) = %q, want %q", key, "default", val, "default")
	}

	// Set - use value
	if err := os.Setenv(key, "actual"); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}
	val = Fetch(key, "default")
	if val != "actual" {
		t.Errorf("Fetch(%q, %q) = %q, want %q", key, "default", val, "actual")
	}
}

func TestSet(t *testing.T) {
	key := "RUGBY_TEST_SET"
	defer func() { _ = os.Unsetenv(key) }()

	if err := Set(key, "new-value"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	val, ok := os.LookupEnv(key)
	if !ok {
		t.Errorf("os.LookupEnv(%q) ok = false, want true", key)
	}
	if val != "new-value" {
		t.Errorf("os.LookupEnv(%q) = %q, want %q", key, val, "new-value")
	}
}

func TestDelete(t *testing.T) {
	key := "RUGBY_TEST_DELETE"

	if err := os.Setenv(key, "to-delete"); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}

	if err := Delete(key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, ok := os.LookupEnv(key)
	if ok {
		t.Errorf("os.LookupEnv(%q) ok = true after Delete(), want false", key)
	}
}

func TestAll(t *testing.T) {
	key := "RUGBY_TEST_ALL"
	defer func() { _ = os.Unsetenv(key) }()

	if err := os.Setenv(key, "all-value"); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}

	all := All()
	if all[key] != "all-value" {
		t.Errorf("All()[%q] = %q, want %q", key, all[key], "all-value")
	}
}

func TestHas(t *testing.T) {
	key := "RUGBY_TEST_HAS"
	defer func() { _ = os.Unsetenv(key) }()

	if Has(key) {
		t.Errorf("Has(%q) = true, want false", key)
	}

	if err := os.Setenv(key, "exists"); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}

	if !Has(key) {
		t.Errorf("Has(%q) = false, want true", key)
	}
}

func TestKeys(t *testing.T) {
	key := "RUGBY_TEST_KEYS"
	defer func() { _ = os.Unsetenv(key) }()

	if err := os.Setenv(key, "keys-value"); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}

	keys := Keys()
	found := false
	for _, k := range keys {
		if k == key {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Keys() does not contain %q", key)
	}
}

func TestAllWithEqualsInValue(t *testing.T) {
	key := "RUGBY_TEST_EQUALS"
	defer func() { _ = os.Unsetenv(key) }()

	// Value contains equals sign
	if err := os.Setenv(key, "key=value"); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}

	all := All()
	if all[key] != "key=value" {
		t.Errorf("All()[%q] = %q, want %q", key, all[key], "key=value")
	}
}

func TestFetchEmptyValue(t *testing.T) {
	key := "RUGBY_TEST_FETCH_EMPTY"
	defer func() { _ = os.Unsetenv(key) }()

	// Set to empty string
	if err := os.Setenv(key, ""); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}

	val := Fetch(key, "default")
	if val != "" {
		t.Errorf("Fetch(%q, %q) = %q, want empty string (set but empty)", key, "default", val)
	}
}
