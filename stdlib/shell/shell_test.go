package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	out, err := Run("echo hello")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "hello" {
		t.Errorf("Run() = %q, want %q", out, "hello")
	}
}

func TestRunWithPipe(t *testing.T) {
	out, err := Run("echo 'hello world' | tr 'a-z' 'A-Z'")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out != "HELLO WORLD" {
		t.Errorf("Run() = %q, want %q", out, "HELLO WORLD")
	}
}

func TestRunError(t *testing.T) {
	_, err := Run("exit 1")
	if err == nil {
		t.Error("Run() should return error for non-zero exit")
	}
}

func TestRunWithDir(t *testing.T) {
	tmp := t.TempDir()
	testFile := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := RunWithDir("ls", tmp)
	if err != nil {
		t.Fatalf("RunWithDir() error = %v", err)
	}
	if out != "test.txt" {
		t.Errorf("RunWithDir() = %q, want %q", out, "test.txt")
	}
}

func TestExec(t *testing.T) {
	result := Exec("echo", []string{"hello"})
	if !result.Success() {
		t.Errorf("Exec() success = false, want true")
	}
	if result.TrimmedStdout() != "hello" {
		t.Errorf("Exec() TrimmedStdout() = %q, want %q", result.TrimmedStdout(), "hello")
	}
	if result.ExitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", result.ExitCode)
	}
}

func TestExecWithStderr(t *testing.T) {
	result := Exec("sh", []string{"-c", "echo error >&2; exit 1"})
	if result.Success() {
		t.Error("Exec() should not succeed with exit 1")
	}
	if !strings.Contains(result.Stderr, "error") {
		t.Errorf("Exec() stderr = %q, want to contain 'error'", result.Stderr)
	}
	if result.ExitCode != 1 {
		t.Errorf("Exec() exit code = %d, want 1", result.ExitCode)
	}
}

func TestExecWithDir(t *testing.T) {
	tmp := t.TempDir()
	testFile := filepath.Join(tmp, "hello.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	result := ExecWithDir("ls", nil, tmp)
	if !result.Success() {
		t.Errorf("ExecWithDir() failed: %s", result.Stderr)
	}
	if result.TrimmedStdout() != "hello.txt" {
		t.Errorf("ExecWithDir() TrimmedStdout() = %q, want %q", result.TrimmedStdout(), "hello.txt")
	}
}

func TestExecWithEnv(t *testing.T) {
	result := ExecWithEnv("sh", []string{"-c", "echo $TEST_VAR"}, map[string]string{
		"TEST_VAR": "hello_env",
	})
	if !result.Success() {
		t.Fatalf("ExecWithEnv() failed: %s", result.Stderr)
	}
	if result.TrimmedStdout() != "hello_env" {
		t.Errorf("ExecWithEnv() TrimmedStdout() = %q, want %q", result.TrimmedStdout(), "hello_env")
	}
}

func TestOutput(t *testing.T) {
	out, err := Output("echo", []string{"hello"})
	if err != nil {
		t.Fatalf("Output() error = %v", err)
	}
	if out != "hello" {
		t.Errorf("Output() = %q, want %q", out, "hello")
	}
}

func TestOutputError(t *testing.T) {
	_, err := Output("sh", []string{"-c", "exit 1"})
	if err == nil {
		t.Error("Output() should return error for non-zero exit")
	}
}

func TestCombinedOutput(t *testing.T) {
	out, err := CombinedOutput("sh", []string{"-c", "echo out; echo err >&2"})
	if err != nil {
		t.Fatalf("CombinedOutput() error = %v", err)
	}
	if !strings.Contains(out, "out") || !strings.Contains(out, "err") {
		t.Errorf("CombinedOutput() = %q, want both stdout and stderr", out)
	}
}

func TestWhich(t *testing.T) {
	// "sh" should exist on all Unix systems
	path, ok := Which("sh")
	if !ok {
		t.Error("Which(sh) should find sh")
	}
	if path == "" {
		t.Error("Which(sh) returned empty path")
	}
}

func TestWhichNotFound(t *testing.T) {
	path, ok := Which("nonexistent_command_12345")
	if ok {
		t.Error("Which() should return false for nonexistent command")
	}
	if path != "" {
		t.Errorf("Which() path = %q, want empty", path)
	}
}

func TestExists(t *testing.T) {
	if !Exists("sh") {
		t.Error("Exists(sh) = false, want true")
	}
	if Exists("nonexistent_command_12345") {
		t.Error("Exists() = true for nonexistent command, want false")
	}
}

func TestResultSuccess(t *testing.T) {
	tests := []struct {
		exitCode int
		want     bool
	}{
		{0, true},
		{1, false},
		{127, false},
		{-1, false},
	}

	for _, tt := range tests {
		r := Result{ExitCode: tt.exitCode}
		if got := r.Success(); got != tt.want {
			t.Errorf("Result{ExitCode: %d}.Success() = %v, want %v", tt.exitCode, got, tt.want)
		}
	}
}

func TestTrimmedOutput(t *testing.T) {
	result := Exec("echo", []string{"hello"})

	// Raw stdout includes newline
	if !strings.HasSuffix(result.Stdout, "\n") {
		t.Errorf("Stdout should end with newline, got %q", result.Stdout)
	}

	// TrimmedStdout removes it
	if result.TrimmedStdout() != "hello" {
		t.Errorf("TrimmedStdout() = %q, want %q", result.TrimmedStdout(), "hello")
	}

	// TrimmedStderr works on stderr too
	result2 := Exec("sh", []string{"-c", "echo err >&2"})
	if result2.TrimmedStderr() != "err" {
		t.Errorf("TrimmedStderr() = %q, want %q", result2.TrimmedStderr(), "err")
	}
}
