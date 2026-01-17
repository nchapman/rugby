#@ run-pass
#@ check-output

# Test Array of tuples with literals and destructuring

def main
  # Array literal with tuple elements
  items: Array<(String, Float)> = [("a", 0.27), ("c", 0.12), ("g", 0.61)]

  # Destructure from array index
  _, p = items[0]
  puts p

  s, _ = items[1]
  puts s

  # Destructure both values
  letter, prob = items[2]
  puts letter
  puts prob

  # Loop with destructuring
  i = 0
  while i < items.length
    name, value = items[i]
    puts name
    i += 1
  end
end

#@ expect:
# 0.27
# c
# g
# 0.61
# a
# c
# g
