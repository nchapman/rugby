#@ run-pass
#@ check-output
#
# Test: Section 10.2 - Multiple return values

def parse(s: String): (Int, Bool)
  if s.empty?
    return 0, false
  end
  42, true
end

# Destructuring both values
n, ok = parse("hello")
puts n
puts ok

# Ignore second with _
value, _ = parse("test")
puts value

# Check empty case
n2, ok2 = parse("")
puts n2
puts ok2

#@ expect:
# 42
# true
# 42
# 0
# false
