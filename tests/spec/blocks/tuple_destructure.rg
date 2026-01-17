#@ run-pass
#@ check-output
# Block tuple destructuring for array of tuples

def main
  pairs = [("a", 1), ("b", 2), ("c", 3)]

  # Destructure tuple elements in block params
  pairs.each -> { |name, value|
    puts name
    puts value
  }
end

#@ expect:
# a
# 1
# b
# 2
# c
# 3
