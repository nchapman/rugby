#@ run-pass
#@ check-output
#
# Test returning a class instance when the return type is an interface

interface INode
  def value: Int
end

class Leaf
  pub getter value: Int

  def initialize(@value: Int)
  end
end

class Branch
  pub getter left: INode
  pub getter right: INode

  def initialize(@left: INode, @right: INode)
  end

  def value: Int
    @left.value + @right.value
  end
end

def make_leaf(n: Int): INode
  Leaf.new(n)
end

def make_tree(depth: Int): INode
  if depth <= 0
    return Leaf.new(1)
  end
  Branch.new(make_tree(depth - 1), make_tree(depth - 1))
end

def main
  leaf = make_leaf(42)
  puts leaf.value

  tree = make_tree(2)
  puts tree.value
end

#@ expect:
# 42
# 4
