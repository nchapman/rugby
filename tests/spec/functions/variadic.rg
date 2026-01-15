#@ compile-fail
#@ skip: Variadic functions not yet implemented (Section 10.5)
#
# Test: Section 10.5 - Variadic functions
# TODO: Implement variadic function syntax

def log(level : Symbol, *messages : String)
  for msg in messages
    puts "[#{level}] #{msg}"
  end
end

log(:info, "Starting", "Loading config", "Ready")

# Splat to pass array as variadic
items = ["a", "b", "c"]
log(:debug, *items)
