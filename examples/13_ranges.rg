# Rugby Ranges
# Demonstrates: range literals, iteration, methods, slicing (spec 4.2.1)

def main
  # Inclusive range (1..5 includes 5)
  puts "Inclusive range 1..5:"
  for i in 1..5
    puts "  #{i}"
  end

  # Exclusive range (1...5 excludes 5)
  puts "Exclusive range 1...5:"
  for i in 1...5
    puts "  #{i}"
  end

  # Range stored in variable
  r = 1..10
  puts "Range: #{r}"

  # Range methods (spec 4.2.1)
  puts "Size: #{r.size}"
  puts "Contains 5? #{r.include?(5)}"
  puts "Contains 15? #{r.include?(15)}"

  # Range to array
  arr = (1..5).to_a
  puts "To array: #{arr}"

  # Range each with block
  puts "Each:"
  (1..3).each do |i|
    puts "  #{i}"
  end

  # Array slicing with ranges (spec 4.2)
  letters = ["a", "b", "c", "d", "e"]
  puts "Array: #{letters}"
  puts "letters[1..3]: #{letters[1..3]}"
  puts "letters[0...2]: #{letters[0...2]}"
  puts "letters[-2..-1]: #{letters[-2..-1]}"

  # String slicing with ranges
  word = "Rugby"
  puts "String: #{word}"
  puts "word[0..2]: #{word[0..2]}"
  puts "word[1...4]: #{word[1...4]}"

  # Negative indices (spec 4.2)
  nums = [10, 20, 30, 40, 50]
  puts "Last element: #{nums[-1]}"
  puts "Last 3: #{nums[-3..-1]}"
end
