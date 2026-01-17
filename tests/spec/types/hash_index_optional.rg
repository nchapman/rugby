#@ run-pass
#@ check-output
#
# Hash access with .get() method returns optional type
# For value-type hashes, use .get(key) to check for key existence
# Direct indexing hash[key] returns zero value for missing keys (Go semantics)

def main
  cache = {"a" => 1, "b" => 2, "c" => 3}

  # For value types, use .get() to get optional
  val = cache.get("a")
  if val == nil
    puts "a: not found"
  else
    puts "a: #{val ?? 0}"
  end

  # Test another existing key
  val2 = cache.get("b")
  if val2 != nil
    puts "b: #{val2 ?? 0}"
  end

  # Test missing key with .get()
  val3 = cache.get("z")
  if val3 == nil
    puts "z: not found"
  else
    puts "z: #{val3 ?? 0}"
  end

  # Test with nil coalescing - .get() returns optional
  puts "d: #{cache.get("d") ?? 99}"
  puts "c: #{cache.get("c") ?? 99}"

  # has_key? check is another option
  if cache.has_key?("a")
    puts "has a"
  end
  if !cache.has_key?("z")
    puts "no z"
  end
end

#@ expect:
# a: 1
# b: 2
# z: not found
# d: 99
# c: 3
# has a
# no z
