import "os"
import "fmt"
import "strconv"

interface INode
  def cal_hash
  def has_hash: Bool
  def get_hash: Int
  def check: Bool
end

class Leaf
  property hash: Int?
  property value: Int

  def initialize(@value: Int)
    @hash = nil
  end

  def cal_hash
    if @hash == nil
      @hash = @value
    end
  end

  def has_hash: Bool
    @hash != nil
  end

  def get_hash: Int
    @hash.unwrap_or(-1)
  end

  def check: Bool
    self.has_hash
  end
end

class Node
  property hash: Int?
  property left: INode
  property right: INode

  def initialize(@left: INode, @right: INode)
    @hash = nil
  end

  def cal_hash
    if @hash == nil
      @left.cal_hash
      @right.cal_hash
      @hash = @left.get_hash + @right.get_hash
    end
  end

  def has_hash: Bool
    @hash != nil
  end

  def get_hash: Int
    @hash.unwrap_or(-1)
  end

  def check: Bool
    if self.has_hash
      return @left.check && @right.check
    end
    false
  end
end

def make_tree(depth: Int): INode
  if depth <= 0
    return Leaf.new(1)
  end
  d = depth - 1
  Node.new(make_tree(d), make_tree(d))
end

def main
  min_depth = 4
  max_depth = 5
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    max_depth = arg
  end

  if min_depth + 2 > max_depth
    max_depth = min_depth + 2
  end
  stretch_depth = max_depth + 1

  stretch_tree = make_tree(stretch_depth)
  stretch_tree.cal_hash
  fmt.Printf("stretch tree of depth %d\t root hash: %d check: %t\n", stretch_depth, stretch_tree.get_hash, stretch_tree.check)

  long_lived_tree = make_tree(max_depth)

  depth = min_depth
  while depth <= max_depth
    iterations = 1 << (max_depth - depth + min_depth)
    sum = 0
    i = 1
    while i <= iterations
      tree = make_tree(depth)
      tree.cal_hash
      sum += tree.get_hash
      i += 1
    end
    fmt.Printf("%d\t trees of depth %d\t root hash sum: %d\n", iterations, depth, sum)
    depth += 2
  end

  long_lived_tree.cal_hash
  fmt.Printf("long lived tree of depth %d\t root hash: %d check: %t\n", max_depth, long_lived_tree.get_hash, long_lived_tree.check)
end
