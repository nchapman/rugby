#@ run-pass
#@ check-output
#
# Test spawn with direct function call (without block syntax)

def worker(ch: Chan<Int>, value: Int)
  ch << value * 2
end

def main
  ch = Chan<Int>.new(2)  # buffered channel

  # Direct call form (new syntax)
  spawn worker(ch, 5)
  spawn worker(ch, 10)

  # Read results
  a = ch.receive
  b = ch.receive
  puts a + b  # 10 + 20 = 30
end

#@ expect:
# 30
