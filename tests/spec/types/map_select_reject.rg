#@ run-pass
#@ check-output

# Test map select and reject methods

scores = {"alice": 85, "bob": 92, "carol": 78, "dave": 88}

# Test select - filter entries where score > 80
puts "High scores:"
high_scores = scores.select -> { |name, score| score > 80 }
high_scores.keys.sorted.each -> { |name| puts "  #{name}: #{high_scores[name]}" }

# Test reject - filter out entries where score > 80
puts "Low scores:"
low_scores = scores.reject -> { |name, score| score > 80 }
low_scores.keys.sorted.each -> { |name| puts "  #{name}: #{low_scores[name]}" }

# Test select with name condition
puts "Names starting with 'a' or 'b':"
ab_scores = scores.select -> { |name, score| name.start_with?("a") || name.start_with?("b") }
ab_scores.keys.sorted.each -> { |name| puts "  #{name}" }

# Test merge
more_scores = {"eve": 95, "bob": 100}  # Bob's score updated
puts "Merged scores:"
merged = scores.merge(more_scores)
merged.keys.sorted.each -> { |name| puts "  #{name}: #{merged[name]}" }

#@ expect:
# High scores:
#   alice: 85
#   bob: 92
#   dave: 88
# Low scores:
#   carol: 78
# Names starting with 'a' or 'b':
#   alice
#   bob
# Merged scores:
#   alice: 85
#   bob: 100
#   carol: 78
#   dave: 88
#   eve: 95
