#@ run-pass
#@ check-output
#
# Test: Section 10.4 - Named parameters

def fetch(url: String, timeout: Int = 30, retry: Bool = true)
  puts url
end

fetch("https://example.com")
fetch("https://example.com", timeout: 60)
fetch("https://example.com", retry: false, timeout: 120)

# Named params can be passed in any order
fetch("https://example.com", retry: true, timeout: 100)

#@ expect:
# https://example.com
# https://example.com
# https://example.com
# https://example.com
