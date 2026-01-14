# Rugby Ranges
# Demonstrates: range literals, iteration, methods, slicing

def main
  # Inclusive range (1..5 includes 5)
  puts "Inclusive 1..5:"
  for i in 1..5
    puts "  #{i}"
  end

  # Exclusive range (1...5 excludes 5)
  puts "Exclusive 1...5:"
  for i in 1...5
    puts "  #{i}"
  end

  # Range methods
  r = 1..10
  puts "Range: #{r}"
  puts "Size: #{r.size}"
  puts "include?(5): #{r.include?(5)}"
  puts "include?(15): #{r.include?(15)}"

  # Range to array
  arr = (1..5).to_a
  puts "To array: #{arr}"

  # Range each with block
  puts "Range each:"
  (1..3).each { |i| puts "  #{i}" }

  # Array slicing with ranges
  letters = %w{a b c d e}
  puts "Array: #{letters}"
  puts "letters[1..3]: #{letters[1..3]}"
  puts "letters[0...2]: #{letters[0...2]}"
  puts "letters[-2..-1]: #{letters[-2..-1]}"

  # String slicing with ranges
  word = "Rugby"
  puts "String: #{word}"
  puts "word[0..2]: #{word[0..2]}"
  puts "word[1...4]: #{word[1...4]}"

  # Negative indices
  nums = [10, 20, 30, 40, 50]
  puts "Last: #{nums[-1]}"
  puts "Last 3: #{nums[-3..-1]}"
end
