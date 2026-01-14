#@ compile-fail
#
# BUG (discovered): If expression spanning multiple lines
# Multi-line if-as-expression fails to parse

x = 10
result = if x > 5
  "yes"
else  #~ unexpected token ELSE
  "no"
end
