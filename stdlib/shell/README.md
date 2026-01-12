# rugby/shell

Command execution for Rugby programs.

## Import

```ruby
import rugby/shell
```

## Quick Start

```ruby
import rugby/shell

# Run a shell command
output = shell.Run("ls -la")!
puts output

# Execute with detailed results
result = shell.Exec("git", ["status", "--short"])
if result.Success()
  puts result.TrimmedStdout()
end

# Find an executable
if let path = shell.Which("go")
  puts "Go is at: #{path}"
end
```

## Functions

### Simple Execution

#### shell.Run(cmd) -> (String, error)

Executes a shell command string and returns stdout (trimmed).
The command is passed to the system shell (`sh -c` on Unix).

```ruby
output = shell.Run("echo hello")!
puts output  # hello

# Pipes work
output = shell.Run("cat file.txt | grep pattern")!
```

#### shell.RunWithDir(cmd, dir) -> (String, error)

Executes a shell command in a specific directory.

```ruby
output = shell.RunWithDir("ls", "/tmp")!
```

### Detailed Execution

#### shell.Exec(name, args) -> Result

Executes a command with arguments and returns detailed results.

```ruby
result = shell.Exec("git", ["commit", "-m", "message"])
puts result.Stdout
puts result.Stderr
puts result.ExitCode
if result.Success()
  puts "Committed!"
end
```

#### shell.ExecWithDir(name, args, dir) -> Result

Executes a command in a specific directory.

```ruby
result = shell.ExecWithDir("npm", ["install"], "/path/to/project")
```

#### shell.ExecWithEnv(name, args, env) -> Result

Executes a command with custom environment variables.

```ruby
result = shell.ExecWithEnv("node", ["app.js"], {
  "NODE_ENV": "production",
  "PORT": "3000"
})
```

### Result Type

The `Result` struct contains:

- `Stdout` - Standard output (includes trailing newline)
- `Stderr` - Standard error (includes trailing newline)
- `ExitCode` - Exit code (0 = success)

Methods:

- `Success()` - Returns true if exit code is 0
- `TrimmedStdout()` - Stdout without trailing whitespace
- `TrimmedStderr()` - Stderr without trailing whitespace

```ruby
result = shell.Exec("ls", ["-la"])
if result.Success()
  puts result.TrimmedStdout()
else
  puts "Error: #{result.TrimmedStderr()}"
end
```

### Output Helpers

#### shell.Output(name, args) -> (String, error)

Executes a command and returns stdout (trimmed).

```ruby
version = shell.Output("node", ["--version"])!
puts version  # v18.0.0
```

#### shell.CombinedOutput(name, args) -> (String, error)

Executes a command and returns combined stdout and stderr (trimmed).

```ruby
output = shell.CombinedOutput("make", ["build"])!
```

### Finding Executables

#### shell.Which(name) -> (String, Bool)

Finds the path to an executable.

```ruby
path, ok = shell.Which("python")
if ok
  puts "Python is at: #{path}"
end

# Or use if let
if let path = shell.Which("python")
  puts "Found: #{path}"
end
```

#### shell.Exists(name) -> Bool

Reports whether an executable exists in PATH.

```ruby
if shell.Exists("docker")
  # Docker is available
end
```

## Error Handling

Commands that fail return errors (for `Run`, `Output`, etc.) or non-zero exit codes (for `Exec`):

```ruby
# Propagate error
output = shell.Run("cat nonexistent.txt")!

# Handle error
output = shell.Run("cat nonexistent.txt") rescue "default"

# Check exit code
result = shell.Exec("grep", ["pattern", "file.txt"])
if !result.Success()
  puts "Not found (exit #{result.ExitCode})"
end
```

## Platform Notes

The `Run` and `RunWithDir` functions use `sh -c` and are designed for Unix-like systems. On Windows, use `Exec` with explicit command paths.

## Examples

### Run Build Script

```ruby
import rugby/shell

def build_project(dir : String) -> error
  result = shell.ExecWithDir("make", ["build"], dir)
  if !result.Success()
    puts result.TrimmedStderr()
    return fmt.Errorf("build failed with exit code %d", result.ExitCode)
  end
  puts result.TrimmedStdout()
  nil
end
```

### Git Operations

```ruby
import rugby/shell

def git_status -> String
  output = shell.Run("git status --short")!
  output
end

def git_commit(message : String) -> error
  result = shell.Exec("git", ["commit", "-m", message])
  if !result.Success()
    return fmt.Errorf("commit failed: %s", result.TrimmedStderr())
  end
  nil
end
```

### Check Dependencies

```ruby
import rugby/shell

def check_dependencies(deps : Array[String]) -> Bool
  for dep in deps
    unless shell.Exists(dep)
      puts "Missing: #{dep}"
      return false
    end
  end
  true
end

if check_dependencies(["node", "npm", "git"])
  puts "All dependencies present"
end
```

### Capture Command Output

```ruby
import rugby/shell

def get_system_info -> Map[String, String]
  {
    "hostname": shell.Run("hostname")! rescue "unknown",
    "user": shell.Run("whoami")! rescue "unknown",
    "pwd": shell.Run("pwd")! rescue "unknown",
    "shell": shell.Run("echo $SHELL")! rescue "unknown"
  }
end
```
