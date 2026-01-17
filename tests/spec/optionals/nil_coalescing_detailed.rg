#@ run-pass
#@ check-output
#
# Test: Section 8.1 - Nil coalescing operator (??)

def find_user(id: Int): String?
  if id == 1
    return "Alice"
  end
  nil
end

# Basic nil coalescing
name1 = find_user(1) ?? "Anonymous"
puts name1

name2 = find_user(99) ?? "Anonymous"
puts name2

# ?? checks presence, not truthiness
# 0 IS present, so ?? doesn't replace it
count: Int? = 0
puts count ?? 10

# Empty string IS present
str: String? = ""
puts str ?? "default"
puts "got empty" if str == ""

# nil triggers the default
missing: Int? = nil
puts missing ?? 42

#@ expect:
# Alice
# Anonymous
# 0
#
# got empty
# 42
