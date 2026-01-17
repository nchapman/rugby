import "os"
import "fmt"
import "bufio"
import "strings"

# Read file until we find ">THREE", then return all following lines
def read_sequence(filename: String): String
  f, err = os.Open(filename)
  if err != nil
    panic(err)
  end

  scanner = bufio.NewScanner(f)
  data = strings.Builder.new

  # Skip until >THREE
  while scanner.Scan
    line = scanner.Text
    if strings.HasPrefix(line, ">THREE")
      break
    end
  end

  # Read the sequence
  while scanner.Scan
    line = scanner.Text
    if line.length > 0 && line[0] != '>'
      data.WriteString(strings.ToUpper(line))
    end
  end

  f.Close
  data.String
end

def frequency(seq: String, length: Int): Hash<String, Int>
  counts: Hash<String, Int> = {}
  max = seq.length - length

  i = 0
  while i <= max
    # Use efficient substring function instead of range slicing
    key = seq.substring(i, i + length)
    current = counts.fetch(key, 0)
    counts[key] = current + 1
    i += 1
  end

  counts
end

def write_frequencies(seq: String, length: Int): String
  counts = frequency(seq, length)
  total = (seq.length - length + 1).to_f

  result = strings.Builder.new
  counts.each -> { |key, count|
    pct = 100.0 * count.to_f / total
    result.WriteString(fmt.Sprintf("%s %.3f\n", key, pct))
  }
  result.WriteString("\n")
  result.String
end

def write_count(seq: String, nucleotide: String): String
  counts = frequency(seq, nucleotide.length)
  count = counts.fetch(nucleotide, 0)
  fmt.Sprintf("%d\t%s", count, nucleotide)
end

def main
  filename = "25000_in"
  if os.Args.length > 1
    filename = os.Args[1]
  end

  seq = read_sequence(filename)

  fmt.Print(write_frequencies(seq, 1))
  fmt.Print(write_frequencies(seq, 2))
  fmt.Println(write_count(seq, "GGT"))
  fmt.Println(write_count(seq, "GGTA"))
  fmt.Println(write_count(seq, "GGTATT"))
  fmt.Println(write_count(seq, "GGTATTTTAATT"))
  fmt.Println(write_count(seq, "GGTATTTTAATTTATAGT"))
end
