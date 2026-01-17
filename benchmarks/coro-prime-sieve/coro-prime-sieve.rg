import "os"
import "fmt"
import "strconv"

# Send the sequence 2, 3, 4, ... to channel
def generate(ch: Chan<Int>)
  i = 2
  while true
    ch << i
    i += 1
  end
end

# Filter values from 'input' to 'output', removing those divisible by 'prime'
def filter(input: Chan<Int>, output: Chan<Int>, prime: Int)
  while true
    i = input.receive
    if i % prime != 0
      output << i
    end
  end
end

def main
  n = 1000
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    n = arg
  end

  ch = Chan<Int>.new(1)
  spawn generate(ch)

  i = 0
  while i < n
    prime = ch.receive
    fmt.Printf("%d\n", prime)

    if i < n - 1
      ch1 = Chan<Int>.new(1)
      spawn filter(ch, ch1, prime)
      ch = ch1
    end
    i += 1
  end
end
