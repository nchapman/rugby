# JSON Parsing and Generation
# Demonstrates: rugby/json stdlib, error handling with !

import rugby/json

def main
  # Parse JSON string into a map
  text = '{"name": "Alice", "age": 30, "active": true}'
  data = json.parse(text)!

  puts "Name: #{data["name"]}"
  puts "Age: #{data["age"]}"

  # Parse a JSON array
  arr_text = "[1, 2, 3, 4, 5]"
  numbers = json.parse_array(arr_text)!
  puts "Numbers: #{numbers}"

  # Generate JSON from a map
  user = {name: "Bob", role: "admin"}
  output = json.generate(user)!
  puts "Generated: #{output}"

  # Pretty-print JSON
  pretty = json.pretty(user)!
  puts "Pretty:\n#{pretty}"

  # Check if string is valid JSON
  if json.valid?(text)
    puts "Valid JSON!"
  end
end
