# Rugby Functions
# Demonstrates: def, return types, multiple returns, implicit return

# Simple function - last expression is the return value
def double(n : Int) -> Int
  n * 2
end

# Explicit return with statement modifier
def abs(n : Int) -> Int
  return -n if n < 0
  n
end

# Multiple return values
def divmod(a : Int, b : Int) -> (Int, Int)
  return a / b, a % b
end

# Function returning optional
def find_user(id : Int) -> String?
  return "Alice" if id == 1
  return "Bob" if id == 2
  nil
end

# Function with array parameter - use blocks for transformation
def sum_all(nums : Array[Int]) -> Int
  nums.sum  # idiomatic: use built-in method
end

# Or use reduce for custom aggregation
def product_all(nums : Array[Int]) -> Int
  nums.reduce(1) { |acc, n| acc * n }
end

def main
  # Call simple functions
  puts "double(5) = #{double(5)}"
  puts "abs(-7) = #{abs(-7)}"
  puts "abs(3) = #{abs(3)}"

  # Multiple return values with destructuring
  q, r = divmod(17, 5)
  puts "17 / 5 = #{q} remainder #{r}"

  # Ignore values with _
  q2, _ = divmod(10, 3)
  puts "10 / 3 = #{q2}"

  # Optional return with ??
  name = find_user(1) ?? "Anonymous"
  puts "User 1: #{name}"

  missing = find_user(99) ?? "Anonymous"
  puts "User 99: #{missing}"

  # Array functions
  nums = [1, 2, 3, 4, 5]
  puts "sum([1,2,3,4,5]) = #{sum_all(nums)}"
  puts "product([1,2,3,4,5]) = #{product_all(nums)}"
end
