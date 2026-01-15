#@ compile-fail
#@ skip: Generics not yet implemented (Section 6.4)
#
# Test: Section 6.4 - Type constraints
# TODO: Implement type constraint syntax

# Constrain to types implementing an interface
def sort<T : Comparable>(arr : Array<T>) -> Array<T>
  # sorting logic
  arr
end

# Multiple constraints
def process<T : Readable & Closeable>(resource : T)
  data = resource.read
  resource.close
end
