#@ run-pass
#@ check-output
# Test comparing arrays and maps to nil

import "regexp"

# Array comparison to nil (like regex.FindAll returns)
def check_matches(matches: Array<Array<Int>>?)
  if matches != nil
    puts "found #{matches.length} matches"
  else
    puts "no matches"
  end
end

# Test with nil
check_matches(nil)

# Test with non-nil (simulate FindAll result)
re = regexp.MustCompile("a+")
text = "aaa bb aa".bytes
matches = re.FindAll(text, -1)
if matches != nil
  puts "got matches"
end

# Hash nil comparison
h: Hash<String, Int>? = nil
if h == nil
  puts "hash is nil"
end

#@ expect:
# no matches
# got matches
# hash is nil
