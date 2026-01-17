import "os"
import "fmt"
import "math"
import "strconv"

def eval_a(i: Int, j: Int): Int
  (i + j) * (i + j + 1) / 2 + i + 1
end

def times(v: Array<Float>, u: Array<Float>, n: Int)
  i = 0
  while i < n
    a = 0.0
    j = 0
    while j < n
      a += u[j] / eval_a(i, j).to_f
      j += 1
    end
    v[i] = a
    i += 1
  end
end

def times_trans(v: Array<Float>, u: Array<Float>, n: Int)
  i = 0
  while i < n
    a = 0.0
    j = 0
    while j < n
      a += u[j] / eval_a(j, i).to_f
      j += 1
    end
    v[i] = a
    i += 1
  end
end

def a_times_transp(v: Array<Float>, u: Array<Float>, n: Int)
  x = Array.new(n, 0.0)
  times(x, u, n)
  times_trans(v, x, n)
end

def main
  n = 100
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    n = arg
  end

  u = Array.new(n, 1.0)
  v = Array.new(n, 1.0)

  10.times -> {
    a_times_transp(v, u, n)
    a_times_transp(u, v, n)
  }

  vbv = 0.0
  vv = 0.0
  i = 0
  while i < n
    vbv += u[i] * v[i]
    vv += v[i] * v[i]
    i += 1
  end

  fmt.Printf("%0.9f\n", math.Sqrt(vbv / vv))
end
