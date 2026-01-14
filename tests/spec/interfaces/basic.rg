#@ run-pass
#@ check-output
#
# Test: Interface structural typing

interface Greeter
  def greet -> String
end

class Person
  getter name : String

  def initialize(@name : String)
  end

  def greet -> String
    "Hello, I'm #{@name}"
  end
end

class Robot
  getter model : String

  def initialize(@model : String)
  end

  def greet -> String
    "Beep boop, I am #{@model}"
  end
end

def say_hello(g : Greeter)
  puts g.greet
end

say_hello(Person.new("Alice"))
say_hello(Robot.new("R2D2"))

#@ expect:
# Hello, I'm Alice
# Beep boop, I am R2D2
