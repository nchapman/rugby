# FizzBuzz - A classic programming exercise
# Demonstrates: ranges, conditionals, bare script mode
#
# This uses bare script mode - no main function needed.
# Top-level statements execute in source order.

for i in 1..15
  case
  when i % 15 == 0
    puts "FizzBuzz"
  when i % 3 == 0
    puts "Fizz"
  when i % 5 == 0
    puts "Buzz"
  else
    puts i
  end
end
