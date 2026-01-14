#@ run-pass
#@ check-output
#
# Test that getters are inherited through multiple levels of inheritance

class Grandparent
  getter id : Int

  def initialize(@id : Int)
  end
end

class Parent < Grandparent
  getter name : String

  def initialize(id : Int, @name : String)
    @id = id
  end
end

class Child < Parent
  getter age : Int

  def initialize(id : Int, name : String, @age : Int)
    @id = id
    @name = name
  end
end

# Test that all inherited getters work
child = Child.new(1, "Alice", 10)
puts child.id
puts child.name
puts child.age

#@ expect:
# 1
# Alice
# 10
