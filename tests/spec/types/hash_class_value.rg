#@ run-pass
#@ check-output
#
# Hash with class value type - indexing returns pointer
# For class-valued hashes, direct indexing returns nil for missing keys
# No .unwrap needed - the pointer IS the value

class Node
  property value: Int

  def initialize(@value: Int)
  end
end

class Container
  property data: Hash<Int, Node>

  def initialize
    @data = {}
  end

  def put(key: Int, value: Int)
    @data[key] = Node.new(value)
  end

  def get(key: Int): (Int, Bool)
    node = @data[key]
    if node == nil
      return 0, false
    end
    # No .unwrap needed - node is already *Node (the class pointer)
    return node.value, true
  end
end

def main
  c = Container.new
  c.put(1, 100)
  c.put(2, 200)

  val1, ok1 = c.get(1)
  puts ok1
  puts val1

  val99, ok99 = c.get(99)
  puts ok99
end

#@ expect:
# true
# 100
# false
