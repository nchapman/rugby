#@ run-pass
#@ check-output

import "rugby/liquid"

const GREETINGS = liquid.compile_glob("liquid/i18n/*.liquid")

def main
  puts GREETINGS["en.liquid"].MustRender({name: "Test"})
  puts GREETINGS["fr.liquid"].MustRender({name: "Test"})
end

#@ expect:
# Hello, Test!
# Bonjour, Test!
