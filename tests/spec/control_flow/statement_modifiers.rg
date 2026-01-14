#@ run-pass
#@ check-output
#
# Test statement modifiers (if/unless at end of line)

# Test simple if modifier
x = 5
puts "x is big" if x > 3
puts "x is small" if x < 3

# Test unless modifier
valid = true
puts "invalid!" unless valid
puts "still checking" unless false

# Test return with if modifier
def check_positive(n : Int) -> String
  return "negative" if n < 0
  return "zero" if n == 0
  "positive"
end

puts check_positive(-5)
puts check_positive(0)
puts check_positive(10)

# Test return with unless modifier
def check_valid(s : String) -> String
  return "empty" unless s.length > 0
  "has content"
end

puts check_valid("")
puts check_valid("hello")

#@ expect:
# x is big
# still checking
# negative
# zero
# positive
# empty
# has content
