#@ compile-fail
#
# Test compile-time template errors

import "rugby/template"

# Error: invalid template syntax
const BAD_SYNTAX = template.compile("{% if %}{% endif %}")  #~ ERROR: template syntax error

