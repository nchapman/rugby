#@ run-pass
#@ check-output

import "rugby/template"

const GREETINGS = template.compile_glob("template/i18n/*.tmpl")

def main
  puts GREETINGS["en.tmpl"].MustRender({name: "Test"})
  puts GREETINGS["fr.tmpl"].MustRender({name: "Test"})
end

#@ expect:
# Hello, Test!
# Bonjour, Test!
