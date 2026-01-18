#@ run-pass
#@ check-output
#
# Test compile-time liquid templates with control flow

import "rugby/liquid"

# If/else
const IF_TMPL = liquid.compile("{% if show %}Visible{% else %}Hidden{% endif %}")

# Unless
const UNLESS_TMPL = liquid.compile("{% unless hidden %}Shown{% endunless %}")

# For loop
const FOR_TMPL = liquid.compile("{% for item in items %}{% unless forloop.first %} {% endunless %}{{ item }}{% endfor %}")

# For loop with forloop variable
const FORLOOP_TMPL = liquid.compile("{% for i in items %}{{ forloop.index }}{% endfor %}")

# For with else
const FOR_ELSE_TMPL = liquid.compile("{% for item in items %}{{ item }}{% else %}Empty{% endfor %}")

# Case/when
const CASE_TMPL = liquid.compile("{% case status %}{% when 1 %}One{% when 2 %}Two{% else %}Other{% endcase %}")

# Nested if
const NESTED_IF = liquid.compile("{% if a %}{% if b %}Both{% else %}A only{% endif %}{% else %}None{% endif %}")

puts IF_TMPL.MustRender({show: true})
puts IF_TMPL.MustRender({show: false})
puts UNLESS_TMPL.MustRender({hidden: false})
puts UNLESS_TMPL.MustRender({hidden: true})
puts FOR_TMPL.MustRender({items: [1, 2, 3]})
puts FORLOOP_TMPL.MustRender({items: ["a", "b", "c"]})
puts FOR_ELSE_TMPL.MustRender({items: []})
puts CASE_TMPL.MustRender({status: 1})
puts CASE_TMPL.MustRender({status: 2})
puts CASE_TMPL.MustRender({status: 99})
puts NESTED_IF.MustRender({a: true, b: true})
puts NESTED_IF.MustRender({a: true, b: false})
puts NESTED_IF.MustRender({a: false, b: true})

#@ expect:
# Visible
# Hidden
# Shown
#
# 1 2 3
# 123
# Empty
# One
# Two
# Other
# Both
# A only
# None
