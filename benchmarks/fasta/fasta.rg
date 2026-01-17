import "os"
import "bufio"
import "strconv"
import "strings"

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

def repeat_fasta(out: *bufio.Writer, seq: String, n: Int)
  len = seq.length
  pos = 0

  # Pre-extend sequence to avoid modulo in inner loop
  seq2 = seq + seq.substring(0, LINE_WIDTH)

  while n > 0
    line_len = n
    if line_len > LINE_WIDTH
      line_len = LINE_WIDTH
    end

    out.WriteString(seq2.substring(pos, pos + line_len))
    out.WriteString("\n")

    pos += line_len
    if pos >= len
      pos -= len
    end
    n -= line_len
  end
end

def random_fasta(out: *bufio.Writer, genelist: Array<(String, Float)>, n: Int, rng: Random)
  # Build cumulative probabilities
  cum = 0.0
  probs = Array.new(genelist.length, 0.0)
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

    line = strings.Builder.new
    j = 0
    while j < line_len
      r = rng.next
      k = 0
      while k < genelist.length
        if probs[k] >= r
          c, _ = genelist[k]
          line.WriteString(c)
          break
        end
        k += 1
      end
      j += 1
    end
    out.WriteString(line.String)
    out.WriteString("\n")

    n -= line_len
  end
end

def main
  n = 1000
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    n = arg
  end

  out = bufio.NewWriter(os.Stdout)

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

  out.WriteString(">ONE Homo sapiens alu\n")
  repeat_fasta(out, ALU, 2 * n)

  out.WriteString(">TWO IUB ambiguity codes\n")
  random_fasta(out, iub, 3 * n, rng)

  out.WriteString(">THREE Homo sapiens frequency\n")
  random_fasta(out, homosapiens, 5 * n, rng)

  out.Flush
end
