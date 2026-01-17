#@ run-pass
#@ check-output

# Test: Optional unwrap method (Section 8.2)
# unwrap returns value if present, panics if nil

# unwrap on present value
def find_number(id: Int): Int?
  return nil if id < 0
  id * 10
end

result = find_number(5)
puts result.unwrap

# unwrap_or with default
missing = find_number(-1)
puts missing.unwrap_or(999)

# Chaining with unwrap_or
value = find_number(-1).unwrap_or(0) + 100
puts value

# unwrap after checking ok?
opt = find_number(3)
if opt.ok?
  puts opt.unwrap
end

# unwrap_or with default expression
missing2: Int? = nil
result2 = missing2.unwrap_or(42)
puts result2

#@ expect:
# 50
# 999
# 100
# 30
# 42
