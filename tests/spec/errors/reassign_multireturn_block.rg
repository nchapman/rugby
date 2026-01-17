#@ compile-fail
# Bug: Reassignment with multi-return inside block uses := instead of =
# When reassigning an existing variable with multi-return inside if/block,
# codegen incorrectly uses := (short declaration) instead of = (assignment)

import "strconv"

n = 4

# Works at top level: n, _ = strconv.atoi("10")
# Fails inside block:
if true
  n, _ = strconv.atoi("10")
end

puts n
#~ ERROR: no new variables on left side of :=
