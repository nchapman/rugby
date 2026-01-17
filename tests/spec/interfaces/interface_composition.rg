#@ run-pass
#@ check-output
#
# Test: Section 13.1 - Interface composition

interface Reader
  def read: String
end

interface Writer
  def write(data: String)
end

interface ReadWriter < Reader, Writer
end

class Buffer implements ReadWriter
  def initialize
    @data = ""
  end

  def read: String
    @data
  end

  def write(data: String)
    @data = @data + data
  end
end

def copy(from: Reader, to: Writer)
  to.write(from.read)
end

buf1 = Buffer.new
buf1.write("hello")

buf2 = Buffer.new
copy(buf1, buf2)

puts buf2.read

#@ expect:
# hello
