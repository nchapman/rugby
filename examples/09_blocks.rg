# Rugby Blocks
# Demonstrates: do...end, {}, iteration methods, symbol-to-proc (spec 5.4, 12.3)

def main
  nums = [1, 2, 3, 4, 5]

  # each with do...end block
  puts "Each with do...end:"
  nums.each do |n|
    puts "  #{n}"
  end

  # each with {} block (single-line preferred)
  puts "Each with {}:"
  nums.each { |n| puts "  #{n}" }

  # map - transform elements (spec 12.3)
  doubled = nums.map { |n| n * 2 }
  puts "Doubled: #{doubled}"

  # select/filter - keep matching elements
  evens = nums.select { |n| n % 2 == 0 }
  puts "Evens: #{evens}"

  # reject - remove matching elements
  odds = nums.reject { |n| n % 2 == 0 }
  puts "Odds: #{odds}"

  # find - first matching element (returns optional)
  first_even = nums.find { |n| n % 2 == 0 }
  puts "First even: #{first_even ?? 0}"

  # reduce - accumulate into single value
  sum = nums.reduce(0) { |acc, n| acc + n }
  puts "Sum: #{sum}"

  product = nums.reduce(1) { |acc, n| acc * n }
  puts "Product: #{product}"

  # Predicate methods
  has_even = nums.any? { |n| n % 2 == 0 }
  puts "Has even: #{has_even}"

  all_positive = nums.all? { |n| n > 0 }
  puts "All positive: #{all_positive}"

  none_negative = nums.none? { |n| n < 0 }
  puts "None negative: #{none_negative}"

  # include?/contains? (spec 12.3)
  puts "Contains 3? #{nums.include?(3)}"
  puts "Contains 10? #{nums.contains?(10)}"

  # each_with_index
  puts "With index:"
  nums.each_with_index do |n, i|
    puts "  [#{i}] = #{n}"
  end

  # Array utility methods (spec 12.3)
  puts "Sum: #{nums.sum}"
  puts "Min: #{nums.min ?? 0}"
  puts "Max: #{nums.max ?? 0}"
  puts "First: #{nums.first ?? 0}"
  puts "Last: #{nums.last ?? 0}"

  # Append with << (spec 12.3)
  arr = [1, 2]
  arr = arr << 3
  puts "Appended: #{arr}"

  # Sorting (spec 12.3)
  unsorted = [3, 1, 4, 1, 5]
  sorted = unsorted.sorted
  puts "Sorted (copy): #{sorted}"
  puts "Original: #{unsorted}"

  # Integer times (spec 12.6)
  puts "Three times:"
  3.times do |i|
    puts "  Iteration #{i}"
  end

  # upto/downto
  puts "1 upto 3:"
  1.upto(3) { |i| puts "  #{i}" }

  puts "3 downto 1:"
  3.downto(1) { |i| puts "  #{i}" }

  # Symbol-to-proc shorthand (spec 5.4)
  words = ["hello", "world"]
  upper = words.map(&:upcase)
  puts "Uppercase: #{upper}"

  # Map iteration methods (spec 12.4)
  puts "Map methods:"
  m = {"a" => 1, "b" => 2, "c" => 3}
  m.each { |k, v| puts "  #{k}: #{v}" }
  puts "Keys: #{m.keys}"
  puts "Values: #{m.values}"
  puts "Has 'a'? #{m.has_key?("a")}"
  puts "Fetch 'z' with default: #{m.fetch("z", 0)}"
end
