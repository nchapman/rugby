#@ run-pass
#@ check-output
#
# Test regex literals and match operators

# Basic regex match
text = "hello world"
if text =~ /hello/
  puts "matched hello"
end

# Regex not match
if text !~ /goodbye/
  puts "no goodbye"
end

# Regex with flags (case insensitive)
if "HELLO" =~ /hello/i
  puts "case insensitive match"
end

# Regex in variable
pattern = /\d+/
number_text = "abc123def"
if number_text =~ pattern
  puts "found digits"
end

# Regex with special characters
email = "test@example.com"
if email =~ /\w+@\w+\.\w+/
  puts "valid email pattern"
end

#@ expect:
# matched hello
# no goodbye
# case insensitive match
# found digits
# valid email pattern
