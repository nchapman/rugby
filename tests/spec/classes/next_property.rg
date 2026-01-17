#@ run-pass
#@ check-output
#
# Test using 'next' as a property name (like Ruby allows)

class Node
  property next: Node?
  property value: Int
  @@next = 0  # Class variable named 'next' also works

  def initialize(value: Int)
    @value = value
    @next = nil
    @@next += 1
  end

  def self.count: Int
    @@next
  end
end

def main
  n1 = Node.new(1)
  n2 = Node.new(2)
  n3 = Node.new(3)

  n1.next = n2
  n2.next = n3

  # Access the linked list
  puts n1.value
  if let n2_node = n1.next
    puts n2_node.value
    if let n3_node = n2_node.next
      puts n3_node.value
    end
  end

  # Test class variable @@next
  puts Node.count
end

#@ expect:
# 1
# 2
# 3
# 3
