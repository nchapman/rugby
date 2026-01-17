n = (ARGV[0] || 100).to_i

u = Array.new(n, 1.0)
v = Array.new(n, 1.0)

def eval_a(i, j)
  1.0 / ((i + j) * (i + j + 1) / 2 + i + 1)
end

def vector_times_array(vector, n)
  arr = []
  i = 0
  while i < n
    sum = 0.0
    j = 0
    while j < n
      sum += eval_a(i, j) * vector[j]
      j += 1
    end
    arr << sum
    i += 1
  end
  arr
end

def vector_times_array_transposed(vector, n)
  arr = []
  i = 0
  while i < n
    sum = 0.0
    j = 0
    while j < n
      sum += eval_a(j, i) * vector[j]
      j += 1
    end
    arr << sum
    i += 1
  end
  arr
end

def a_times_transp(vector, n)
  vector_times_array_transposed(vector_times_array(vector, n), n)
end

10.times do
  v = a_times_transp(u, n)
  u = a_times_transp(v, n)
end

vbv = 0.0
vv = 0.0
i = 0
while i < n
  vbv += u[i] * v[i]
  vv += v[i] * v[i]
  i += 1
end

puts "%0.9f" % Math.sqrt(vbv / vv)
