#@ run-pass
#@ check-output
#
# Test super calls with arguments
#
# Super calls should pass arguments to parent class methods.
# This tests both constructor super and method super.

class Animal
  getter name: String
  getter sound: String

  def initialize(@name: String, @sound: String)
  end

  def speak: String
    "#{@name} says #{@sound}"
  end

  def greet(greeting: String): String
    "#{greeting}, I'm #{@name}"
  end
end

class Dog < Animal
  getter breed: String

  # Super in constructor with args
  def initialize(name: String, @breed: String)
    super(name, "woof")
  end

  # Super in method (no args - uses same args as current method)
  def greet(greeting: String): String
    result = super(greeting)
    "#{result} the #{@breed}"
  end
end

dog = Dog.new("Rex", "German Shepherd")
puts dog.name
puts dog.sound
puts dog.breed
puts dog.speak
puts dog.greet("Hello")

#@ expect:
# Rex
# woof
# German Shepherd
# Rex says woof
# Hello, I'm Rex the German Shepherd
