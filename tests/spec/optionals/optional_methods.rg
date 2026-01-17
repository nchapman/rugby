#@ run-pass
#@ check-output
#
# Test: Section 8.2 - Optional methods

def find_value(id: Int): Int?
  if id == 1
    return 42
  end
  nil
end

# ok? / present?
val1 = find_value(1)
puts val1.ok?

val2 = find_value(99)
puts val2.ok?

# nil? / absent?
puts val2.nil?

# unwrap_or
puts val1.unwrap_or(0)
puts val2.unwrap_or(0)

#@ expect:
# true
# false
# true
# 42
# 0
