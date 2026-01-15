#@ run-pass
#@ check-output
#
# Test: Section 16.5 - Inline rescue

import "errors"

def get_port(s : String) -> (Int, Error)
  if s == ""
    return 0, errors.new("empty")
  end
  8080, nil
end

# Inline default
port1 = get_port("config") rescue 3000
puts port1

port2 = get_port("") rescue 3000
puts port2

#@ expect:
# 8080
# 3000
