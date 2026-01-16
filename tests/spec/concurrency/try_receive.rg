#@ run-pass
#@ check-output

# Test non-blocking channel receive with try_receive

ch = Chan<Int>.new(2)

# Try receive on empty channel - should return nil
if let val = ch.try_receive
  puts "unexpected: #{val}"
else
  puts "empty channel"
end

# Send some values
ch << 10
ch << 20

# Try receive should get the first value
if let val = ch.try_receive
  puts val
else
  puts "no value"
end

# Direct assignment also works
result = ch.try_receive
if let r = result
  puts r
end

# Channel should be empty now
if let val = ch.try_receive
  puts "unexpected"
else
  puts "channel drained"
end

#@ expect:
# empty channel
# 10
# 20
# channel drained
