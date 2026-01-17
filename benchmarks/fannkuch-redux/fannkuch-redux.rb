def fannkuch(n)
  perm = Array.new(n)
  perm1 = (0...n).to_a
  count = Array.new(n)

  max_flips = 0
  checksum = 0
  r = n
  sign = true

  loop do
    # Generate next permutation
    while r != 1
      count[r - 1] = r
      r -= 1
    end

    # Copy and count flips
    perm = perm1.dup
    flips = 0
    while perm[0] != 0
      k = perm[0]
      # Reverse perm[0..k]
      i = 0
      j = k
      while i < j
        perm[i], perm[j] = perm[j], perm[i]
        i += 1
        j -= 1
      end
      flips += 1
    end

    max_flips = flips if flips > max_flips
    if sign
      checksum += flips
    else
      checksum -= flips
    end

    # Next permutation
    loop do
      return [checksum, max_flips] if r == n
      p0 = perm1[0]
      i = 0
      while i < r
        perm1[i] = perm1[i + 1]
        i += 1
      end
      perm1[r] = p0

      count[r] -= 1
      break if count[r] > 0
      r += 1
    end
    sign = !sign
  end
end

n = (ARGV[0] || 7).to_i
checksum, max_flips = fannkuch(n)
puts checksum
puts "Pfannkuchen(#{n}) = #{max_flips}"
