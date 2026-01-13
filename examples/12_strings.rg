# Rugby Strings
# Demonstrates: string methods, interpolation, heredocs (spec 3.3, 3.4, 12.5)

def main
  s = "  Hello, World!  "

  # Basic info
  puts "Original: '#{s}'"
  puts "Length (bytes): #{s.length}"

  # Transformation methods (spec 12.5)
  puts "Upcase: #{s.upcase}"
  puts "Downcase: #{s.downcase}"
  puts "Strip: '#{s.strip}'"

  # Query methods
  greeting = "Hello, World!"
  puts "Contains 'World': #{greeting.contains?("World")}"
  puts "Starts with 'Hello': #{greeting.start_with?("Hello")}"
  puts "Ends with '!': #{greeting.end_with?("!")}"
  puts "Empty?: #{greeting.empty?}"

  # Splitting
  csv = "apple,banana,cherry"
  parts = csv.split(",")
  puts "Split: #{parts}"
  puts "First part: #{parts[0]}"

  # Replace
  text = "foo bar foo"
  puts "Replace foo->baz: #{text.replace("foo", "baz")}"

  # Lines
  multiline = "line1\nline2\nline3"
  lines = multiline.lines
  puts "Line count: #{lines.length}"

  # Characters
  word = "Rugby"
  chars = word.chars
  puts "Chars: #{chars}"

  # Indexing with negative indices (spec 4.2)
  puts "word[0]: #{word[0]}"
  puts "word[-1]: #{word[-1]}"

  # Range slicing
  puts "word[0..2]: #{word[0..2]}"

  # Heredoc basic (spec 3.4)
  doc = <<END
This is a heredoc.
It preserves lines.
END
  puts "Heredoc:"
  puts doc

  # Heredoc with indented delimiter <<- (spec 3.4)
  msg = <<-MSG
Hello there!
  MSG
  puts "Heredoc <<-:"
  puts msg

  # Heredoc with indent stripping <<~ (spec 3.4)
  sql = <<~SQL
    SELECT *
    FROM users
    WHERE active = true
  SQL
  puts "SQL Query:"
  puts sql

  # String to number conversion (spec 12.5)
  num_str = "42"
  n, err = num_str.to_i
  if err == nil
    puts "Parsed: #{n}"
  end

  float_str = "3.14"
  f, err2 = float_str.to_f
  if err2 == nil
    puts "Parsed float: #{f}"
  end
end
