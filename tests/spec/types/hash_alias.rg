#@ run-pass
#@ check-output
#
# Test Hash as an alias for Map

def main
  # Hash should work like Map
  h: Hash<String, Int> = {}
  h["one"] = 1
  h["two"] = 2

  puts h["one"]
  puts h["two"]

  # Hash literal
  scores: Hash<String, Int> = {"alice" => 100, "bob" => 85}
  puts scores["alice"]
end

#@ expect:
# 1
# 2
# 100
