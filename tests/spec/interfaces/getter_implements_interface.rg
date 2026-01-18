#@ run-pass
#@ check-output

# Test that getters can satisfy interface methods
# When called through an interface, it should use method dispatch (not direct field access)

interface Named
  def name: String
end

interface Describable
  def name: String
  def description: String
end

pub class User implements Named
  getter name: String

  def initialize(@name: String)
  end
end

pub class Product implements Describable
  getter name: String
  property description: String

  def initialize(@name: String, @description: String)
  end
end

def greet(obj: Named)
  puts "Hello, " + obj.name
end

def show_info(obj: Describable)
  puts obj.name + ": " + obj.description
end

# Direct access (uses direct field access optimization)
user = User.new("Alice")
puts user.name

# Through interface (uses method dispatch)
greet(user)

# Multiple getters implementing interface
product = Product.new("Widget", "A useful thing")
puts product.name
puts product.description

# Through interface
show_info(product)

#@ expect:
# Alice
# Hello, Alice
# Widget
# A useful thing
# Widget: A useful thing
