#@ run-pass
#@ check-output
# Empty hash literal type inference from property type annotation

class Node
  property val: Int

  def initialize(@val: Int)
  end
end

class Container
  property cache: Hash<Int, Node>
  property counts: Hash<String, Int>

  def initialize
    @cache = {}
    @counts = {}
  end
end

def main
  c = Container.new

  # Assign to hash with class value type
  c.cache[1] = Node.new(42)
  puts c.cache[1].val

  # Assign to hash with primitive value type
  c.counts["hello"] = 5
  puts c.counts["hello"]
end

#@ expect:
# 42
# 5
