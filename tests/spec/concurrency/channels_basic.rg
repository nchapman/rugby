#@ run-pass
#@ check-output
#
# Test: Section 17.2 - Channels

# Buffered channel
ch = Chan<Int>.new(3)

# Send values
ch << 1
ch << 2
ch << 3

# Receive values
puts ch.receive
puts ch.receive
puts ch.receive

# Close and iterate
ch2 = Chan<String>.new(2)
ch2 << "hello"
ch2 << "world"
ch2.close

for msg in ch2
  puts msg
end

puts "done"

#@ expect:
# 1
# 2
# 3
# hello
# world
# done
