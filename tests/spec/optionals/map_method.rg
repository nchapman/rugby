#@ run-pass
#@ check-output

# Test: Optional map method (Section 8.2)
# map transforms the value if present, returns nil if absent

def find_name(id: Int): String?
  return nil if id < 0
  "User#{id}"
end

# map on present value
name = find_name(42)
upper = name.map -> { |n| n.upcase }
puts upper.unwrap_or("NONE")

# map on nil value
missing = find_name(-1)
upper_missing = missing.map -> { |n| n.upcase }
puts upper_missing.unwrap_or("NONE")

# Chained maps
result = find_name(5)
  .map -> { |n| n.upcase }
  .map -> { |n| "Hello, #{n}!" }
puts result.unwrap_or("No greeting")

# map with type transformation
num: Int? = 123
str = num.map -> { |n| "num#{n}" }
puts str.unwrap_or("no number")

# map vs flat_map distinction
# map wraps result in optional: T? -> R?
# flat_map expects lambda to return optional: T? -> R?

def lookup(n: Int): String?
  return nil if n < 0
  "Item#{n}"
end

# Using map (would return String??)
opt_id: Int? = 10
# result = opt_id.map -> { |id| lookup(id) }  # This would be String??

# Using flat_map for optional-returning lambdas
result2 = opt_id.flat_map -> { |id| lookup(id) }
puts result2.unwrap_or("not found")

#@ expect:
# USER42
# NONE
# Hello, USER5!
# num123
# Item10
