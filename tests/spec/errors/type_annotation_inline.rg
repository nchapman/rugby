#@ compile-fail
#
# BUG-036: Inline type annotation syntax
# Empty typed array literals with inline annotation syntax fail

# This should work but doesn't
empty_strs = [] : Array[String]  #~ undefined: 'Array'
