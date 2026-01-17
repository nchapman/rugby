import "os"
import "fmt"
import "strconv"

const IM = 139968
const IA = 3877
const IC = 29573
const LINE_WIDTH = 60

const ALU = "GGCCGGGCGCGGTGGCTCACGCCTGTAATCCCAGCACTTTGGGAGGCCGAGGCGGGCGGATCACCTGAGGTCAGGAGTTCGAGACCAGCCTGGCCAACATGGTGAAACCCCGTCTCTACTAAAAATACAAAAATTAGCCGGGCGTGGTGGCGCGCGCCTGTAATCCCAGCTACTCGGGAGGCTGAGGCAGGAGAATCGCTTGAACCCGGGAGGCGGAGGTTGCAGTGAGCCGAGATCGCGCCACTGCACTCCAGCCTGGGCGACAGAGCGAGACTCCGTCTCAAAAA"

class Random
  property last: Int

  def initialize(@last: Int)
  end

  def next: Float
    @last = (@last * IA + IC) % IM
    @last.to_f / IM.to_f
  end
end

def repeat_fasta(seq: String, n: Int)
  len = seq.length
  pos = 0

  while n > 0
    line_len = n
    if line_len > LINE_WIDTH
      line_len = LINE_WIDTH
    end

    i = 0
    while i < line_len
      fmt.Printf("%s", seq[(pos + i) % len])
      i += 1
    end
    fmt.Printf("\n")

    pos = (pos + line_len) % len
    n -= line_len
  end
end

def random_fasta(genelist: Array<(String, Float)>, n: Int, rng: Random)
  # Build cumulative probabilities
  cum = 0.0
  probs = Array<Float>.new(genelist.length, 0.0)
  i = 0
  while i < genelist.length
    _, p = genelist[i]
    cum += p
    probs[i] = cum
    i += 1
  end

  while n > 0
    line_len = n
    if line_len > LINE_WIDTH
      line_len = LINE_WIDTH
    end

    j = 0
    while j < line_len
      r = rng.next
      k = 0
      while k < genelist.length
        if probs[k] >= r
          c, _ = genelist[k]
          fmt.Printf("%s", c)
          break
        end
        k += 1
      end
      j += 1
    end
    fmt.Printf("\n")

    n -= line_len
  end
end

def main
  n = 1000
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    n = arg
  end

  iub = [
    ("a", 0.27),
    ("c", 0.12),
    ("g", 0.12),
    ("t", 0.27),
    ("B", 0.02),
    ("D", 0.02),
    ("H", 0.02),
    ("K", 0.02),
    ("M", 0.02),
    ("N", 0.02),
    ("R", 0.02),
    ("S", 0.02),
    ("V", 0.02),
    ("W", 0.02),
    ("Y", 0.02)
  ]

  homosapiens = [
    ("a", 0.3029549426680),
    ("c", 0.1979883004921),
    ("g", 0.1975473066391),
    ("t", 0.3015094502008)
  ]

  rng = Random.new(42)

  fmt.Printf(">ONE Homo sapiens alu\n")
  repeat_fasta(ALU, 2 * n)

  fmt.Printf(">TWO IUB ambiguity codes\n")
  random_fasta(iub, 3 * n, rng)

  fmt.Printf(">THREE Homo sapiens frequency\n")
  random_fasta(homosapiens, 5 * n, rng)
end
