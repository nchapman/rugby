# Rugby Functions
# Demonstrates: def, return types, multiple returns, implicit return (spec 6)

# Simple function with implicit return (last expression)
def double(n : Int) -> Int
  n * 2
end

# Explicit return with statement modifier
def abs(n : Int) -> Int
  return -n if n < 0
  n
end

# Multiple return values (spec 6)
def divmod(a : Int, b : Int) -> (Int, Int)
  return a / b, a % b
end

# Function returning optional (spec 4.4)
def find_user(id : Int) -> String?
  return "Alice" if id == 1
  return "Bob" if id == 2
  nil
end

# Function with array parameter
def sum_all(nums : Array[Int]) -> Int
  total = 0
  for n in nums
    total += n
  end
  total
end

def main
  # Call simple functions
  puts "double(5) = #{double(5)}"
  puts "abs(-7) = #{abs(-7)}"
  puts "abs(3) = #{abs(3)}"

  # Multiple return values with destructuring
  q, r = divmod(17, 5)
  puts "17 / 5 = #{q} remainder #{r}"

  # Optional return with ?? (spec 4.4.1)
  name = find_user(1) ?? "Anonymous"
  puts "User 1: #{name}"

  missing = find_user(99) ?? "Anonymous"
  puts "User 99: #{missing}"

  # Array argument
  nums = [1, 2, 3, 4, 5]
  puts "sum([1,2,3,4,5]) = #{sum_all(nums)}"
end
