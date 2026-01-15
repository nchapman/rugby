#@ run-pass
#@ check-output
#
# Test: Section 9.7 - While and until loops

# While loop
i = 0
while i < 3
  puts i
  i += 1
end

# Until loop (while NOT)
j = 3
until j == 0
  puts j
  j -= 1
end

# Loop with break
k = 0
loop do
  puts k
  k += 1
  break if k >= 2
end

#@ expect:
# 0
# 1
# 2
# 3
# 2
# 1
# 0
# 1
