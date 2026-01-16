#@ run-pass
#@ check-output
#
# Integration Test: Concurrency Patterns
# Features: Channels + Spawn/Await + Closures + Goroutines

import "time"

# Pattern 1: Producer-Consumer with channel
puts "Producer-Consumer:"
ch = Chan<Int>.new(3)

# Producer goroutine
go do
  for i in 1..3
    ch << i * 10
  end
  ch.close
end

# Give producer time to send (ensures deterministic output order)
time.sleep(50 * time.Millisecond)

# Consumer reads all values
for val in ch
  puts "  Received: #{val}"
end

# Pattern 2: Spawn and Await
puts "Spawn/Await:"
task1 = spawn { 10 + 20 }
task2 = spawn { 30 + 40 }

result1 = await task1
result2 = await task2
puts "  Task 1: #{result1}"
puts "  Task 2: #{result2}"

# Pattern 3: Closure capturing
puts "Closure capture:"
counter = 0
increment = -> { counter += 1 }

3.times -> { increment.() }
puts "  Counter: #{counter}"

# Pattern 4: Channel operations
puts "Channel operations:"
str_ch = Chan<String>.new(2)

str_ch << "first"
str_ch << "second"

msg1 = str_ch.receive
msg2 = str_ch.receive
puts "  Received: #{msg1}"
puts "  Received: #{msg2}"

puts "Done!"

#@ expect:
# Producer-Consumer:
#   Received: 10
#   Received: 20
#   Received: 30
# Spawn/Await:
#   Task 1: 30
#   Task 2: 70
# Closure capture:
#   Counter: 3
# Channel operations:
#   Received: first
#   Received: second
# Done!
