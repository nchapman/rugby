#@ run-pass
#@ check-output
#
# Test: Section 16.4 - Postfix bang operator (!)

import "errors"

def step1(ok : Bool) -> (String, Error)
  if !ok
    return "", errors.new("step1 failed")
  end
  "step1 done", nil
end

def step2(input : String) -> (String, Error)
  "#{input} -> step2 done", nil
end

def process(ok : Bool) -> (String, Error)
  # ! propagates errors
  s1 = step1(ok)!
  s2 = step2(s1)!
  s2, nil
end

# Success path
result, err = process(true)
if err == nil
  puts result
end

# Error path
result2, err2 = process(false)
if err2 != nil
  puts "Error: #{err2}"
end

#@ expect:
# step1 done -> step2 done
# Error: step1 failed
