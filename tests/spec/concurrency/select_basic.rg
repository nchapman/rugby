#@ run-pass
#@ check-output
#
# Test: Section 17.3 - Select

ch1 = Chan<Int>.new(1)
ch2 = Chan<String>.new(1)

ch1 << 42

select
when val = ch1.receive
  puts "got int: #{val}"
when msg = ch2.receive
  puts "got string: #{msg}"
else
  puts "nothing ready"
end

#@ expect:
# got int: 42
