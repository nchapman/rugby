import "os"
import "fmt"
import "strconv"
import "crypto/md5"

def mbrot8(cr: Array<Float>, ci: Float): Int
  zr = Array.new(8, 0.0)
  zi = Array.new(8, 0.0)
  tr = Array.new(8, 0.0)
  ti = Array.new(8, 0.0)

  10.times -> {
    5.times -> {
      # zi = 2*zr*zi + ci
      i = 0
      while i < 8
        zi[i] = 2.0 * zr[i] * zi[i] + ci
        i += 1
      end

      # zr = tr - ti + cr
      i = 0
      while i < 8
        zr[i] = tr[i] - ti[i] + cr[i]
        i += 1
      end

      # tr = zr * zr
      i = 0
      while i < 8
        tr[i] = zr[i] * zr[i]
        i += 1
      end

      # ti = zi * zi
      i = 0
      while i < 8
        ti[i] = zi[i] * zi[i]
        i += 1
      end
    }

    # Check if all escaped
    terminate = true
    i = 0
    while i < 8
      if tr[i] + ti[i] <= 4.0
        terminate = false
        break
      end
      i += 1
    end
    if terminate
      return 0
    end
  }

  # Calculate result byte
  accu = 0
  i = 0
  while i < 8
    if tr[i] + ti[i] <= 4.0
      accu = accu | (0x80 >> i)
    end
    i += 1
  end
  accu
end

def main
  size = 200
  if os.Args.length > 1
    arg, _ = strconv.atoi(os.Args[1])
    size = arg
  end

  size = (size + 7) / 8 * 8
  chunk_size = size / 8
  inv = 2.0 / size.to_f

  # Build xloc array
  xloc = Array<Array<Float>>.new(chunk_size, Array<Float>.new(8, 0.0))
  i = 0
  while i < size
    xloc[i / 8][i % 8] = i.to_f * inv - 1.5
    i += 1
  end

  fmt.Printf("P4\n%d %d\n", size, size)

  pixels = Array.new(size * chunk_size, 0)
  chunk_id = 0
  while chunk_id < size
    ci = chunk_id.to_f * inv - 1.0
    offset = chunk_id * chunk_size
    i = 0
    while i < chunk_size
      r = mbrot8(xloc[i], ci)
      if r > 0
        pixels[offset + i] = r
      end
      i += 1
    end
    chunk_id += 1
  end

  hasher = md5.New()
  # hasher.Write(pixels)  # Need byte conversion
  fmt.Printf("%x\n", hasher.Sum(nil))
end
