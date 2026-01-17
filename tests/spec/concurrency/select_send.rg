#@ run-pass
#@ check-output

# Test: select with send case (Section 17.3)
# select can include send operations, not just receive

# Test select with send to buffered channel
ch1 = Chan<Int>.new(1)

# ch1 has space (buffer size 1)
select
when ch1 << 42
  puts "sent to ch1"
else
  puts "could not send"
end

# Verify value was sent
puts ch1.receive

# Test select with receive
ch3 = Chan<Int>.new(1)
ch3 << 100  # ch3 has a value ready to receive

select
when val = ch3.receive
  puts "received #{val} from ch3"
else
  puts "nothing ready"
end

# Test select with full channel (send would block)
full = Chan<Int>.new(1)
full << 1  # fill the buffer

select
when full << 2
  puts "sent to full channel"
else
  puts "channel full, skipped"
end

#@ expect:
# sent to ch1
# 42
# received 100 from ch3
# channel full, skipped
