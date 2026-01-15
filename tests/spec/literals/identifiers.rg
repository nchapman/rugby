#@ run-pass
#@ check-output
#
# Test: Section 4.2 - Identifiers
# snake_case for functions/methods/variables
# CamelCase for types
# predicate? methods return Bool
# setter= methods for assignment

class User
  getter name : String
  getter active : Bool

  def initialize(@name : String, @active : Bool)
  end

  def active? -> Bool
    @active
  end
end

# snake_case variables
user_name = "alice"
total_count = 42
puts user_name
puts total_count

# Predicate methods
user = User.new("Bob", true)
puts user.active?

# Instance variables via getter
puts user.name

#@ expect:
# alice
# 42
# true
# Bob
