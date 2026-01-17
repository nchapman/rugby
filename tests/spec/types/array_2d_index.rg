#@ run-pass
#@ check-output
# 2D array indexing with variable indices

# Test Array<Array<Int>>
matrix = Array<Array<Int>>.new(3, Array<Int>.new(3, 0))
i = 1
j = 2
matrix[i][j] = 42
puts matrix[i][j]

# Test Array<Array<Float>> (mandelbrot case)
xloc = Array<Array<Float>>.new(2, Array<Float>.new(8, 0.0))
xloc[0][0] = 1.5
row = xloc[0]
puts row.length

# Variable index into nested array
idx = 0
inner = xloc[idx]
puts inner[0]

#@ expect:
# 42
# 8
# 1.5
