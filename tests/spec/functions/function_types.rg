#@ run-pass
#@ check-output
#
# Test: Section 10.9 - Function types

# Function type as parameter
def apply(x: Int, f : (Int): Int): Int
  f.(x)
end

double = -> { |n: Int| n * 2 }
triple = -> { |n: Int| n * 3 }

puts apply(5, double)
puts apply(5, triple)

# Type alias for function type
type Predicate = (Int): Bool

def filter_nums(arr: Array<Int>, pred: Predicate): Array<Int>
  arr.select -> { |n| pred.(n) }
end

evens = filter_nums([1, 2, 3, 4, 5], -> { |n| n % 2 == 0 })
puts evens.length
puts evens[0]
puts evens[1]

#@ expect:
# 10
# 15
# 2
# 2
# 4
