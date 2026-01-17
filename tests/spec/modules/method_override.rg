#@ run-pass
#@ check-output
#
# Test module method resolution and overriding
#
# When a class includes a module and defines its own method with the
# same name, the class method should take precedence.

module Greetable
  def greet: String
    "Hello from module!"
  end

  def farewell: String
    "Goodbye from module!"
  end
end

class Person
  include Greetable

  getter name: String

  def initialize(@name: String)
  end

  # Override the greet method from module
  def greet: String
    "Hello, I'm #{@name}!"
  end
end

person = Person.new("Alice")

# This should use the class's overridden method
puts person.greet

# This should use the module's method
puts person.farewell

#@ expect:
# Hello, I'm Alice!
# Goodbye from module!
