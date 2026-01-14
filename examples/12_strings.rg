# Rugby Strings
# Demonstrates: string methods, interpolation, heredocs, conversion

def main
  s = "  Hello, World!  "

  puts "Original: '#{s}'"
  puts "Length: #{s.length}"

  # Transformation
  puts "upcase: #{s.upcase}"
  puts "downcase: #{s.downcase}"
  puts "strip: '#{s.strip}'"

  # Query methods
  greeting = "Hello, World!"
  puts "contains? 'World': #{greeting.contains?("World")}"
  puts "start_with? 'Hello': #{greeting.start_with?("Hello")}"
  puts "end_with? '!': #{greeting.end_with?("!")}"
  puts "empty?: #{greeting.empty?}"

  # Splitting and joining
  csv = "apple,banana,cherry"
  parts = csv.split(",")
  puts "Split: #{parts}"

  # Replace
  text = "foo bar foo"
  puts "Replace: #{text.replace("foo", "baz")}"

  # Characters and indexing
  word = "Rugby"
  puts "Chars: #{word.chars}"
  puts "word[0]: #{word[0]}, word[-1]: #{word[-1]}"
  puts "word[0..2]: #{word[0..2]}"

  # Heredoc basic
  doc = <<END
This is a heredoc.
It preserves lines.
END
  puts "Heredoc:"
  puts doc

  # Heredoc with indent stripping (most useful)
  sql = <<~SQL
    SELECT *
    FROM users
    WHERE active = true
  SQL
  puts "SQL Query:"
  puts sql

  # Heredoc with indented delimiter
  msg = <<-MSG
Hello there!
  MSG
  puts "<<- heredoc:"
  puts msg

  # String to number (returns tuple with error)
  n = "42".to_i rescue 0
  f = "3.14".to_f rescue 0.0
  puts "Parsed: #{n}, #{f}"
end
