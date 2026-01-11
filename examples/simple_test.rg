# Simple test file
def test_addition
  result = 1 + 1
  assert_equal(2, result)
end

def assert_equal(expected: Int, actual: Int)
  if expected != actual
    puts "FAIL: expected #{expected}, got #{actual}"
    exit(1)
  else
    puts "PASS"
  end
end

test_addition()
