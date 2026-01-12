// Package shell provides command execution for Rugby programs.
// Rugby: import rugby/shell
//
// Note: The Run and RunWithDir functions use the system shell (sh -c) and
// are designed for Unix-like systems. On Windows, use Exec with explicit
// command paths.
//
// Example:
//
//	output = shell.run("ls -la")!
//	result = shell.exec("git", ["status", "--short"])
//	if let path = shell.which("go")
//	  puts path
//	end
package shell

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

// Result contains the output from a command execution.
// Note: Stdout and Stderr include trailing newlines. Use TrimmedStdout()
// and TrimmedStderr() for output without trailing newlines.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Success reports whether the command exited with code 0.
// Ruby: result.success?
func (r *Result) Success() bool {
	return r.ExitCode == 0
}

// TrimmedStdout returns stdout with trailing whitespace removed.
// Ruby: result.trimmed_stdout
func (r *Result) TrimmedStdout() string {
	return strings.TrimRight(r.Stdout, "\n\r")
}

// TrimmedStderr returns stderr with trailing whitespace removed.
// Ruby: result.trimmed_stderr
func (r *Result) TrimmedStderr() string {
	return strings.TrimRight(r.Stderr, "\n\r")
}

// Run executes a shell command string and returns its stdout (trimmed).
// The command is passed to the system shell (sh -c on Unix).
// Ruby: shell.run(cmd)
func Run(cmd string) (string, error) {
	return RunWithDir(cmd, "")
}

// RunWithDir executes a shell command in a specific directory.
// Returns stdout with trailing newlines removed.
// Ruby: shell.run_with_dir(cmd, dir)
func RunWithDir(cmd, dir string) (string, error) {
	c := exec.Command("sh", "-c", cmd)
	if dir != "" {
		c.Dir = dir
	}
	out, err := c.Output()
	if err != nil {
		return string(out), err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

// execCommand runs a configured command and returns the result.
func execCommand(c *exec.Cmd) Result {
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

// Exec executes a command with arguments and returns detailed results.
// Ruby: shell.exec(name, args)
func Exec(name string, args []string) Result {
	return execCommand(exec.Command(name, args...))
}

// ExecWithDir executes a command in a specific directory.
// Ruby: shell.exec_with_dir(name, args, dir)
func ExecWithDir(name string, args []string, dir string) Result {
	c := exec.Command(name, args...)
	c.Dir = dir
	return execCommand(c)
}

// ExecWithEnv executes a command with custom environment variables.
// The provided env vars are added to the current environment.
// Ruby: shell.exec_with_env(name, args, env)
func ExecWithEnv(name string, args []string, env map[string]string) Result {
	c := exec.Command(name, args...)
	c.Env = os.Environ()
	for k, v := range env {
		c.Env = append(c.Env, k+"="+v)
	}
	return execCommand(c)
}

// Output executes a command and returns its stdout as a string (trimmed).
// Ruby: shell.output(name, args)
func Output(name string, args []string) (string, error) {
	c := exec.Command(name, args...)
	out, err := c.Output()
	if err != nil {
		return string(out), err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

// CombinedOutput executes a command and returns combined stdout and stderr (trimmed).
// Ruby: shell.combined_output(name, args)
func CombinedOutput(name string, args []string) (string, error) {
	c := exec.Command(name, args...)
	out, err := c.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

// Which finds the path to an executable.
// Returns empty string and false if not found.
// Ruby: shell.which(name)
func Which(name string) (string, bool) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", false
	}
	return path, true
}

// Exists reports whether an executable exists in PATH.
// Ruby: shell.exists?(name)
func Exists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
