#@ run-pass
#@ check-output
# Chained nil coalesce works with optional fallbacks

# All optionals
a : Int? = nil
b : Int? = 50
c : Int? = nil

# a is nil, b has value -> returns b (50)
result1 = a ?? b ?? c
puts result1 ?? -1

# a is nil, b is nil, c is nil -> result is nil
d : Int? = nil
e : Int? = nil
f : Int? = nil
result2 = d ?? e ?? f
puts result2 ?? -999

# First non-nil wins
g : Int? = 10
h : Int? = 20
i : Int? = 30
result3 = g ?? h ?? i
puts result3 ?? -1

# With final non-optional fallback
j : Int? = nil
k : Int? = nil
result4 = j ?? k ?? 100
puts result4

# String chaining
s1 : String? = nil
s2 : String? = "hello"
s3 : String? = "world"
puts s1 ?? s2 ?? s3 ?? "default"

#@ expect:
# 50
# -999
# 10
# 100
# hello
