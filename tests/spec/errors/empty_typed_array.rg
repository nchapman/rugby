#@ compile-fail
#
# BUG-036: Empty typed array declaration
# Empty arrays with type annotation on left-hand side

empty_nums : Array[Int] = []  #~ type mismatch: expected Array\[Int\], got Array\[any\]
