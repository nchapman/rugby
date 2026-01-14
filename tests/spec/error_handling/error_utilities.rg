#@ run-pass
#@ check-output
#
# Test error_is? and error_as utilities
#
# These wrap Go's errors.Is and errors.As for checking error types.

import errors
import io

# Create test errors
def make_eof_error -> error
  io.EOF
end

def make_custom_error -> error
  errors.New("custom error")
end

# Test error_is?
eof_err = make_eof_error()
custom_err = make_custom_error()

if error_is?(eof_err, io.EOF)
  puts "is EOF"
end

if error_is?(custom_err, io.EOF)
  puts "custom is EOF"
else
  puts "custom is not EOF"
end

#@ expect:
# is EOF
# custom is not EOF
