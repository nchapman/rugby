import "os"
import "fmt"
import "strconv"
import "math/big"

# Global state for the spigot algorithm
class PiState
  property tmp1: *big.Int
  property tmp2: *big.Int
  property y2: *big.Int
  property bigk: *big.Int
  property accum: *big.Int
  property denom: *big.Int
  property numer: *big.Int
  property ten: *big.Int
  property three: *big.Int
  property four: *big.Int

  def initialize
    @tmp1 = big.NewInt(0)
    @tmp2 = big.NewInt(0)
    @y2 = big.NewInt(1)
    @bigk = big.NewInt(0)
    @accum = big.NewInt(0)
    @denom = big.NewInt(1)
    @numer = big.NewInt(1)
    @ten = big.NewInt(10)
    @three = big.NewInt(3)
    @four = big.NewInt(4)
  end

  def next_term(k: Int64): Int64
    while true
      k += 1
      @y2.SetInt64(k * 2 + 1)
      @bigk.SetInt64(k)

      @tmp1.Lsh(@numer, 1)
      @accum.Add(@accum, @tmp1)
      @accum.Mul(@accum, @y2)
      @denom.Mul(@denom, @y2)
      @numer.Mul(@numer, @bigk)

      if @accum.Cmp(@numer) > 0
        return k
      end
    end
    k
  end

  def extract_digit(nth: *big.Int): Int64
    @tmp1.Mul(nth, @numer)
    @tmp2.Add(@tmp1, @accum)
    @tmp1.Div(@tmp2, @denom)
    @tmp1.Int64()
  end

  def next_digit(k: Int64): (Int64, Int64)
    while true
      k = self.next_term(k)
      d3 = self.extract_digit(@three)
      d4 = self.extract_digit(@four)
      if d3 == d4
        return d3, k
      end
    end
    return 0, k
  end

  def eliminate_digit(d: Int64)
    @tmp1.SetInt64(d)
    @accum.Sub(@accum, @tmp1.Mul(@denom, @tmp1))
    @accum.Mul(@accum, @ten)
    @numer.Mul(@numer, @ten)
  end
end

def main
  n = 27
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    n = arg
  end

  state = PiState.new
  line = ""
  k: Int64 = 0
  d: Int64 = 0

  i = 1
  while i <= n
    d, k = state.next_digit(k)
    line = line + fmt.Sprintf("%d", d)

    if line.length == 10
      fmt.Printf("%s\t:%d\n", line, i)
      line = ""
    end

    state.eliminate_digit(d)
    i += 1
  end

  if line.length > 0
    fmt.Printf("%-10s\t:%d\n", line, n)
  end
end
