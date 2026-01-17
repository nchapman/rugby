#@ run-pass
#@ check-output
#
# Test: Section 10.8 - Symbol-to-proc (&:method)

class User
  getter name : String
  getter active : Bool

  def initialize(@name : String, @active : Bool)
  end

  def active? -> Bool
    @active
  end
end

users = [
  User.new("Alice", true),
  User.new("Bob", false),
  User.new("Charlie", true)
]

# &:name creates -> { |u| u.name }
names = users.map(&:name)
puts names[0]
puts names[1]
puts names[2]

# Works with predicates
active = users.select(&:active?)
puts active.length

#@ expect:
# Alice
# Bob
# Charlie
# 2
