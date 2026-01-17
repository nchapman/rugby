#@ compile-fail
# Bug: Top-level assignment (not const) with def main not allowed
# Regular assignments at top level cannot coexist with def main
# Note: 'const X = ...' works, but 'X = ...' does not

SOLAR_MASS = 39.47841760435743

def main
  puts SOLAR_MASS
end
#~ ERROR: cannot mix top-level statements with 'def main'
