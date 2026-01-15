#@ run-pass
#@ check-output
#
# Test basic channel operations
# Note: Only buffered channels are tested here. Unbuffered channels require
# goroutines to avoid deadlock and are demonstrated in goroutines.rg

# Create a buffered channel
ch = Chan<Int>.new(3)

# Send values
ch << 1
ch << 2
ch << 3

# Receive values
puts ch.receive
puts ch.receive
puts ch.receive

# Test with strings
strCh = Chan<String>.new(2)
strCh << "hello"
strCh << "world"
puts strCh.receive
puts strCh.receive

#@ expect:
# 1
# 2
# 3
# hello
# world
