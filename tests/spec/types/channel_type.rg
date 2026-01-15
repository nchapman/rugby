#@ run-pass
#@ check-output
#
# Test: Section 5.11 - Channel type

# Buffered channel
ch = Chan<Int>.new(2)

# Send values
ch << 1
ch << 2

# Receive values
puts ch.receive
puts ch.receive

puts "done"

#@ expect:
# 1
# 2
# done
