import "os"
import "fmt"
import "io/ioutil"
import "regexp"

def main
  filename = "25000_in"
  if os.Args.length > 1
    filename = os.Args[1]
  end

  f, err = os.Open(filename)
  if err != nil
    panic(err)
  end

  bytes, err = ioutil.ReadAll(f)
  if err != nil
    panic(err)
  end
  f.Close

  original_len = bytes.length

  # Clean the input - remove headers and newlines
  clean_re = regexp.MustCompile("(>[^\n]+)?\n")
  bytes = clean_re.ReplaceAllLiteral(bytes, "".bytes)
  cleaned_len = bytes.length

  # Variant patterns to count
  variants = [
    "agggtaaa|tttaccct",
    "[cgt]gggtaaa|tttaccc[acg]",
    "a[act]ggtaaa|tttacc[agt]t",
    "ag[act]gtaaa|tttac[agt]ct",
    "agg[act]taaa|ttta[agt]cct",
    "aggg[acg]aaa|ttt[cgt]ccct",
    "agggt[cgt]aa|tt[acg]accct",
    "agggta[cgt]a|t[acg]taccct",
    "agggtaa[cgt]|[acg]ttaccct"
  ]

  # Count each variant
  variants.each -> { |pattern|
    re = regexp.MustCompile(pattern)
    matches = re.FindAll(bytes, -1)
    count = 0
    if matches != nil
      count = matches.length
    end
    fmt.Printf("%s %d\n", pattern, count)
  }

  # Substitutions
  substitutions = [
    ("tHa[Nt]", "<4>"),
    ("aND|caN|Ha[DS]|WaS", "<3>"),
    ("a[NSt]|BY", "<2>"),
    ("<[^>]*>", "|"),
    ("\\|[^|][^|]*\\|", "-")
  ]

  substitutions.each -> { |pattern, replacement|
    re = regexp.MustCompile(pattern)
    bytes = re.ReplaceAll(bytes, replacement.bytes)
  }

  fmt.Printf("\n%d\n%d\n%d\n", original_len, cleaned_len, bytes.length)
end
