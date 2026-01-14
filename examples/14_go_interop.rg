# Rugby Go Interop
# Demonstrates: importing Go packages, snake_case mapping, defer
#
# Rugby automatically maps snake_case to Go's CamelCase.
# Use whichever feels natural - both work.

import strings
import strconv
import fmt

def main
  s = "hello world"

  # Call Go's strings package
  # You can use original Go names or snake_case (both work)
  upper = strings.to_upper(s)    # or strings.ToUpper
  puts "Upper: #{upper}"

  has_world = strings.contains(s, "world")  # or strings.Contains
  puts "Contains 'world': #{has_world}"

  parts = strings.split(s, " ")  # or strings.Split
  puts "Split: #{parts}"

  words = ["one", "two", "three"]
  joined = strings.join(words, "-")  # or strings.Join
  puts "Joined: #{joined}"

  replaced = strings.replace_all(s, "world", "Rugby")  # or strings.ReplaceAll
  puts "Replaced: #{replaced}"

  # strconv package
  num_str = strconv.itoa(42)  # or strconv.Itoa
  puts "Int to string: #{num_str}"

  parsed, err = strconv.atoi("123")  # or strconv.Atoi
  if err == nil
    puts "String to int: #{parsed}"
  end

  # fmt.Sprintf for complex formatting
  formatted = fmt.sprintf("Name: %s, Age: %d", "Alice", 30)  # or fmt.Sprintf
  puts formatted

  # defer - runs at function exit
  cleanup_demo
end

def cleanup_demo
  puts "Starting cleanup demo"
  defer puts "Deferred: runs last"
  puts "Middle of function"
  puts "End of function body"
end
