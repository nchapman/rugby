#@ run-pass
#@ check-output
#
# Test: Section 16.9 - error_is? and error_as utilities
#
# These wrap Go's errors.Is and errors.As for checking error types.

import "errors"
import "io"

# Create test errors
def make_eof_error -> Error
  io.EOF
end

def make_custom_error -> Error
  errors.new("custom error")
end

# Test errors.is?
eof_err = make_eof_error
custom_err = make_custom_error

if errors.is?(eof_err, io.EOF)
  puts "is EOF"
end

if errors.is?(custom_err, io.EOF)
  puts "custom is EOF"
else
  puts "custom is not EOF"
end

#@ expect:
# is EOF
# custom is not EOF
