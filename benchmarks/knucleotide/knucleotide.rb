def frequency(seq, length)
  counts = Hash.new(0)
  (0..seq.length - length).each do |i|
    counts[seq[i, length]] += 1
  end
  counts
end

def write_frequencies(seq, length)
  counts = frequency(seq, length)
  total = seq.length - length + 1

  sorted = counts.sort_by { |k, v| [-v, k] }
  sorted.each do |key, value|
    puts "%s %.3f" % [key, 100.0 * value / total]
  end
  puts
end

def write_count(seq, nucleotide)
  counts = frequency(seq, nucleotide.length)
  count = counts[nucleotide] || 0
  puts "%d\t%s" % [count, nucleotide]
end

# Read input
file_name = ARGV[0] || "25000_in"
seq = ""
three_reached = false

File.foreach(file_name) do |line|
  if three_reached
    seq << line.chomp.upcase
  elsif line.start_with?(">THREE")
    three_reached = true
  end
end

write_frequencies(seq, 1)
write_frequencies(seq, 2)
write_count(seq, "GGT")
write_count(seq, "GGTA")
write_count(seq, "GGTATT")
write_count(seq, "GGTATTTTAATT")
write_count(seq, "GGTATTTTAATTTATAGT")
