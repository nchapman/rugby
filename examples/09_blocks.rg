# Rugby Lambdas
# Demonstrates: -> { |params| }, -> do |params| ... end, iteration, transformation, symbol-to-proc
#
# Lambdas are Rugby's bread and butter. Prefer them over manual loops.
# NOTE: return/break/next are NOT allowed inside lambdas (use loops for control flow).

def main
  nums = [1, 2, 3, 4, 5]

  # Braces for single-line lambdas
  puts "Doubled: #{nums.map -> { |n| n * 2 }}"

  # do...end for multi-line lambdas
  results = nums.map -> do |n|
    squared = n * n
    squared + 1
  end
  puts "Squared+1: #{results}"

  # Symbol-to-proc shorthand
  words = ["hello", "world"]
  puts "Uppercase: #{words.map(&:upcase)}"

  # Filtering
  evens = nums.select -> { |n| n.even? }
  odds = nums.reject -> { |n| n.even? }
  puts "Evens: #{evens}, Odds: #{odds}"

  # Short form with symbol-to-proc
  items = ["", "foo", "", "bar"]
  non_empty = items.reject(&:empty?)
  puts "Non-empty: #{non_empty}"

  # find returns an optional
  first_even = nums.find -> { |n| n.even? }
  puts "First even: #{first_even ?? 0}"

  # reduce for aggregation
  sum = nums.reduce(0) -> { |acc, n| acc + n }
  product = nums.reduce(1) -> { |acc, n| acc * n }
  puts "Sum: #{sum}, Product: #{product}"

  # Predicate methods
  puts "any even? #{nums.any? -> { |n| n.even? }}"
  puts "all positive? #{nums.all? -> { |n| n > 0 }}"
  puts "none negative? #{nums.none? -> { |n| n < 0 }}"

  # Built-in aggregations
  puts "Sum: #{nums.sum}, Min: #{nums.min ?? 0}, Max: #{nums.max ?? 0}"

  # each for side effects
  puts "Iteration:"
  nums.each -> { |n| puts "  #{n}" }

  # each_with_index
  puts "With index:"
  nums.each_with_index -> do |n, i|
    puts "  [#{i}] = #{n}"
  end

  # Integer iteration methods
  puts "3 times:"
  3.times -> { |i| puts "  #{i}" }

  puts "1 upto 3:"
  1.upto(3) -> { |i| puts "  #{i}" }

  # Map iteration
  ages = {"alice" => 30, "bob" => 25}
  ages.each -> { |k, v| puts "  #{k}: #{v}" }
  puts "Keys: #{ages.keys}, Values: #{ages.values}"

  # Chaining (functional pipelines)
  result = [1, 2, 3, 4, 5, 6]
    .select(-> { |n| n.even? })
    .map(-> { |n| n * 10 })
  puts "Chained: #{result}"
end
