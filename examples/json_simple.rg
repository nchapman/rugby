# Example: JSON parsing and generation

import rugby/json

def main
  # Parse JSON string into a map
  text = '{"name": "Alice", "age": 30, "active": true}'
  data = json.Parse(text)!

  puts "Name: #{data["name"]}"
  puts "Age: #{data["age"]}"

  # Parse a JSON array
  arr_text = '[1, 2, 3, 4, 5]'
  numbers = json.ParseArray(arr_text)!
  puts "Numbers: #{numbers}"

  # Generate JSON from a map (maps must have consistent value types)
  user = {"name": "Bob", "role": "admin"}
  output = json.Generate(user)!
  puts "Generated: #{output}"

  # Pretty-print JSON
  pretty = json.Pretty(user)!
  puts "Pretty:\n#{pretty}"

  # Check if string is valid JSON
  if json.Valid(text)
    puts "Valid JSON!"
  end
end
