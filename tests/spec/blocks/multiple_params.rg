#@ run-pass
#@ check-output
#
# Test blocks with multiple parameters
#
# Some block methods pass multiple values to the block.

# each_with_index passes element and index
names = ["Alice", "Bob", "Charlie"]
names.each_with_index do |name, i|
  puts "#{i}: #{name}"
end

# reduce passes accumulator and element
numbers = [1, 2, 3, 4, 5]
sum = numbers.reduce(0) do |acc, n|
  acc + n
end
puts sum

# Map hash entries (if supported)
scores = {"Alice" => 100, "Bob" => 85}
scores.each do |name, score|
  puts "#{name} scored #{score}"
end

#@ expect:
# 0: Alice
# 1: Bob
# 2: Charlie
# 15
# Alice scored 100
# Bob scored 85
