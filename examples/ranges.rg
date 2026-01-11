def main
  # Range literal stored in variable
  r = 1..5
  puts("Range stored")
  puts(r)

  # Inclusive range in for loop
  puts("Inclusive range (0..3):")
  for i in 0..3
    puts(i)
  end

  # Exclusive range in for loop
  puts("Exclusive range (0...3):")
  for i in 0...3
    puts(i)
  end

  # Range methods
  nums = (1..5).to_a
  puts("Range to array:")
  for n in nums
    puts(n)
  end

  puts("Range size:")
  size = (1..10).size
  puts(size)

  puts("Range contains 5?")
  contains5 = (1..10).include?(5)
  puts(contains5)

  puts("Range contains 15?")
  contains15 = (1..10).include?(15)
  puts(contains15)

  # Range each with block
  puts("Range each:")
  (1..3).each do |i|
    puts(i)
  end
end
