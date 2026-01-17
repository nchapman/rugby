#@ run-pass
#@ check-output
#
# Test: Section 9.2 - Constants

const MAX_SIZE = 1024
const DEFAULT_HOST = "localhost"
const PI = 3.14159

puts MAX_SIZE
puts DEFAULT_HOST
puts PI

# Typed constant
const TIMEOUT: Int64 = 30
puts TIMEOUT

# Computed constant
const BUFFER_SIZE = MAX_SIZE * 2
puts BUFFER_SIZE

#@ expect:
# 1024
# localhost
# 3.14159
# 30
# 2048
