#@ run-pass
#@ check-output

import "rugby/liquid"

const TEMPLATES = liquid.compile_dir("liquid/i18n/")

def main
  puts TEMPLATES["en.liquid"].MustRender({name: "World"})
  puts TEMPLATES["fr.liquid"].MustRender({name: "World"})
end

#@ expect:
# Hello, World!
# Bonjour, World!
