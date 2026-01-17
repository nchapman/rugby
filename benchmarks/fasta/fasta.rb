IM = 139968
IA = 3877
IC = 29573
LINE_WIDTH = 60

ALU = 'GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGGAGGCGGAGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCCAGCCTGGGCGACAGAGCGAGACTCCGTCTCAAAAA'

class Random
  def initialize(seed)
    @last = seed
  end

  def next
    @last = (@last * IA + IC) % IM
    @last.to_f / IM.to_f
  end
end

def repeat_fasta(seq, n)
  length = seq.length
  pos = 0

  while n > 0
    line_len = [n, LINE_WIDTH].min

    line_len.times do |i|
      print seq[(pos + i) % length]
    end
    puts

    pos = (pos + line_len) % length
    n -= line_len
  end
end

def random_fasta(genelist, n, rng)
  # Build cumulative probabilities
  cum = 0.0
  probs = genelist.map do |_, p|
    cum += p
    cum
  end

  while n > 0
    line_len = [n, LINE_WIDTH].min

    line_len.times do
      r = rng.next
      genelist.each_with_index do |(c, _), k|
        if probs[k] >= r
          print c
          break
        end
      end
    end
    puts

    n -= line_len
  end
end

n = ARGV.size > 0 ? ARGV[0].to_i : 1000

iub = [
  ['a', 0.27],
  ['c', 0.12],
  ['g', 0.12],
  ['t', 0.27],
  ['B', 0.02],
  ['D', 0.02],
  ['H', 0.02],
  ['K', 0.02],
  ['M', 0.02],
  ['N', 0.02],
  ['R', 0.02],
  ['S', 0.02],
  ['V', 0.02],
  ['W', 0.02],
  ['Y', 0.02]
]

homosapiens = [
  ['a', 0.3029549426680],
  ['c', 0.1979883004921],
  ['g', 0.1975473066391],
  ['t', 0.3015094502008]
]

rng = Random.new(42)

puts ">ONE Homo sapiens alu"
repeat_fasta(ALU, 2 * n)

puts ">TWO IUB ambiguity codes"
random_fasta(iub, 3 * n, rng)

puts ">THREE Homo sapiens frequency"
random_fasta(homosapiens, 5 * n, rng)
