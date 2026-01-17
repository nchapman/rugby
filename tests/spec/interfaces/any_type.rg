#@ run-pass
#@ check-output
#
# Test: Section 13.5 - The Any type

def debug(value: Any)
  puts value
end

debug(42)
debug("hello")
debug(true)
debug(3.14)

# Array of Any
items: Array<Any> = [1, "two", 3.0]
puts items.length

#@ expect:
# 42
# hello
# true
# 3.14
# 3
