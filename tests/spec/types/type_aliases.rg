#@ run-pass
#@ check-output
#
# Test type aliases (Section 5.3)

# Basic type aliases
type UserID = Int64
type Name = String

# Alias for complex types
type Strings = Array[String]

# Chained aliases
type ID = UserID

# Public alias
pub type PublicID = Int64

def greet(name : Name) -> String
  "Hello, #{name}"
end

def main
  # Use basic aliases
  id : UserID = 42
  name : Name = "Alice"

  puts id
  puts name
  puts greet(name)

  # Use alias for array type
  names : Strings = ["Bob", "Carol"]
  names.each -> (n) { puts n }

  # Use chained alias
  second_id : ID = 100
  puts second_id
end

#@ expect:
# 42
# Alice
# Hello, Alice
# Bob
# Carol
# 100
