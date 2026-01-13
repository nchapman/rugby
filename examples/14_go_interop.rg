# Rugby Go Interop
# Demonstrates: importing Go packages, snake_case mapping, defer (spec 11)

import strings
import strconv
import fmt

def main
  s = "hello world"

  # Go's strings package
  # snake_case calls map to CamelCase (spec 11)
  upper = strings.ToUpper(s)
  puts "Upper: #{upper}"

  has_world = strings.Contains(s, "world")
  puts "Contains 'world': #{has_world}"

  parts = strings.Split(s, " ")
  puts "Split: #{parts}"

  words = ["one", "two", "three"]
  joined = strings.Join(words, "-")
  puts "Joined: #{joined}"

  replaced = strings.ReplaceAll(s, "world", "Rugby")
  puts "Replaced: #{replaced}"

  # strconv package
  num_str = strconv.Itoa(42)
  puts "Int to string: #{num_str}"

  parsed, err = strconv.Atoi("123")
  if err == nil
    puts "String to int: #{parsed}"
  end

  # fmt.Sprintf for complex formatting
  formatted = fmt.Sprintf("Name: %s, Age: %d", "Alice", 30)
  puts formatted

  # defer - runs at function exit (spec 11)
  cleanup_demo()
end

def cleanup_demo
  puts "Starting cleanup demo"
  defer puts("Deferred: runs last")
  puts "Middle of function"
  puts "End of function body"
end
