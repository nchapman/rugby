import "os"
import "fmt"
import "strconv"

class LCG
  property seed: Int

  def initialize(@seed: Int)
  end

  def next(a: Int, c: Int, m: Int): Int
    @seed = (@seed * a + c) % m
    @seed
  end
end

# Simple LRU cache using a hash map and doubly-linked list
class LRUNode
  property key: Int
  property value: Int
  property prev: LRUNode?
  property next: LRUNode?

  def initialize(@key: Int, @value: Int)
    @prev = nil
    @next = nil
  end
end

class LRU
  property size: Int
  property cache: Hash<Int, LRUNode>
  property head: LRUNode?
  property tail: LRUNode?
  property count: Int

  def initialize(@size: Int)
    @cache = {}
    @head = nil
    @tail = nil
    @count = 0
  end

  def get(key: Int): (Int, Bool)
    node = @cache[key]
    if node == nil
      return 0, false
    end
    self.move_to_end(node.unwrap)
    return node.unwrap.value, true
  end

  def put(key: Int, value: Int)
    node = @cache[key]
    if node != nil
      node.unwrap.value = value
      self.move_to_end(node.unwrap)
      return
    end

    if @count == @size
      self.evict
    end

    new_node = LRUNode.new(key, value)
    @cache[key] = new_node
    self.add_to_end(new_node)
    @count += 1
  end

  def move_to_end(node: LRUNode)
    self.remove_node(node)
    self.add_to_end(node)
  end

  def remove_node(node: LRUNode)
    if node.prev != nil
      node.prev.unwrap.next = node.next
    else
      @head = node.next
    end

    if node.next != nil
      node.next.unwrap.prev = node.prev
    else
      @tail = node.prev
    end
  end

  def add_to_end(node: LRUNode)
    node.prev = @tail
    node.next = nil
    if @tail != nil
      @tail.unwrap.next = node
    end
    @tail = node
    if @head == nil
      @head = node
    end
  end

  def evict
    if @head != nil
      @cache.delete(@head.unwrap.key)
      self.remove_node(@head.unwrap)
      @count -= 1
    end
  end
end

def main
  a = 1103515245
  c = 12345
  m = 1 << 31

  size = 100
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    size = arg
  end

  n = 10000
  if os.Args.length > 2
    arg, _ = strconv.atoi(os.Args[2])
    n = arg
  end

  mod = size * 10
  rng0 = LCG.new(0)
  rng1 = LCG.new(1)
  hit = 0
  missed = 0
  lru = LRU.new(size)

  n.times -> {
    n0 = rng0.next(a, c, m) % mod
    lru.put(n0, n0)

    n1 = rng1.next(a, c, m) % mod
    _, ok = lru.get(n1)
    if ok
      hit += 1
    else
      missed += 1
    end
  }

  fmt.Printf("%d\n%d\n", hit, missed)
end
