MATCHERS = [
  /agggtaaa|tttaccct/,
  /[cgt]gggtaaa|tttaccc[acg]/,
  /a[act]ggtaaa|tttacc[agt]t/,
  /ag[act]gtaaa|tttac[agt]ct/,
  /agg[act]taaa|ttta[agt]cct/,
  /aggg[acg]aaa|ttt[cgt]ccct/,
  /agggt[cgt]aa|tt[acg]accct/,
  /agggta[cgt]a|t[acg]taccct/,
  /agggtaa[cgt]|[acg]ttaccct/
]

SUBSTITUTIONS = [
  [/tHa[Nt]/, '<4>'],
  [/aND|caN|Ha[DS]|WaS/, '<3>'],
  [/a[NSt]|BY/, '<2>'],
  [/<[^>]*>/, '|'],
  [/\|[^|][^|]*\|/, '-']
]

file_name = ARGV[0] || "25000_in"
seq = File.read(file_name)
original_len = seq.size

# Clean the input
seq.gsub!(/>.*\n|\n/, '')
cleaned_len = seq.size

# Count variants
MATCHERS.each do |matcher|
  count = 0
  seq.scan(matcher) { count += 1 }
  puts "#{matcher.source} #{count}"
end

# Apply substitutions
SUBSTITUTIONS.each do |pattern, replacement|
  seq.gsub!(pattern, replacement)
end

puts
puts original_len
puts cleaned_len
puts seq.size
