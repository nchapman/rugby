#@ run-pass
#@ check-output

# Test: Nested lambdas and higher-order functions (Section 10.6/10.7)

# Lambda returning lambda (currying)
def make_multiplier(factor: Int): (Int): Int
  -> { |x: Int| x * factor }
end

mult2 = make_multiplier(2)
mult3 = make_multiplier(3)

puts mult2.(5)
puts mult3.(5)

# Nested closures capturing outer scope
def make_transform(add: Int, mult: Int): (Int): Int
  -> { |x: Int| (x + add) * mult }
end

xform = make_transform(10, 2)
puts xform.(5)  # (5 + 10) * 2 = 30

# Higher-order: function accepting and returning lambdas
def compose(f: (Int): Int, g: (Int): Int): (Int): Int
  -> { |x: Int| f.(g.(x)) }
end

inc = -> { |x: Int| x + 1 }
dbl = -> { |x: Int| x * 2 }

composed1 = compose(dbl, inc)
puts composed1.(4)  # dbl(inc(4)) = dbl(5) = 10

composed2 = compose(inc, dbl)
puts composed2.(4)  # inc(dbl(4)) = inc(8) = 9

# Closure capturing and mutating state
def make_counter: (): Int
  count = 0
  -> {
    count += 1
    count
  }
end

counter = make_counter()
puts counter.()
puts counter.()
puts counter.()

#@ expect:
# 10
# 15
# 30
# 10
# 9
# 1
# 2
# 3
