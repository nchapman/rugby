#@ run-pass
#@ check-output
#
# Test: Section 4.5 - Heredoc interpolation

name = "World"

# Interpolation works by default
greeting = <<MSG
Hello #{name}!
MSG
puts greeting
puts "---"

# Single-quoted delimiter prevents interpolation
template = <<'MSG'
Hello #{name}!
MSG
puts template

#@ expect:
# Hello World!
# ---
# Hello #{name}!
