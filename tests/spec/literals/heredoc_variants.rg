#@ run-pass
#@ check-output

# Test: Heredoc variants (Section 4.5)

# Basic heredoc - delimiter at line start
basic = <<END
Line one
Line two
END
puts "Basic has content: #{basic.contains?("Line one")}"

# Squiggly heredoc (<<~) - strips common leading whitespace
def get_sql: String
  <<~SQL
    SELECT *
    FROM users
    WHERE active = true
  SQL
end

sql = get_sql()
puts "SQL starts with SELECT: #{sql.start_with?("SELECT")}"

# Heredoc with different delimiter
html = <<HTML
<div>
  <p>Hello</p>
</div>
HTML
puts "HTML contains div: #{html.contains?("<div>")}"

# Heredoc with interpolation (default)
name = "World"
greeting = <<MSG
Hello #{name}!
Welcome.
MSG
puts "Greeting contains World: #{greeting.contains?("World")}"

# Literal heredoc (single-quoted delimiter) - no interpolation
template = <<'TEMPLATE'
Hello #{name}!
This is literal.
TEMPLATE
# Check that #{name} was NOT interpolated (still contains the literal text)
puts "Template has Hello: #{template.start_with?("Hello")}"
puts "Template length: #{template.length}"

#@ expect:
# Basic has content: true
# SQL starts with SELECT: true
# HTML contains div: true
# Greeting contains World: true
# Template has Hello: true
# Template length: 32
