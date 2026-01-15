#@ run-pass
#@ check-output
#
# Test: Section 10.6 - Lambda vs for loop control flow

# Lambda return skips to next item (doesn't exit enclosing function)
def process_with_lambda(items : Array<Int>) -> Int
  sum = 0
  items.each -> (n) do
    return if n < 0  # skips this item
    sum += n
  end
  sum
end

puts process_with_lambda([1, -2, 3, -4, 5])

# For loop return exits enclosing function
def find_first_positive(items : Array<Int>) -> Int
  for n in items
    return n if n > 0
  end
  0
end

puts find_first_positive([-1, -2, 5, 10])

# For loop with break
def sum_until_negative(items : Array<Int>) -> Int
  sum = 0
  for n in items
    break if n < 0
    sum += n
  end
  sum
end

puts sum_until_negative([1, 2, 3, -1, 5])

#@ expect:
# 9
# 5
# 6
