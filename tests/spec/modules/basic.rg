#@ run-pass
#@ check-output
#
# Test module definition and include

module Greetable
  def greet -> String
    "Hello!"
  end

  def farewell -> String
    "Goodbye!"
  end
end

class Person
  include Greetable

  getter name : String

  def initialize(@name : String)
  end
end

person = Person.new("Alice")
puts person.name
puts person.greet
puts person.farewell

#@ expect:
# Alice
# Hello!
# Goodbye!
