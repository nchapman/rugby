#@ run-pass
#@ check-output
#
# Test: Section 5.3 - Type aliases

# Simple type alias
type UserID = Int64

id: UserID = 12345
puts id

# Type alias for map
type StringMap = Map<String, String>

m: StringMap = {"key" => "value"}
puts m["key"]

# Type alias for function type
type Handler = (Int): Int

double: Handler = -> { |x: Int| x * 2 }
puts double.(5)

#@ expect:
# 12345
# value
# 10
