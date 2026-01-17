#@ run-pass
#@ check-output
#
# Test: Section 11.3 - Accessors (getter, setter, property)

class Person
  getter age: Int           # read-only
  setter status: String     # write-only
  property name: String     # read-write

  def initialize(@name: String, @age: Int)
    @status = "active"
  end

  def get_status: String
    @status
  end
end

p = Person.new("Alice", 30)

# getter
puts p.name
puts p.age

# setter
p.status = "inactive"
puts p.get_status

# property (both)
p.name = "Bob"
puts p.name

#@ expect:
# Alice
# 30
# inactive
# Bob
