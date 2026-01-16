#@ run-pass
#@ check-output

# Test optional.flat_map method
# flat_map { |x| f(x) } where f returns R?, result is R?
# This flattens nested optionals

arr = [1, 2, 3, 4, 5]

# Test 1: flat_map on Some value with block returning Some
name : String? = "Alice"
result1 = name.flat_map do |n|
  arr.find do |x|
    x > 3
  end
end
if let r = result1
  puts "1: #{r}"
else
  puts "1: none"
end

# Test 2: flat_map on Some value with block returning None
name2 : String? = "Bob"
result2 = name2.flat_map do |n|
  arr.find do |x|
    x > 10  # Nothing > 10 in array
  end
end
if let r = result2
  puts "2: #{r}"
else
  puts "2: none"
end

# Test 3: flat_map on None value
name3 : String? = nil
result3 = name3.flat_map do |n|
  arr.find do |x|
    x > 0  # Would find something if there was input
  end
end
if let r = result3
  puts "3: #{r}"
else
  puts "3: none"
end

# Test 4: Chained flat_map
opt1 : Int? = 5
result4 = opt1.flat_map do |n|
  arr.find do |x|
    x == n  # Find the value in array
  end
end
if let r = result4
  puts "4: #{r}"
else
  puts "4: none"
end

#@ expect:
# 1: 4
# 2: none
# 3: none
# 4: 5
