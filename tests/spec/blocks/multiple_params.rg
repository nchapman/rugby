#@ run-pass
#@ check-output
#
# Test lambdas with multiple parameters
#
# Some iteration methods pass multiple values to the lambda.

# each_with_index passes element and index
names = ["Alice", "Bob", "Charlie"]
names.each_with_index -> (name, i) do
  puts "#{i}: #{name}"
end

# reduce passes accumulator and element
numbers = [1, 2, 3, 4, 5]
sum = numbers.reduce(0) -> (acc, n) do
  acc + n
end
puts sum

# Map hash entries - verify each passes key and value
# (output deliberately suppressed since Go map order is non-deterministic)
scores = {"Alice" => 100, "Bob" => 85}
count = 0
scores.each -> (name, score) do
  # Use both variables to avoid unused warning
  if score > 0
    if name.length > 0
      count += 1
    end
  end
end
puts "Iterated #{count} entries"

#@ expect:
# 0: Alice
# 1: Bob
# 2: Charlie
# 15
# Iterated 2 entries
