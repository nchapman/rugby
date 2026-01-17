#@ run-pass
#@ check-output
#
# Test: Section 10.5 - Variadic functions

import "fmt"

def log(level: String, *messages: String)
  for msg in messages
    puts "[#{level}] #{msg}"
  end
end

log("info", "Starting", "Loading config", "Ready")

# Splat to pass array as variadic
items = ["a", "b", "c"]
log("debug", *items)

# Variadic with Any type (uses Go interop for sprintf)
def format(template: String, *args: Any): String
  fmt.sprintf(template, *args)
end

result = format("Hello %s, you have %d messages", "Alice", 5)
puts result

#@ expect:
# [info] Starting
# [info] Loading config
# [info] Ready
# [debug] a
# [debug] b
# [debug] c
# Hello Alice, you have 5 messages

