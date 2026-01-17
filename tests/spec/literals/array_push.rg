#@ run-pass
#@ check-output
#
# Test Array push method

def main
  arr = [1, 2, 3]
  arr.push(4)
  arr.push(5)

  arr.each do |n|
    puts n
  end
end

#@ expect:
# 1
# 2
# 3
# 4
# 5
