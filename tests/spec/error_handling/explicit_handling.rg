#@ run-pass
#@ check-output
#
# Test: Section 16.7 - Explicit error handling

import "errors"

def read_config(path: String): (String, Error)
  if path == ""
    return "", errors.new("empty path")
  end
  if path == "missing"
    return "", errors.new("file not found")
  end
  "config data", nil
end

# Standard Go-style error handling
data, err = read_config("missing")
if err != nil
  puts "failed to read config: #{err}"
  data = "defaults"
end

puts data

#@ expect:
# failed to read config: file not found
# defaults
