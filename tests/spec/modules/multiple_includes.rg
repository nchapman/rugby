#@ run-pass
#@ check-output
#
# Test multiple module includes in a class

module Greetable
  def greet -> String
    "Hello!"
  end
end

module Printable
  def info -> String
    "A printable object"
  end
end

class Thing
  include Greetable
  include Printable

  getter name : String

  def initialize(@name : String)
  end
end

thing = Thing.new("Widget")
puts thing.name
puts thing.greet
puts thing.info

#@ expect:
# Widget
# Hello!
# A printable object
