#@ run-pass
#@ check-output

import "rugby/template"

const TEMPLATES = template.compile_dir("template/i18n/")

def main
  puts TEMPLATES["en.tmpl"].MustRender({name: "World"})
  puts TEMPLATES["fr.tmpl"].MustRender({name: "World"})
end

#@ expect:
# Hello, World!
# Bonjour, World!
