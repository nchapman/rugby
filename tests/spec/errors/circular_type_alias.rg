#@ compile-fail
#
# Test circular type alias detection

type A = B
type B = A  #~ ERROR: circular type alias

def main
  x: A = 1
end
