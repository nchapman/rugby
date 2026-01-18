#@ compile-fail
#
# Test compile-time liquid template errors

import "rugby/liquid"

# Error: invalid template syntax
const BAD_SYNTAX = liquid.compile("{% if %}{% endif %}")  #~ ERROR: liquid template syntax error

