#@ run-pass
#@ check-output
#@ skip: Variadic functions not yet implemented (Section 10.5)
#
# Test: Section 10.5 - Variadic functions
# TODO: Implement variadic function syntax

import "fmt"

def log(level : Symbol, *messages : String)
  for msg in messages
    puts "[#{level}] #{msg}"
  end
end

log(:info, "Starting", "Loading config", "Ready")

# Splat to pass array as variadic
items = ["a", "b", "c"]
log(:debug, *items)

# Variadic with Any type (uses Go interop for sprintf)
def format(template : String, *args : Any) -> String
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
