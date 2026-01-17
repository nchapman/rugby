#@ run-pass
#@ check-output
#
# Test: Section 6.4 - Type constraints

# Constrain to types that are ordered (can use < > operators)
def max_of<T: Ordered>(a: T, b: T): T
  a > b ? a: b
end

puts max_of(10, 20)
puts max_of(3.14, 2.71)

# Constrain to custom interface
interface Printable
  def to_string: String
end

class Item
  def initialize(@name: String)
  end

  def to_string: String
    @name
  end
end

def print_it<T: Printable>(item: T)
  puts item.to_string
end

item = Item.new("Widget")
print_it(item)

#@ expect:
# 20
# 3.14
# Widget
