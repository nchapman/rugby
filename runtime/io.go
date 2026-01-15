package runtime

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"time"
)

// stdinScanner is reused across Gets() calls to avoid losing buffered input.
var stdinScanner = bufio.NewScanner(os.Stdin)

// deref dereferences a pointer value for printing.
// Returns the dereferenced value, or nil if the pointer is nil.
// Non-pointer values are returned unchanged.
func deref(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		return rv.Elem().Interface()
	}
	return v
}

// Puts prints values to stdout, each followed by a newline.
// Automatically dereferences pointers (from first/last/etc) for clean output.
// Ruby: puts
func Puts(args ...any) {
	for _, arg := range args {
		fmt.Println(deref(arg))
	}
}

// Print prints values to stdout without a trailing newline.
// Automatically dereferences pointers for clean output.
// Ruby: print
func Print(args ...any) {
	for _, arg := range args {
		fmt.Print(deref(arg))
	}
}

// P prints values with debug formatting (like Ruby's p).
// Shows type information and quotes strings.
// Unlike Puts/Print, does NOT dereference pointers - showing the type is useful for debugging.
// Ruby: p value
func P(args ...any) {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		formatDebugValue(arg)
	}
	fmt.Println()
}

// formatDebugValue prints a value in Ruby-like debug format.
func formatDebugValue(arg any) {
	switch v := arg.(type) {
	case []int:
		fmt.Print("[")
		for i, elem := range v {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(elem)
		}
		fmt.Print("]")
	case []string:
		fmt.Print("[")
		for i, elem := range v {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Printf("%q", elem)
		}
		fmt.Print("]")
	case []any:
		fmt.Print("[")
		for i, elem := range v {
			if i > 0 {
				fmt.Print(" ")
			}
			formatDebugValue(elem)
		}
		fmt.Print("]")
	case string:
		fmt.Printf("%q", v)
	default:
		// Fall back to Go's default formatting
		fmt.Printf("%#v", v)
	}
}

// Gets reads a line from stdin, returning the string without the newline.
// Returns an empty string on EOF or error.
// Ruby: gets
func Gets() string {
	if stdinScanner.Scan() {
		return stdinScanner.Text()
	}
	return ""
}

// GetsWithPrompt prints a prompt then reads a line from stdin.
// Ruby: print prompt; gets
func GetsWithPrompt(prompt string) string {
	fmt.Print(prompt)
	return Gets()
}

// Exit terminates the program with the given status code.
// Ruby: exit(code)
func Exit(code int) {
	os.Exit(code)
}

// Fatal prints an error to stderr and exits with status code 1.
// Used by error propagation (!) in main/scripts.
func Fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// Error creates a new error with the given message.
// Rugby: error("message")
func Error(msg string) error {
	return errors.New(msg)
}

// Sleep pauses execution for the given number of seconds.
// Ruby: sleep(seconds)
func Sleep(seconds float64) {
	time.Sleep(time.Duration(seconds * float64(time.Second)))
}

// SleepMs pauses execution for the given number of milliseconds.
func SleepMs(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// RandInt returns a random integer in [0, n).
// Ruby: rand(n)
func RandInt(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}

// RandFloat returns a random float in [0.0, 1.0).
// Ruby: rand
func RandFloat() float64 {
	return rand.Float64()
}

// RandRange returns a random integer in [min, max].
// Ruby: rand(min..max)
func RandRange(min, max int) int {
	if min > max {
		min, max = max, min
	}
	return min + rand.Intn(max-min+1)
}
