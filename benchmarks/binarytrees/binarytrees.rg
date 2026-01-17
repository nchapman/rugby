import "os"
import "fmt"
import "strconv"

class Node
  getter left: Node?
  getter right: Node?

  def initialize(@left: Node?, @right: Node?)
  end

  def item_check: Int
    if @left == nil
      return 1
    end
    1 + @left.item_check + @right.item_check
  end
end

def bottom_up_tree(depth: Int): Node
  if depth <= 0
    return Node.new(nil, nil)
  end
  Node.new(bottom_up_tree(depth - 1), bottom_up_tree(depth - 1))
end

def main
  max_depth = 10
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    max_depth = arg
  end

  min_depth = 4
  if min_depth + 2 > max_depth
    max_depth = min_depth + 2
  end
  stretch_depth = max_depth + 1

  check = bottom_up_tree(stretch_depth).item_check
  fmt.Printf("stretch tree of depth %d\t check: %d\n", stretch_depth, check)

  long_lived_tree = bottom_up_tree(max_depth)

  depth = min_depth
  while depth <= max_depth
    iterations = 1 << (max_depth - depth + min_depth)
    check = 0

    iterations.times -> {
      check += bottom_up_tree(depth).item_check
    }

    fmt.Printf("%d\t trees of depth %d\t check: %d\n", iterations, depth, check)
    depth += 2
  end

  fmt.Printf("long lived tree of depth %d\t check: %d\n", max_depth, long_lived_tree.item_check)
end
