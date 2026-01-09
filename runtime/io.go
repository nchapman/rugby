package runtime

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// stdinScanner is reused across Gets() calls to avoid losing buffered input.
var stdinScanner = bufio.NewScanner(os.Stdin)

// Puts prints values to stdout, each followed by a newline.
// Ruby: puts
func Puts(args ...any) {
	for _, arg := range args {
		fmt.Println(arg)
	}
}

// Print prints values to stdout without a trailing newline.
// Ruby: print
func Print(args ...any) {
	for _, arg := range args {
		fmt.Print(arg)
	}
}

// P prints values with debug formatting (like Ruby's p).
// Shows type information and quotes strings.
// Ruby: p value
func P(args ...any) {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%#v", arg)
	}
	fmt.Println()
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
