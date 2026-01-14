# Rugby Loops
# Demonstrates: for/in, while, until, break, next
#
# NOTE: Use loops when you need control flow (break, next, return).
# For data transformation, prefer blocks (see 09_blocks.rg).

def main
  # for/in with array
  puts "Array iteration:"
  fruits = ["apple", "banana", "cherry"]
  for fruit in fruits
    puts "  #{fruit}"
  end

  # for/in with inclusive range (1..3 includes 3)
  puts "Range 1..3:"
  for i in 1..3
    puts "  #{i}"
  end

  # for/in with exclusive range (0...3 excludes 3)
  puts "Range 0...3:"
  for i in 0...3
    puts "  #{i}"
  end

  # while loop
  puts "While loop:"
  n = 1
  while n <= 3
    puts "  #{n}"
    n += 1
  end

  # until loop (inverse of while)
  puts "Until loop:"
  m = 3
  until m == 0
    puts "  #{m}"
    m -= 1
  end

  # break to exit early
  puts "Break at 3:"
  for i in 1..10
    break if i > 3
    puts "  #{i}"
  end

  # next to skip iteration
  puts "Skip evens:"
  for i in 1..5
    next if i.even?
    puts "  #{i}"
  end

  # Compound assignment
  sum = 0
  for i in 1..5
    sum += i
  end
  puts "Sum 1..5: #{sum}"

  # Loop modifiers - execute repeatedly
  puts "Loop modifier (while):"
  items = [1, 2, 3]
  puts "  #{items.shift}" while items.any?

  puts "Loop modifier (until):"
  counter = 0
  counter += 1 until counter == 3
  puts "  Counter: #{counter}"
end
