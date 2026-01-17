import "os"
import "fmt"
import "strconv"

def fannkuch(n: Int): (Int, Int)
  perm = Array.new(n, 0)
  perm1 = Array.new(n, 0)
  count = Array.new(n, 0)

  # Initialize perm1 to [0, 1, 2, ..., n-1]
  i = 0
  while i < n
    perm1[i] = i
    i += 1
  end

  max_flips = 0
  checksum = 0
  r = n
  sign = true

  while true
    # Generate next permutation
    while r != 1
      count[r - 1] = r
      r -= 1
    end

    # Copy perm1 to perm and count flips
    i = 0
    while i < n
      perm[i] = perm1[i]
      i += 1
    end

    flips = 0
    while perm[0] != 0
      k = perm[0]
      # Reverse perm[0..k]
      i = 0
      j = k
      while i < j
        tmp = perm[i]
        perm[i] = perm[j]
        perm[j] = tmp
        i += 1
        j -= 1
      end
      flips += 1
    end

    if flips > max_flips
      max_flips = flips
    end

    if sign
      checksum += flips
    else
      checksum -= flips
    end

    # Next permutation
    while true
      if r == n
        return checksum, max_flips
      end

      p0 = perm1[0]
      i = 0
      while i < r
        perm1[i] = perm1[i + 1]
        i += 1
      end
      perm1[r] = p0

      count[r] -= 1
      if count[r] > 0
        break
      end
      r += 1
    end

    sign = !sign
  end

  return checksum, max_flips
end

def main
  n = 7
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    n = arg
  end

  checksum, max_flips = fannkuch(n)
  fmt.Printf("%d\nPfannkuchen(%d) = %d\n", checksum, n, max_flips)
end
