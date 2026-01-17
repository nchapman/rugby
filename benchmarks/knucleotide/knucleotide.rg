import "os"
import "fmt"
import "bufio"
import "strings"

# Read stdin until we find ">THREE", then return all following lines
def read_sequence(filename: String): Array<Int>
  f, err = os.Open(filename)
  if err != nil
    panic(err)
  end

  scanner = bufio.NewScanner(f)
  found_three = false
  data = Array<Int>.new(0, 0)

  while scanner.Scan
    line = scanner.Text
    if line.length == 0
      continue
    end

    if !found_three
      if line[0] == '>' && strings.HasPrefix(line, ">THREE")
        found_three = true
      end
      continue
    end

    # Convert to 2-bit encoding: A=0, C=1, T=2, G=3
    i = 0
    while i < line.length
      c = line[i]
      val = (c >> 1) & 3
      data.push(val)
      i += 1
    end
  end

  f.Close
  data
end

def count_frequencies(data: Array<Int>, size: Int): Hash<Int, Int>
  counts = {}

  i = 0
  while i <= data.length - size
    key = 0
    j = 0
    while j < size
      key = (key << 2) | data[i + j]
      j += 1
    end

    if counts[key] == nil
      counts[key] = 0
    end
    counts[key] += 1

    i += 1
  end

  counts
end

def decompress(num: Int, length: Int): String
  chars = ["A", "C", "T", "G"]
  result = ""
  i = 0
  while i < length
    result = chars[num & 3] + result
    num = num >> 2
    i += 1
  end
  result
end

def compress(sequence: String): Int
  to_num = { "A" => 0, "C" => 1, "T" => 2, "G" => 3 }
  num = 0
  i = 0
  while i < sequence.length
    num = (num << 2) | to_num[sequence[i]]
    i += 1
  end
  num
end

def write_frequencies(data: Array<Int>, size: Int): String
  counts = count_frequencies(data, size)
  total = (data.length - size + 1).to_f

  # Sort by count descending (would need sorting support)
  result = ""
  counts.each -> { |key, count|
    pct = 100.0 * count.to_f / total
    result = result + fmt.Sprintf("%s %.3f\n", decompress(key, size), pct)
  }
  result
end

def write_count(data: Array<Int>, nucleotide: String): String
  size = nucleotide.length
  counts = count_frequencies(data, size)
  key = compress(nucleotide)
  count = counts[key]
  if count == nil
    count = 0
  end
  fmt.Sprintf("%d\t%s", count, nucleotide)
end

def main
  filename = "25000_in"
  if os.Args.length > 1
    filename = os.Args[1]
  end

  data = read_sequence(filename)

  fmt.Println(write_frequencies(data, 1))
  fmt.Println(write_frequencies(data, 2))
  fmt.Println(write_count(data, "GGT"))
  fmt.Println(write_count(data, "GGTA"))
  fmt.Println(write_count(data, "GGTATT"))
  fmt.Println(write_count(data, "GGTATTTTAATT"))
  fmt.Println(write_count(data, "GGTATTTTAATTTATAGT"))
end
