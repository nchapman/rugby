#@ run-pass
#@ check-output

# Test arrow lambda literals

def main
  # Simple lambda with no params
  greet = -> { "hello" }
  puts greet.()

  # Lambda with one param
  double = -> { |x : Int| x * 2 }
  puts double.(5)

  # Lambda with multiple params
  add = -> { |a : Int, b : Int| a + b }
  puts add.(3, 4)

  # Lambda with return type annotation
  triple = -> Int { |x : Int| x * 3 }
  puts triple.(10)

  # Multiline lambda with do...end
  calculate = -> do |x : Int|
    y = x * 2
    y + 1
  end
  puts calculate.(5)

  # Lambda capturing outer variable (closure)
  multiplier = 3
  scale = -> { |x : Int| x * multiplier }
  puts scale.(4)
end

#@ expect:
# hello
# 10
# 7
# 30
# 11
# 12
