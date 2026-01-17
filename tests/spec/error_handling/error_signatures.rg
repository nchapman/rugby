#@ run-pass
#@ check-output
#
# Test: Section 16.1 - Error signatures

import "errors"

def read_file(path: String): (String, Error)
  if path == ""
    return "", errors.new("empty path")
  end
  "content of #{path}", nil
end

# Success case
content, err = read_file("test.txt")
if err == nil
  puts content
end

# Error case
content2, err2 = read_file("")
if err2 != nil
  puts "Error: #{err2}"
end

#@ expect:
# content of test.txt
# Error: empty path
