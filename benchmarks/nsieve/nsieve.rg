import "os"
import "fmt"
import "strconv"

def nsieve(n: Int)
  flags = Array.new(n, false)
  count = 0

  i = 2
  while i < n
    if !flags[i]
      count += 1
      j = i * 2
      while j < n
        flags[j] = true
        j += i
      end
    end
    i += 1
  end

  fmt.Printf("Primes up to %8d %8d\n", n, count)
end

def main
  n = 4
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    n = arg
  end

  3.times -> { |i|
    nsieve(10000 << (n - i))
  }
end
