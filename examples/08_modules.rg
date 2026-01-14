# Rugby Modules (Mixins)
# Demonstrates: module, include

# Module with reusable methods
module Greetable
  def greet -> String
    "Hello!"
  end

  def farewell -> String
    "Goodbye!"
  end
end

module Debuggable
  def debug -> String
    "Debug info"
  end
end

# Class including a module
class Greeter
  include Greetable

  def initialize(@name : String)
  end

  def personalized_greet -> String
    "#{greet} I'm #{@name}."
  end
end

# Class including multiple modules
class Service
  include Greetable
  include Debuggable

  def initialize(@name : String)
  end

  def status -> String
    "#{@name}: #{greet} #{debug}"
  end
end

def main
  greeter = Greeter.new("Alice")
  puts greeter.greet
  puts greeter.farewell
  puts greeter.personalized_greet

  svc = Service.new("API")
  puts svc.greet
  puts svc.debug
  puts svc.status
end
